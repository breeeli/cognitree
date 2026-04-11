package service

import (
	"context"
	"fmt"

	"github.com/cognitree/backend/internal/application/dto"
	"github.com/cognitree/backend/internal/domain/entity"
	"github.com/cognitree/backend/internal/domain/repository"
	domainservice "github.com/cognitree/backend/internal/domain/service"
)

type AIClient interface {
	Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

type ChatService struct {
	nodeRepo       repository.NodeRepository
	qaPairRepo     repository.QAPairRepository
	blockRepo      repository.BlockRepository
	contextBuilder domainservice.ContextBuilder
	aiClient       AIClient
}

func NewChatService(
	nodeRepo repository.NodeRepository,
	qaPairRepo repository.QAPairRepository,
	blockRepo repository.BlockRepository,
	contextBuilder domainservice.ContextBuilder,
	aiClient AIClient,
) *ChatService {
	return &ChatService{
		nodeRepo:       nodeRepo,
		qaPairRepo:     qaPairRepo,
		blockRepo:      blockRepo,
		contextBuilder: contextBuilder,
		aiClient:       aiClient,
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

	answer, err := s.aiClient.Chat(ctx, payload.SystemPrompt, payload.UserPrompt)
	if err != nil {
		return nil, fmt.Errorf("ai chat: %w", err)
	}

	qaPair := &entity.QAPair{
		NodeID:   nodeID,
		Question: req.Question,
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
