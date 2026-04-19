package service

import (
	"context"
	"fmt"

	"github.com/cognitree/backend/internal/application/dto"
	"github.com/cognitree/backend/internal/domain/entity"
	"github.com/cognitree/backend/internal/domain/repository"
	domainservice "github.com/cognitree/backend/internal/domain/service"
	"github.com/cognitree/backend/pkg/logger"
)

type ChatClient interface {
	Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error)
	ChatStream(ctx context.Context, systemPrompt, userPrompt string, onDelta func(string) error) error
}

type AIClient = ChatClient

type ChatService struct {
	nodeRepo          repository.NodeRepository
	qaPairRepo        repository.QAPairRepository
	blockRepo         repository.BlockRepository
	contextBuilder    domainservice.ContextBuilder
	aiClient          ChatClient
	summaryDispatcher SummaryDispatcher
}

func NewChatService(
	nodeRepo repository.NodeRepository,
	qaPairRepo repository.QAPairRepository,
	blockRepo repository.BlockRepository,
	contextBuilder domainservice.ContextBuilder,
	aiClient ChatClient,
	summaryDispatcher SummaryDispatcher,
) *ChatService {
	return &ChatService{
		nodeRepo:          nodeRepo,
		qaPairRepo:        qaPairRepo,
		blockRepo:         blockRepo,
		contextBuilder:    contextBuilder,
		aiClient:          aiClient,
		summaryDispatcher: summaryDispatcher,
	}
}

func (s *ChatService) Chat(ctx context.Context, nodeID string, req dto.ChatRequest) (*dto.ChatResponse, error) {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("get node: %w", err)
	}

	payload, err := s.contextBuilder.BuildContext(ctx, node.TreeID, nodeID, req.Question)
	if err != nil {
		return nil, fmt.Errorf("build context: %w", err)
	}
	if payload.Degraded && logger.L != nil {
		logger.L.Warnw("context built with degradation",
			"node_id", nodeID,
			"tree_id", node.TreeID,
			"warnings", payload.Warnings,
		)
	}

	answer, err := s.aiClient.Chat(ctx, payload.SystemPrompt, payload.UserPrompt)
	if err != nil {
		return nil, fmt.Errorf("ai chat: %w", err)
	}

	return s.persistChatAnswer(ctx, node, req.Question, answer)
}

func (s *ChatService) StreamChat(
	ctx context.Context,
	nodeID string,
	req dto.ChatRequest,
	emit func(dto.ChatStreamEvent) error,
) error {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("get node: %w", err)
	}

	payload, err := s.contextBuilder.BuildContext(ctx, node.TreeID, nodeID, req.Question)
	if err != nil {
		return fmt.Errorf("build context: %w", err)
	}
	if payload.Degraded && logger.L != nil {
		logger.L.Warnw("context built with degradation",
			"node_id", nodeID,
			"tree_id", node.TreeID,
			"warnings", payload.Warnings,
		)
	}

	var answer string
	if err := s.aiClient.ChatStream(ctx, payload.SystemPrompt, payload.UserPrompt, func(delta string) error {
		answer += delta
		return emit(dto.ChatStreamEvent{Type: "answer_delta", Delta: delta})
	}); err != nil {
		_ = emit(dto.ChatStreamEvent{Type: "error", Message: err.Error()})
		return fmt.Errorf("ai chat stream: %w", err)
	}

	if _, err := s.persistChatAnswer(ctx, node, req.Question, answer); err != nil {
		_ = emit(dto.ChatStreamEvent{Type: "error", Message: err.Error()})
		return err
	}

	qaPairResponse, err := s.buildLatestQAPairResponse(ctx, node.ID)
	if err != nil {
		_ = emit(dto.ChatStreamEvent{Type: "error", Message: err.Error()})
		return err
	}
	if err := emit(dto.ChatStreamEvent{Type: "qa_pair_ready", QAPair: qaPairResponse}); err != nil {
		return err
	}

	if err := emit(dto.ChatStreamEvent{Type: "completed"}); err != nil {
		return err
	}

	return nil
}

func (s *ChatService) persistChatAnswer(ctx context.Context, node *entity.Node, question, answer string) (*dto.ChatResponse, error) {
	qaPair := &entity.QAPair{
		NodeID:   node.ID,
		Question: question,
	}
	if err := s.qaPairRepo.Create(ctx, qaPair); err != nil {
		return nil, fmt.Errorf("create qa_pair: %w", err)
	}

	block := &entity.Block{
		QAPairID: qaPair.ID,
		Type:     entity.BlockTypeParagraph,
		Content:  answer,
	}
	if err := s.blockRepo.CreateBatch(ctx, []*entity.Block{block}); err != nil {
		return nil, fmt.Errorf("create block: %w", err)
	}

	if node.Status == entity.NodeStatusDraft {
		node.Status = entity.NodeStatusAnswered
		if err := s.nodeRepo.Update(ctx, node); err != nil {
			return nil, fmt.Errorf("update node status: %w", err)
		}
	}

	if s.summaryDispatcher != nil {
		s.summaryDispatcher.EnqueueForNode(node.TreeID, node.ID)
	}

	return &dto.ChatResponse{
		QAPair: dto.QAPairResponse{
			ID:        qaPair.ID,
			Question:  qaPair.Question,
			CreatedAt: qaPair.CreatedAt,
			Blocks: []dto.BlockResponse{
				{
					ID:      block.ID,
					Type:    string(block.Type),
					Content: block.Content,
				},
			},
		},
	}, nil
}

func (s *ChatService) buildLatestQAPairResponse(ctx context.Context, nodeID string) (*dto.QAPairResponse, error) {
	qaPairs, err := s.qaPairRepo.GetByNodeID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("get qa_pairs: %w", err)
	}
	if len(qaPairs) == 0 {
		return nil, fmt.Errorf("qa_pair not found")
	}

	latest := qaPairs[len(qaPairs)-1]
	blocks, err := s.blockRepo.GetByQAPairID(ctx, latest.ID)
	if err != nil {
		return nil, fmt.Errorf("get blocks: %w", err)
	}

	blockResponses := make([]dto.BlockResponse, len(blocks))
	for i, block := range blocks {
		blockResponses[i] = dto.BlockResponse{
			ID:      block.ID,
			Type:    string(block.Type),
			Content: block.Content,
		}
	}

	return &dto.QAPairResponse{
		ID:        latest.ID,
		Question:  latest.Question,
		Blocks:    blockResponses,
		CreatedAt: latest.CreatedAt,
	}, nil
}
