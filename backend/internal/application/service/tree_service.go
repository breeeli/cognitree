package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/cognitree/backend/internal/application/dto"
	"github.com/cognitree/backend/internal/domain/entity"
	"github.com/cognitree/backend/internal/domain/repository"
)

type TreeService struct {
	treeRepo repository.TreeRepository
	nodeRepo repository.NodeRepository
}

func NewTreeService(treeRepo repository.TreeRepository, nodeRepo repository.NodeRepository) *TreeService {
	return &TreeService{treeRepo: treeRepo, nodeRepo: nodeRepo}
}

func (s *TreeService) Create(ctx context.Context, req dto.CreateTreeRequest) (*dto.CreateTreeResponse, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = deriveTreeTitle(req.Question)
	}

	tree := &entity.Tree{
		Title: title,
	}
	if err := s.treeRepo.Create(ctx, tree); err != nil {
		return nil, fmt.Errorf("create tree: %w", err)
	}

	rootNode := &entity.Node{
		TreeID:   tree.ID,
		Question: req.Question,
		Status:   entity.NodeStatusDraft,
	}
	if err := s.nodeRepo.Create(ctx, rootNode); err != nil {
		return nil, fmt.Errorf("create root node: %w", err)
	}

	return &dto.CreateTreeResponse{
		Tree:     toTreeResponse(tree),
		RootNode: toNodeResponse(rootNode),
	}, nil
}

func (s *TreeService) GetByID(ctx context.Context, id string) (*dto.TreeDetailResponse, error) {
	tree, err := s.treeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get tree: %w", err)
	}

	nodes, err := s.nodeRepo.GetByTreeID(ctx, tree.ID)
	if err != nil {
		return nil, fmt.Errorf("get nodes: %w", err)
	}

	nodeResponses := make([]dto.NodeResponse, len(nodes))
	for i, n := range nodes {
		nodeResponses[i] = toNodeResponse(n)
	}

	return &dto.TreeDetailResponse{
		Tree:  toTreeResponse(tree),
		Nodes: nodeResponses,
	}, nil
}

func (s *TreeService) List(ctx context.Context) ([]dto.TreeResponse, error) {
	trees, err := s.treeRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list trees: %w", err)
	}

	responses := make([]dto.TreeResponse, len(trees))
	for i, t := range trees {
		responses[i] = toTreeResponse(t)
	}
	return responses, nil
}

func (s *TreeService) Delete(ctx context.Context, id string) error {
	if err := s.nodeRepo.DeleteByTreeID(ctx, id); err != nil {
		return fmt.Errorf("delete nodes: %w", err)
	}
	if err := s.treeRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete tree: %w", err)
	}
	return nil
}

func toTreeResponse(t *entity.Tree) dto.TreeResponse {
	return dto.TreeResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func deriveTreeTitle(question string) string {
	trimmed := strings.TrimSpace(question)
	if trimmed == "" {
		return "Untitled tree"
	}

	runes := []rune(trimmed)
	if len(runes) <= 24 {
		return trimmed
	}

	return strings.TrimSpace(string(runes[:24])) + "…"
}

func toNodeResponse(n *entity.Node) dto.NodeResponse {
	return dto.NodeResponse{
		ID:           n.ID,
		TreeID:       n.TreeID,
		ParentNodeID: n.ParentNodeID,
		AnchorID:     n.AnchorID,
		Question:     n.Question,
		Status:       string(n.Status),
		CreatedAt:    n.CreatedAt,
		UpdatedAt:    n.UpdatedAt,
	}
}
