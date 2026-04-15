package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/cognitree/backend/internal/application/dto"
	"github.com/cognitree/backend/internal/domain/entity"
	"github.com/cognitree/backend/internal/domain/repository"
	domainservice "github.com/cognitree/backend/internal/domain/service"
)

type StreamAIClient interface {
	ChatStream(ctx context.Context, systemPrompt, userPrompt string, onDelta func(string) error) error
}

type TreeStreamService struct {
	treeRepo          repository.TreeRepository
	nodeRepo          repository.NodeRepository
	qaPairRepo        repository.QAPairRepository
	blockRepo         repository.BlockRepository
	contextBuilder    domainservice.ContextBuilder
	aiClient          StreamAIClient
	summaryDispatcher SummaryDispatcher
}

func NewTreeStreamService(
	treeRepo repository.TreeRepository,
	nodeRepo repository.NodeRepository,
	qaPairRepo repository.QAPairRepository,
	blockRepo repository.BlockRepository,
	contextBuilder domainservice.ContextBuilder,
	aiClient StreamAIClient,
	summaryDispatcher SummaryDispatcher,
) *TreeStreamService {
	return &TreeStreamService{
		treeRepo:          treeRepo,
		nodeRepo:          nodeRepo,
		qaPairRepo:        qaPairRepo,
		blockRepo:         blockRepo,
		contextBuilder:    contextBuilder,
		aiClient:          aiClient,
		summaryDispatcher: summaryDispatcher,
	}
}

func (s *TreeStreamService) StreamFirstQuestion(
	ctx context.Context,
	req dto.CreateTreeStreamRequest,
	emit func(dto.TreeStreamEvent) error,
) error {
	question := strings.TrimSpace(req.Question)
	if question == "" {
		return fmt.Errorf("question is required")
	}

	tree, rootNode, err := s.resolveTreeAndRoot(ctx, req, question)
	if err != nil {
		return fmt.Errorf("resolve tree and root: %w", err)
	}

	treeResp := toTreeResponse(tree)
	if err := emit(dto.TreeStreamEvent{Type: "tree_ready", Tree: &treeResp}); err != nil {
		return err
	}
	rootNodeResp := toNodeResponse(rootNode)
	if err := emit(dto.TreeStreamEvent{Type: "root_node_ready", RootNode: &rootNodeResp}); err != nil {
		return err
	}

	payload, err := s.contextBuilder.BuildContext(ctx, tree.ID, rootNode.ID, question)
	if err != nil {
		_ = emit(dto.TreeStreamEvent{Type: "error", Message: err.Error()})
		return fmt.Errorf("build context: %w", err)
	}

	var answer strings.Builder
	if err := s.aiClient.ChatStream(ctx, payload.SystemPrompt, payload.UserPrompt, func(delta string) error {
		answer.WriteString(delta)
		return emit(dto.TreeStreamEvent{Type: "answer_delta", Delta: delta})
	}); err != nil {
		_ = emit(dto.TreeStreamEvent{Type: "error", Message: err.Error()})
		return fmt.Errorf("ai chat stream: %w", err)
	}

	qaPair := &entity.QAPair{
		NodeID:   rootNode.ID,
		Question: question,
	}
	if err := s.qaPairRepo.Create(ctx, qaPair); err != nil {
		_ = emit(dto.TreeStreamEvent{Type: "error", Message: err.Error()})
		return fmt.Errorf("create qa_pair: %w", err)
	}

	block := &entity.Block{
		QAPairID: qaPair.ID,
		Type:     entity.BlockTypeParagraph,
		Content:  answer.String(),
	}
	if err := s.blockRepo.CreateBatch(ctx, []*entity.Block{block}); err != nil {
		_ = emit(dto.TreeStreamEvent{Type: "error", Message: err.Error()})
		return fmt.Errorf("create block: %w", err)
	}

	if rootNode.Status == entity.NodeStatusDraft {
		rootNode.Status = entity.NodeStatusAnswered
		if err := s.nodeRepo.Update(ctx, rootNode); err != nil {
			_ = emit(dto.TreeStreamEvent{Type: "error", Message: err.Error()})
			return fmt.Errorf("update root node status: %w", err)
		}
	}

	if s.summaryDispatcher != nil {
		s.summaryDispatcher.EnqueueForNode(tree.ID, rootNode.ID)
	}

	if err := emit(dto.TreeStreamEvent{Type: "completed"}); err != nil {
		return err
	}

	return nil
}

func (s *TreeStreamService) resolveTreeAndRoot(
	ctx context.Context,
	req dto.CreateTreeStreamRequest,
	question string,
) (*entity.Tree, *entity.Node, error) {
	if strings.TrimSpace(req.TreeID) != "" {
		tree, err := s.treeRepo.GetByID(ctx, strings.TrimSpace(req.TreeID))
		if err != nil {
			return nil, nil, fmt.Errorf("get tree: %w", err)
		}

		rootNode, err := s.resolveRootNode(ctx, tree.ID, strings.TrimSpace(req.RootNodeID))
		if err != nil {
			return nil, nil, err
		}

		return tree, rootNode, nil
	}

	tree := &entity.Tree{
		Title: deriveTreeTitle(question),
	}
	if err := s.treeRepo.Create(ctx, tree); err != nil {
		return nil, nil, fmt.Errorf("create tree: %w", err)
	}

	rootNode := &entity.Node{
		TreeID:   tree.ID,
		Question: question,
		Status:   entity.NodeStatusDraft,
	}
	if err := s.nodeRepo.Create(ctx, rootNode); err != nil {
		return nil, nil, fmt.Errorf("create root node: %w", err)
	}

	return tree, rootNode, nil
}

func (s *TreeStreamService) resolveRootNode(ctx context.Context, treeID string, rootNodeID string) (*entity.Node, error) {
	if rootNodeID != "" {
		rootNode, err := s.nodeRepo.GetByID(ctx, rootNodeID)
		if err != nil {
			return nil, fmt.Errorf("get root node: %w", err)
		}
		if rootNode.TreeID != treeID {
			return nil, fmt.Errorf("root node does not belong to tree")
		}
		return rootNode, nil
	}

	nodes, err := s.nodeRepo.GetByTreeID(ctx, treeID)
	if err != nil {
		return nil, fmt.Errorf("get tree nodes: %w", err)
	}

	for _, node := range nodes {
		if node.ParentNodeID == nil {
			return node, nil
		}
	}

	return nil, fmt.Errorf("root node not found")
}
