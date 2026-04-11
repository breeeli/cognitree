package service

import (
	"context"
	"fmt"

	"github.com/cognitree/backend/internal/application/dto"
	"github.com/cognitree/backend/internal/domain/entity"
	"github.com/cognitree/backend/internal/domain/repository"
)

type NodeService struct {
	nodeRepo   repository.NodeRepository
	qaPairRepo repository.QAPairRepository
	blockRepo  repository.BlockRepository
	anchorRepo repository.AnchorRepository
}

func NewNodeService(
	nodeRepo repository.NodeRepository,
	qaPairRepo repository.QAPairRepository,
	blockRepo repository.BlockRepository,
	anchorRepo repository.AnchorRepository,
) *NodeService {
	return &NodeService{
		nodeRepo:   nodeRepo,
		qaPairRepo: qaPairRepo,
		blockRepo:  blockRepo,
		anchorRepo: anchorRepo,
	}
}

func (s *NodeService) GetByID(ctx context.Context, id string) (*dto.NodeResponse, error) {
	node, err := s.nodeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get node: %w", err)
	}

	qaPairs, err := s.qaPairRepo.GetByNodeID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get qa_pairs: %w", err)
	}

	qaPairResponses := make([]dto.QAPairResponse, len(qaPairs))
	for i, qp := range qaPairs {
		blocks, err := s.blockRepo.GetByQAPairID(ctx, qp.ID)
		if err != nil {
			return nil, fmt.Errorf("get blocks: %w", err)
		}
		blockResponses := make([]dto.BlockResponse, len(blocks))
		for j, b := range blocks {
			blockResponses[j] = dto.BlockResponse{
				ID:      b.ID,
				Type:    string(b.Type),
				Content: b.Content,
			}
		}
		qaPairResponses[i] = dto.QAPairResponse{
			ID:        qp.ID,
			Question:  qp.Question,
			Blocks:    blockResponses,
			CreatedAt: qp.CreatedAt,
		}
	}

	resp := toNodeResponse(node)
	resp.QAPairs = qaPairResponses
	return &resp, nil
}

func (s *NodeService) CreateChild(ctx context.Context, parentNodeID string, req dto.CreateChildNodeRequest) (*dto.NodeResponse, error) {
	parent, err := s.nodeRepo.GetByID(ctx, parentNodeID)
	if err != nil {
		return nil, fmt.Errorf("get parent node: %w", err)
	}

	child := &entity.Node{
		TreeID:       parent.TreeID,
		ParentNodeID: &parentNodeID,
		AnchorID:     req.AnchorID,
		Question:     req.Question,
		Status:       entity.NodeStatusDraft,
	}
	if err := s.nodeRepo.Create(ctx, child); err != nil {
		return nil, fmt.Errorf("create child node: %w", err)
	}

	resp := toNodeResponse(child)
	return &resp, nil
}

func (s *NodeService) Delete(ctx context.Context, id string) error {
	children, err := s.nodeRepo.GetChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("get children: %w", err)
	}
	for _, child := range children {
		if err := s.Delete(ctx, child.ID); err != nil {
			return err
		}
	}

	if err := s.anchorRepo.DeleteByNodeID(ctx, id); err != nil {
		return fmt.Errorf("delete anchors: %w", err)
	}

	qaPairs, err := s.qaPairRepo.GetByNodeID(ctx, id)
	if err != nil {
		return fmt.Errorf("get qa_pairs for delete: %w", err)
	}
	for _, qp := range qaPairs {
		if err := s.blockRepo.DeleteByQAPairID(ctx, qp.ID); err != nil {
			return fmt.Errorf("delete blocks: %w", err)
		}
	}
	if err := s.qaPairRepo.DeleteByNodeID(ctx, id); err != nil {
		return fmt.Errorf("delete qa_pairs: %w", err)
	}

	if err := s.nodeRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete node: %w", err)
	}
	return nil
}

func (s *NodeService) GetThread(ctx context.Context, nodeID string) (*dto.ThreadResponse, error) {
	var thread []*entity.Node
	currentID := nodeID

	for {
		node, err := s.nodeRepo.GetByID(ctx, currentID)
		if err != nil {
			return nil, fmt.Errorf("get node in thread: %w", err)
		}
		thread = append([]*entity.Node{node}, thread...)
		if node.ParentNodeID == nil {
			break
		}
		currentID = *node.ParentNodeID
	}

	nodes := make([]dto.NodeResponse, len(thread))
	for i, n := range thread {
		nodes[i] = toNodeResponse(n)
	}
	return &dto.ThreadResponse{Nodes: nodes}, nil
}

func (s *NodeService) CreateAnchor(ctx context.Context, req dto.CreateAnchorRequest) (*dto.CreateAnchorResponse, error) {
	block, err := s.blockRepo.GetByID(ctx, req.BlockID)
	if err != nil {
		return nil, fmt.Errorf("get block: %w", err)
	}

	qaPair, err := s.qaPairRepo.GetByID(ctx, block.QAPairID)
	if err != nil {
		return nil, fmt.Errorf("get qa_pair: %w", err)
	}

	sourceNodeID := qaPair.NodeID

	anchor := &entity.Anchor{
		BlockID:      req.BlockID,
		SourceNodeID: sourceNodeID,
		StartOffset:  req.StartOffset,
		EndOffset:    req.EndOffset,
		QuotedText:   req.QuotedText,
	}
	if err := s.anchorRepo.Create(ctx, anchor); err != nil {
		return nil, fmt.Errorf("create anchor: %w", err)
	}

	parentNode, err := s.nodeRepo.GetByID(ctx, sourceNodeID)
	if err != nil {
		return nil, fmt.Errorf("get source node: %w", err)
	}

	childNode := &entity.Node{
		TreeID:       parentNode.TreeID,
		ParentNodeID: &sourceNodeID,
		AnchorID:     &anchor.ID,
		Question:     req.ChildQuestion,
		Status:       entity.NodeStatusDraft,
	}
	if err := s.nodeRepo.Create(ctx, childNode); err != nil {
		return nil, fmt.Errorf("create child node: %w", err)
	}

	anchor.ChildNodeID = &childNode.ID
	if err := s.anchorRepo.Update(ctx, anchor); err != nil {
		return nil, fmt.Errorf("update anchor with child: %w", err)
	}

	return &dto.CreateAnchorResponse{
		Anchor: dto.AnchorResponse{
			ID:           anchor.ID,
			BlockID:      anchor.BlockID,
			SourceNodeID: anchor.SourceNodeID,
			StartOffset:  anchor.StartOffset,
			EndOffset:    anchor.EndOffset,
			QuotedText:   anchor.QuotedText,
			ChildNodeID:  anchor.ChildNodeID,
		},
		ChildNode: toNodeResponse(childNode),
	}, nil
}

func (s *NodeService) GetAnchors(ctx context.Context, nodeID string) ([]dto.AnchorResponse, error) {
	anchors, err := s.anchorRepo.GetByNodeID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("get anchors: %w", err)
	}

	responses := make([]dto.AnchorResponse, len(anchors))
	for i, a := range anchors {
		responses[i] = dto.AnchorResponse{
			ID:           a.ID,
			BlockID:      a.BlockID,
			SourceNodeID: a.SourceNodeID,
			StartOffset:  a.StartOffset,
			EndOffset:    a.EndOffset,
			QuotedText:   a.QuotedText,
			ChildNodeID:  a.ChildNodeID,
		}
	}
	return responses, nil
}
