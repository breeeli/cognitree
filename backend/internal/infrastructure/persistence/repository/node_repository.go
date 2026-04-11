package repository

import (
	"context"
	"fmt"

	"github.com/cognitree/backend/internal/domain/entity"
	"github.com/cognitree/backend/internal/infrastructure/persistence/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type nodeRepository struct {
	db *gorm.DB
}

func NewNodeRepository(db *gorm.DB) *nodeRepository {
	return &nodeRepository{db: db}
}

func (r *nodeRepository) Create(ctx context.Context, node *entity.Node) error {
	if node.ID == "" {
		node.ID = uuid.New().String()
	}
	m := toNodeModel(node)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return fmt.Errorf("create node: %w", err)
	}
	node.CreatedAt = m.CreatedAt
	node.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *nodeRepository) GetByID(ctx context.Context, id string) (*entity.Node, error) {
	var m model.Node
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("get node: %w", err)
	}
	return toNodeEntity(&m), nil
}

func (r *nodeRepository) GetByTreeID(ctx context.Context, treeID string) ([]*entity.Node, error) {
	var models []model.Node
	if err := r.db.WithContext(ctx).Where("tree_id = ?", treeID).Order("created_at ASC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("get nodes by tree: %w", err)
	}
	nodes := make([]*entity.Node, len(models))
	for i := range models {
		nodes[i] = toNodeEntity(&models[i])
	}
	return nodes, nil
}

func (r *nodeRepository) GetChildren(ctx context.Context, parentNodeID string) ([]*entity.Node, error) {
	var models []model.Node
	if err := r.db.WithContext(ctx).Where("parent_node_id = ?", parentNodeID).Order("created_at ASC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("get children: %w", err)
	}
	nodes := make([]*entity.Node, len(models))
	for i := range models {
		nodes[i] = toNodeEntity(&models[i])
	}
	return nodes, nil
}

func (r *nodeRepository) Update(ctx context.Context, node *entity.Node) error {
	m := toNodeModel(node)
	if err := r.db.WithContext(ctx).Save(m).Error; err != nil {
		return fmt.Errorf("update node: %w", err)
	}
	node.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *nodeRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&model.Node{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete node: %w", err)
	}
	return nil
}

func (r *nodeRepository) DeleteByTreeID(ctx context.Context, treeID string) error {
	if err := r.db.WithContext(ctx).Where("tree_id = ?", treeID).Delete(&model.Node{}).Error; err != nil {
		return fmt.Errorf("delete nodes by tree: %w", err)
	}
	return nil
}

func toNodeModel(e *entity.Node) *model.Node {
	return &model.Node{
		ID:           e.ID,
		TreeID:       e.TreeID,
		ParentNodeID: e.ParentNodeID,
		AnchorID:     e.AnchorID,
		Question:     e.Question,
		Status:       string(e.Status),
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}
}

func toNodeEntity(m *model.Node) *entity.Node {
	return &entity.Node{
		ID:           m.ID,
		TreeID:       m.TreeID,
		ParentNodeID: m.ParentNodeID,
		AnchorID:     m.AnchorID,
		Question:     m.Question,
		Status:       entity.NodeStatus(m.Status),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}
