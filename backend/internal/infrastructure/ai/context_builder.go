package ai

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/cognitree/backend/internal/domain/entity"
	"github.com/cognitree/backend/internal/domain/repository"
	domainservice "github.com/cognitree/backend/internal/domain/service"
	"github.com/cognitree/backend/pkg/logger"
)

type contextBuilder struct {
	treeRepo        repository.TreeRepository
	nodeRepo        repository.NodeRepository
	qaPairRepo      repository.QAPairRepository
	blockRepo       repository.BlockRepository
	anchorRepo      repository.AnchorRepository
	summaryProvider summaryProvider
}

type contextSection struct {
	priority int
	title    string
	content  string
}

func NewContextBuilder(
	treeRepo repository.TreeRepository,
	nodeRepo repository.NodeRepository,
	qaPairRepo repository.QAPairRepository,
	blockRepo repository.BlockRepository,
	anchorRepo repository.AnchorRepository,
	summaryRepo repository.SummaryRepository,
) domainservice.ContextBuilder {
	return &contextBuilder{
		treeRepo:        treeRepo,
		nodeRepo:        nodeRepo,
		qaPairRepo:      qaPairRepo,
		blockRepo:       blockRepo,
		anchorRepo:      anchorRepo,
		summaryProvider: newRepositorySummaryProvider(summaryRepo),
	}
}

func (b *contextBuilder) BuildContext(ctx context.Context, treeID string, currentNodeID string, newQuestion string) (*domainservice.ContextPayload, error) {
	nodes, err := b.nodeRepo.GetByTreeID(ctx, treeID)
	if err != nil {
		return nil, fmt.Errorf("get tree nodes: %w", err)
	}

	nodeMap := make(map[string]*entity.Node)
	childrenMap := make(map[string][]*entity.Node)
	var rootNode *entity.Node

	for _, n := range nodes {
		nodeMap[n.ID] = n
		if n.ParentNodeID == nil {
			rootNode = n
			continue
		}
		childrenMap[*n.ParentNodeID] = append(childrenMap[*n.ParentNodeID], n)
	}

	if rootNode == nil {
		return nil, fmt.Errorf("root node not found")
	}

	thread, err := b.buildThread(nodeMap, currentNodeID)
	if err != nil {
		return nil, fmt.Errorf("build thread: %w", err)
	}

	currentNode := thread[len(thread)-1]
	warnings := make([]string, 0)

	treeGoal, treeWarnings := b.collectTreeGoal(ctx, treeID, rootNode)
	warnings = append(warnings, treeWarnings...)

	treeOverview := b.collectTreeOverview(rootNode, childrenMap, currentNodeID)

	threadDetail, threadWarnings := b.collectThreadDetail(ctx, thread)
	warnings = append(warnings, threadWarnings...)

	anchorEvidence, anchorWarnings := b.collectAnchorEvidence(ctx, currentNode, nodeMap)
	warnings = append(warnings, anchorWarnings...)

	nodeSummary, pathSummary, subtreeSummary, summaryWarnings := b.collectSummarySections(ctx, currentNodeID)
	warnings = append(warnings, summaryWarnings...)

	sections := b.selectSections([]contextSection{
		{priority: 1, title: "Tree Goal", content: treeGoal},
		{priority: 2, title: "Anchor Evidence", content: anchorEvidence},
		{priority: 3, title: "Current Path", content: threadDetail},
		{priority: 4, title: "Node Summary", content: nodeSummary},
		{priority: 5, title: "Path Summary", content: pathSummary},
		{priority: 6, title: "Subtree Summary", content: subtreeSummary},
		{priority: 7, title: "Tree Overview", content: treeOverview},
		{priority: 99, title: "Current Ask", content: strings.TrimSpace(newQuestion)},
	})

	return &domainservice.ContextPayload{
		SystemPrompt: b.formatSystemPrompt(),
		UserPrompt:   b.formatUserPrompt(sections),
		Degraded:     len(warnings) > 0,
		Warnings:     warnings,
	}, nil
}

func (b *contextBuilder) collectTreeGoal(ctx context.Context, treeID string, rootNode *entity.Node) (string, []string) {
	tree, err := b.treeRepo.GetByID(ctx, treeID)
	if err != nil {
		return fmt.Sprintf("- Root Question: %s", rootNode.Question), []string{
			fmt.Sprintf("tree goal degraded: %v", err),
		}
	}

	lines := make([]string, 0, 3)
	if title := strings.TrimSpace(tree.Title); title != "" {
		lines = append(lines, fmt.Sprintf("- Title: %s", title))
	}
	if description := strings.TrimSpace(tree.Description); description != "" {
		lines = append(lines, fmt.Sprintf("- Description: %s", description))
	}
	lines = append(lines, fmt.Sprintf("- Root Question: %s", rootNode.Question))

	return strings.Join(lines, "\n"), nil
}

func (b *contextBuilder) collectTreeOverview(rootNode *entity.Node, childrenMap map[string][]*entity.Node, currentNodeID string) string {
	var sb strings.Builder
	b.renderTreeOverview(&sb, rootNode, childrenMap, 0, currentNodeID)
	return strings.TrimSpace(sb.String())
}

func (b *contextBuilder) collectThreadDetail(ctx context.Context, thread []*entity.Node) (string, []string) {
	var sb strings.Builder
	warnings := make([]string, 0)

	for _, n := range thread {
		sb.WriteString(fmt.Sprintf("### Node: %s\n", n.Question))

		qaPairs, err := b.qaPairRepo.GetByNodeID(ctx, n.ID)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("thread degraded: node %s qa_pairs unavailable: %v", n.ID, err))
			sb.WriteString("- [Q&A omitted due to context degradation]\n\n")
			continue
		}

		if len(qaPairs) == 0 {
			sb.WriteString("- No historical Q&A yet.\n\n")
			continue
		}

		for _, qp := range qaPairs {
			sb.WriteString(fmt.Sprintf("**Q:** %s\n", qp.Question))

			blocks, err := b.blockRepo.GetByQAPairID(ctx, qp.ID)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("thread degraded: qa_pair %s blocks unavailable: %v", qp.ID, err))
				sb.WriteString("**A:** [answer omitted due to context degradation]\n\n")
				continue
			}

			if len(blocks) == 0 {
				sb.WriteString("**A:** [no answer blocks]\n\n")
				continue
			}

			for _, block := range blocks {
				sb.WriteString(fmt.Sprintf("**A:** %s\n", block.Content))
			}
			sb.WriteString("\n")
		}
	}

	return strings.TrimSpace(sb.String()), warnings
}

func (b *contextBuilder) collectAnchorEvidence(ctx context.Context, currentNode *entity.Node, nodeMap map[string]*entity.Node) (string, []string) {
	if currentNode.AnchorID == nil {
		return "", nil
	}

	anchor, err := b.anchorRepo.GetByID(ctx, *currentNode.AnchorID)
	if err != nil {
		return fmt.Sprintf("- Anchor ID: %s\n- Status: unavailable due to context degradation", *currentNode.AnchorID), []string{
			fmt.Sprintf("anchor evidence degraded: anchor %s unavailable: %v", *currentNode.AnchorID, err),
		}
	}

	lines := make([]string, 0, 3)
	if sourceNode, ok := nodeMap[anchor.SourceNodeID]; ok {
		lines = append(lines, fmt.Sprintf("- Source Node: %s", sourceNode.Question))
	} else {
		lines = append(lines, fmt.Sprintf("- Source Node ID: %s", anchor.SourceNodeID))
	}
	lines = append(lines, fmt.Sprintf("- Quoted Text: %s", anchor.QuotedText))
	lines = append(lines, fmt.Sprintf("- Offsets: %d-%d", anchor.StartOffset, anchor.EndOffset))

	return strings.Join(lines, "\n"), nil
}

func (b *contextBuilder) collectSummarySections(ctx context.Context, currentNodeID string) (string, string, string, []string) {
	warnings := make([]string, 0)

	nodeSummary, found, err := b.summaryProvider.GetNodeSummary(ctx, currentNodeID)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("summary degraded: node summary unavailable: %v", err))
	} else if !found {
		logMissingSummary("node", currentNodeID)
	}

	pathSummary, found, err := b.summaryProvider.GetPathSummary(ctx, currentNodeID)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("summary degraded: path summary unavailable: %v", err))
	} else if !found {
		logMissingSummary("path", currentNodeID)
	}

	subtreeSummary, found, err := b.summaryProvider.GetSubtreeSummary(ctx, currentNodeID)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("summary degraded: subtree summary unavailable: %v", err))
	} else if !found {
		logMissingSummary("subtree", currentNodeID)
	}

	return strings.TrimSpace(nodeSummary), strings.TrimSpace(pathSummary), strings.TrimSpace(subtreeSummary), warnings
}

func logMissingSummary(scope string, nodeID string) {
	if logger.L != nil {
		logger.L.Infow("summary missing",
			"scope", scope,
			"node_id", nodeID,
		)
	}
}

func (b *contextBuilder) selectSections(sections []contextSection) []contextSection {
	selected := make([]contextSection, 0, len(sections))
	for _, section := range sections {
		if strings.TrimSpace(section.content) == "" {
			continue
		}
		selected = append(selected, section)
	}

	sort.Slice(selected, func(i, j int) bool {
		return selected[i].priority < selected[j].priority
	})

	return selected
}

func (b *contextBuilder) formatSystemPrompt() string {
	return strings.TrimSpace(`
你是一个知识探索助手，正在帮助用户在一棵“思维树”上逐步深化问题。

回答时请遵循以下原则：
1. 先理解整棵树当前要解决的目标，再回答当前问题。
2. 如果提供了 anchor evidence，要把它视为当前问题的直接语义证据。
3. 优先围绕当前节点与当前路径给出清晰、结构化的回答。
4. 使用 Markdown 输出，并适当提示用户下一步可以继续深入的方向。`)
}

func (b *contextBuilder) formatUserPrompt(sections []contextSection) string {
	var sb strings.Builder

	for idx, section := range sections {
		if idx > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString("## ")
		sb.WriteString(section.title)
		sb.WriteString("\n\n")
		sb.WriteString(strings.TrimSpace(section.content))
	}

	return sb.String()
}

func (b *contextBuilder) buildThread(nodeMap map[string]*entity.Node, currentNodeID string) ([]*entity.Node, error) {
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

func (b *contextBuilder) renderTreeOverview(sb *strings.Builder, node *entity.Node, childrenMap map[string][]*entity.Node, depth int, currentID string) {
	indent := strings.Repeat("  ", depth)
	marker := ""
	if node.ID == currentID {
		marker = " <- current"
	}

	sb.WriteString(fmt.Sprintf("%s- %s [%s]%s\n", indent, node.Question, node.Status, marker))

	for _, child := range childrenMap[node.ID] {
		b.renderTreeOverview(sb, child, childrenMap, depth+1, currentID)
	}
}
