package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cognitree/backend/internal/domain/entity"
	"github.com/cognitree/backend/internal/domain/repository"
	"github.com/cognitree/backend/pkg/logger"
	"github.com/google/uuid"
)

type SummaryDispatcher interface {
	EnqueueForNode(treeID string, nodeID string)
}

type summaryJob struct {
	treeID       string
	targetNodeID string
	scope        entity.SummaryScope
}

type treeSummaryContext struct {
	Title        string
	Description  string
	RootQuestion string
}

type SummaryService struct {
	treeRepo    repository.TreeRepository
	nodeRepo    repository.NodeRepository
	qaPairRepo  repository.QAPairRepository
	blockRepo   repository.BlockRepository
	summaryRepo repository.SummaryRepository
	aiClient    AIClient

	queue                chan summaryJob
	startOnce            sync.Once
	maxImmediateAttempts int
	maxTotalAttempts     int
	compensateEvery      time.Duration
	workerBackoffs       []time.Duration
}

func NewSummaryService(
	treeRepo repository.TreeRepository,
	nodeRepo repository.NodeRepository,
	qaPairRepo repository.QAPairRepository,
	blockRepo repository.BlockRepository,
	summaryRepo repository.SummaryRepository,
	aiClient AIClient,
) *SummaryService {
	return &SummaryService{
		treeRepo:             treeRepo,
		nodeRepo:             nodeRepo,
		qaPairRepo:           qaPairRepo,
		blockRepo:            blockRepo,
		summaryRepo:          summaryRepo,
		aiClient:             aiClient,
		queue:                make(chan summaryJob, 128),
		maxImmediateAttempts: 3,
		maxTotalAttempts:     6,
		compensateEvery:      30 * time.Second,
		workerBackoffs:       []time.Duration{200 * time.Millisecond, 500 * time.Millisecond, 1 * time.Second},
	}
}

func (s *SummaryService) Start(ctx context.Context) {
	s.startOnce.Do(func() {
		go s.workerLoop(ctx)
		go s.compensatorLoop(ctx)
	})
}

func (s *SummaryService) EnqueueForNode(treeID string, nodeID string) {
	jobs := []summaryJob{
		{treeID: treeID, targetNodeID: nodeID, scope: entity.SummaryScopeNode},
		{treeID: treeID, targetNodeID: nodeID, scope: entity.SummaryScopePath},
		{treeID: treeID, targetNodeID: nodeID, scope: entity.SummaryScopeSubtree},
	}

	for _, job := range jobs {
		select {
		case s.queue <- job:
		default:
			if logger.L != nil {
				logger.L.Warnw("summary queue full, falling back to direct async processing",
					"tree_id", treeID,
					"node_id", nodeID,
					"scope", job.scope,
				)
			}
			go s.processJob(context.Background(), job)
		}
	}
}

func (s *SummaryService) workerLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-s.queue:
			s.processJob(ctx, job)
		}
	}
}

func (s *SummaryService) compensatorLoop(ctx context.Context) {
	ticker := time.NewTicker(s.compensateEvery)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.compensateFailedSummaries(ctx)
		}
	}
}

func (s *SummaryService) compensateFailedSummaries(ctx context.Context) {
	candidates, err := s.loadCompensationCandidates(ctx)
	if err != nil {
		if logger.L != nil {
			logger.L.Warnw("summary compensation scan failed", "error", err)
		}
		return
	}

	for _, summary := range candidates {
		if summary.AttemptCount >= s.maxTotalAttempts {
			continue
		}

		if logger.L != nil {
			logger.L.Infow("re-enqueue failed summary for compensation",
				"summary_id", summary.ID,
				"tree_id", summary.TreeID,
				"target_node_id", summary.TargetNodeID,
				"scope", summary.Scope,
				"attempt_count", summary.AttemptCount,
			)
		}

		summary.Status = entity.SummaryStatusPending
		if err := s.summaryRepo.Update(ctx, summary); err != nil {
			s.logFailure(summaryJob{treeID: summary.TreeID, targetNodeID: summary.TargetNodeID, scope: summary.Scope}, "mark pending for compensation", err)
			continue
		}

		select {
		case s.queue <- summaryJob{treeID: summary.TreeID, targetNodeID: summary.TargetNodeID, scope: summary.Scope}:
		default:
			go s.processJob(context.Background(), summaryJob{treeID: summary.TreeID, targetNodeID: summary.TargetNodeID, scope: summary.Scope})
		}
	}
}

func (s *SummaryService) loadCompensationCandidates(ctx context.Context) ([]*entity.Summary, error) {
	failed, err := s.summaryRepo.ListByStatus(ctx, entity.SummaryStatusFailed)
	if err != nil {
		return nil, err
	}

	stalePending, err := s.summaryRepo.ListStalePendingBefore(ctx, time.Now().Add(-s.compensateEvery))
	if err != nil {
		return nil, err
	}

	return append(failed, stalePending...), nil
}

func (s *SummaryService) processJob(ctx context.Context, job summaryJob) {
	summary, err := s.ensureSummary(ctx, job)
	if err != nil {
		s.logFailure(job, "ensure summary", err)
		return
	}

	maxAttempt := summary.AttemptCount + s.maxImmediateAttempts
	if maxAttempt > s.maxTotalAttempts {
		maxAttempt = s.maxTotalAttempts
	}

	for attempt := summary.AttemptCount + 1; attempt <= maxAttempt; attempt++ {
		summary.AttemptCount = attempt
		summary.Status = entity.SummaryStatusPending
		summary.ErrorMessage = ""
		if err := s.summaryRepo.Update(ctx, summary); err != nil {
			s.logFailure(job, "update pending summary", err)
			return
		}

		content, genErr := s.generateSummary(ctx, job)
		if genErr == nil {
			summary.Content = content
			summary.Status = entity.SummaryStatusReady
			summary.ErrorMessage = ""
			if err := s.summaryRepo.Update(ctx, summary); err != nil {
				s.logFailure(job, "persist ready summary", err)
				return
			}
			if logger.L != nil {
				logger.L.Infow("summary generated",
					"summary_id", summary.ID,
					"tree_id", summary.TreeID,
					"target_node_id", summary.TargetNodeID,
					"scope", summary.Scope,
					"attempt_count", summary.AttemptCount,
				)
			}
			return
		}

		summary.ErrorMessage = genErr.Error()
		if attempt >= maxAttempt {
			summary.Status = entity.SummaryStatusFailed
			if err := s.summaryRepo.Update(ctx, summary); err != nil {
				s.logFailure(job, "persist failed summary", err)
			}
			s.logFailure(job, "summary generation exhausted retries", genErr)
			return
		}

		if err := s.summaryRepo.Update(ctx, summary); err != nil {
			s.logFailure(job, "persist retry state", err)
			return
		}

		if attempt-1 < len(s.workerBackoffs) {
			time.Sleep(s.workerBackoffs[attempt-1])
		} else {
			time.Sleep(s.workerBackoffs[len(s.workerBackoffs)-1])
		}
	}
}

func (s *SummaryService) ensureSummary(ctx context.Context, job summaryJob) (*entity.Summary, error) {
	summary, err := s.summaryRepo.GetLatestByScopeAndTarget(ctx, job.scope, job.targetNodeID)
	if err == nil {
		return summary, nil
	}

	summary = &entity.Summary{
		ID:           uuid.New().String(),
		TreeID:       job.treeID,
		TargetNodeID: job.targetNodeID,
		Scope:        job.scope,
		Version:      1,
		AttemptCount: 0,
		Status:       entity.SummaryStatusPending,
		Content:      "",
		ErrorMessage: "",
	}

	if createErr := s.summaryRepo.Create(ctx, summary); createErr != nil {
		if errors.Is(createErr, repository.ErrSummaryAlreadyExists) {
			existing, getErr := s.summaryRepo.GetLatestByScopeAndTarget(ctx, job.scope, job.targetNodeID)
			if getErr != nil {
				return nil, fmt.Errorf("refetch summary after conflict: %w", getErr)
			}
			return existing, nil
		}
		return nil, fmt.Errorf("create summary: %w", createErr)
	}

	return summary, nil
}

func (s *SummaryService) generateSummary(ctx context.Context, job summaryJob) (string, error) {
	switch job.scope {
	case entity.SummaryScopeNode:
		return s.generateNodeSummary(ctx, job.treeID, job.targetNodeID)
	case entity.SummaryScopePath:
		return s.generatePathSummary(ctx, job.treeID, job.targetNodeID)
	case entity.SummaryScopeSubtree:
		return s.generateSubtreeSummary(ctx, job.treeID, job.targetNodeID)
	default:
		return "", fmt.Errorf("unknown summary scope: %s", job.scope)
	}
}

func (s *SummaryService) generateNodeSummary(ctx context.Context, treeID, nodeID string) (string, error) {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return "", fmt.Errorf("load node: %w", err)
	}

	treeContext, err := s.loadTreeSummaryContext(ctx, treeID)
	if err != nil {
		return "", fmt.Errorf("load tree context: %w", err)
	}

	qaPairs, err := s.qaPairRepo.GetByNodeID(ctx, nodeID)
	if err != nil {
		return "", fmt.Errorf("load qa pairs: %w", err)
	}

	userPrompt := buildNodeSummaryPrompt(ctx, treeID, treeContext, node, qaPairs, s.blockRepo)
	return s.aiClient.Chat(ctx, summarySystemPrompt("node"), userPrompt)
}

func (s *SummaryService) generatePathSummary(ctx context.Context, treeID, nodeID string) (string, error) {
	nodes, err := s.nodeRepo.GetByTreeID(ctx, treeID)
	if err != nil {
		return "", fmt.Errorf("load tree nodes: %w", err)
	}

	treeContext, err := s.loadTreeSummaryContext(ctx, treeID)
	if err != nil {
		return "", fmt.Errorf("load tree context: %w", err)
	}

	thread, err := buildThread(nodes, nodeID)
	if err != nil {
		return "", fmt.Errorf("build thread: %w", err)
	}

	userPrompt := buildPathSummaryPrompt(ctx, treeID, treeContext, thread, s.qaPairRepo, s.blockRepo)
	return s.aiClient.Chat(ctx, summarySystemPrompt("path"), userPrompt)
}

func (s *SummaryService) generateSubtreeSummary(ctx context.Context, treeID, nodeID string) (string, error) {
	nodes, err := s.nodeRepo.GetByTreeID(ctx, treeID)
	if err != nil {
		return "", fmt.Errorf("load tree nodes: %w", err)
	}

	treeContext, err := s.loadTreeSummaryContext(ctx, treeID)
	if err != nil {
		return "", fmt.Errorf("load tree context: %w", err)
	}

	userPrompt := buildSubtreeSummaryPrompt(ctx, treeID, treeContext, nodeID, nodes, s.qaPairRepo, s.blockRepo)
	return s.aiClient.Chat(ctx, summarySystemPrompt("subtree"), userPrompt)
}

func (s *SummaryService) loadTreeSummaryContext(ctx context.Context, treeID string) (*treeSummaryContext, error) {
	tree, err := s.treeRepo.GetByID(ctx, treeID)
	if err != nil {
		return nil, err
	}

	rootNodes, err := s.nodeRepo.GetByTreeID(ctx, treeID)
	if err != nil {
		return nil, err
	}

	var rootQuestion string
	for _, node := range rootNodes {
		if node.ParentNodeID == nil {
			rootQuestion = node.Question
			break
		}
	}

	return &treeSummaryContext{
		Title:        strings.TrimSpace(tree.Title),
		Description:  strings.TrimSpace(tree.Description),
		RootQuestion: strings.TrimSpace(rootQuestion),
	}, nil
}

func (s *SummaryService) logFailure(job summaryJob, stage string, err error) {
	if logger.L != nil {
		logger.L.Warnw("summary generation failed",
			"tree_id", job.treeID,
			"target_node_id", job.targetNodeID,
			"scope", job.scope,
			"stage", stage,
			"error", err,
		)
	}
}

func summarySystemPrompt(scope string) string {
	return strings.TrimSpace(fmt.Sprintf(`
你正在为 Cognitree 生成 %s summary。

要求：
1. 只输出适合复用的摘要文本。
2. 保持简洁、结构化、可作为后续上下文的一部分。
3. 优先提炼结论、线索和关键边界，不要重复原始长文。
4. 输出 Markdown。`, scope))
}

func buildNodeSummaryPrompt(ctx context.Context, treeID string, treeContext *treeSummaryContext, node *entity.Node, qaPairs []*entity.QAPair, blockRepo repository.BlockRepository) string {
	var sb strings.Builder
	writeTreeSummaryContext(&sb, treeID, treeContext)
	sb.WriteString("\nNode Focus:\n")
	sb.WriteString("- Question: ")
	sb.WriteString(node.Question)
	sb.WriteString("\n- Scope: summarize this node as a reusable knowledge capsule.\n")
	sb.WriteString("\nHistorical Q&A:\n")

	if len(qaPairs) == 0 {
		sb.WriteString("- No historical Q&A yet.\n")
		return sb.String()
	}

	for _, qp := range qaPairs {
		sb.WriteString("- Q: ")
		sb.WriteString(qp.Question)
		sb.WriteString("\n")

		blocks, err := blockRepo.GetByQAPairID(ctx, qp.ID)
		if err != nil || len(blocks) == 0 {
			continue
		}
		for _, block := range blocks {
			sb.WriteString("  - A: ")
			sb.WriteString(block.Content)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func buildPathSummaryPrompt(ctx context.Context, treeID string, treeContext *treeSummaryContext, thread []*entity.Node, qaPairRepo repository.QAPairRepository, blockRepo repository.BlockRepository) string {
	var sb strings.Builder
	writeTreeSummaryContext(&sb, treeID, treeContext)
	sb.WriteString("\nPath Focus:\n")
	sb.WriteString("- Scope: summarize the exploration path as a reusable narrative.\n")
	sb.WriteString("\nPath:\n")

	for _, node := range thread {
		sb.WriteString("- Node: ")
		sb.WriteString(node.Question)
		sb.WriteString(" [")
		sb.WriteString(string(node.Status))
		sb.WriteString("]\n")

		qaPairs, err := qaPairRepo.GetByNodeID(ctx, node.ID)
		if err != nil || len(qaPairs) == 0 {
			continue
		}
		for _, qp := range qaPairs {
			sb.WriteString("  - Q: ")
			sb.WriteString(qp.Question)
			sb.WriteString("\n")

			blocks, err := blockRepo.GetByQAPairID(ctx, qp.ID)
			if err != nil || len(blocks) == 0 {
				continue
			}
			for _, block := range blocks {
				sb.WriteString("    - A: ")
				sb.WriteString(block.Content)
				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}

func buildSubtreeSummaryPrompt(ctx context.Context, treeID string, treeContext *treeSummaryContext, nodeID string, nodes []*entity.Node, qaPairRepo repository.QAPairRepository, blockRepo repository.BlockRepository) string {
	nodeMap := make(map[string]*entity.Node)
	childrenMap := make(map[string][]*entity.Node)
	for _, n := range nodes {
		nodeMap[n.ID] = n
		if n.ParentNodeID != nil {
			childrenMap[*n.ParentNodeID] = append(childrenMap[*n.ParentNodeID], n)
		}
	}

	var sb strings.Builder
	writeTreeSummaryContext(&sb, treeID, treeContext)
	sb.WriteString("\nSubtree Focus:\n")
	sb.WriteString("- Scope: summarize this branch as reusable knowledge.\n")
	sb.WriteString("\nSubtree Root:\n")

	root, ok := nodeMap[nodeID]
	if !ok {
		sb.WriteString("- missing subtree root\n")
		return sb.String()
	}

	appendSubtreePrompt(ctx, &sb, root, childrenMap, qaPairRepo, blockRepo, 0)
	return sb.String()
}

func writeTreeSummaryContext(sb *strings.Builder, treeID string, treeContext *treeSummaryContext) {
	sb.WriteString("Tree ID: ")
	sb.WriteString(treeID)
	sb.WriteString("\nTree Context:\n")
	if treeContext != nil {
		if treeContext.Title != "" {
			sb.WriteString("- Title: ")
			sb.WriteString(treeContext.Title)
			sb.WriteString("\n")
		}
		if treeContext.Description != "" {
			sb.WriteString("- Description: ")
			sb.WriteString(treeContext.Description)
			sb.WriteString("\n")
		}
		if treeContext.RootQuestion != "" {
			sb.WriteString("- Root Question: ")
			sb.WriteString(treeContext.RootQuestion)
			sb.WriteString("\n")
		}
	}
}

func appendSubtreePrompt(ctx context.Context, sb *strings.Builder, node *entity.Node, childrenMap map[string][]*entity.Node, qaPairRepo repository.QAPairRepository, blockRepo repository.BlockRepository, depth int) {
	indent := strings.Repeat("  ", depth)
	sb.WriteString(indent)
	sb.WriteString("- Node: ")
	sb.WriteString(node.Question)
	sb.WriteString(" [")
	sb.WriteString(string(node.Status))
	sb.WriteString("]\n")

	qaPairs, err := qaPairRepo.GetByNodeID(ctx, node.ID)
	if err == nil {
		for _, qp := range qaPairs {
			sb.WriteString(indent)
			sb.WriteString("  - Q: ")
			sb.WriteString(qp.Question)
			sb.WriteString("\n")

			blocks, err := blockRepo.GetByQAPairID(ctx, qp.ID)
			if err != nil {
				continue
			}
			for _, block := range blocks {
				sb.WriteString(indent)
				sb.WriteString("    - A: ")
				sb.WriteString(block.Content)
				sb.WriteString("\n")
			}
		}
	}

	for _, child := range childrenMap[node.ID] {
		appendSubtreePrompt(ctx, sb, child, childrenMap, qaPairRepo, blockRepo, depth+1)
	}
}

func buildThread(nodes []*entity.Node, currentNodeID string) ([]*entity.Node, error) {
	nodeMap := make(map[string]*entity.Node)
	for _, node := range nodes {
		nodeMap[node.ID] = node
	}

	var thread []*entity.Node
	id := currentNodeID
	for {
		node, ok := nodeMap[id]
		if !ok {
			if len(thread) == 0 {
				return nil, fmt.Errorf("current node %s not found in tree", currentNodeID)
			}
			return nil, fmt.Errorf("broken parent chain at node %s", id)
		}

		thread = append([]*entity.Node{node}, thread...)
		if node.ParentNodeID == nil {
			return thread, nil
		}
		id = *node.ParentNodeID
	}
}
