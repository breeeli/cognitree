package repository

import (
	"context"
	"fmt"

	"github.com/cognitree/backend/internal/domain/entity"
	"github.com/cognitree/backend/internal/infrastructure/persistence/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type treeRepository struct {
	db *gorm.DB
}

func NewTreeRepository(db *gorm.DB) *treeRepository {
	return &treeRepository{db: db}
}

func (r *treeRepository) Create(ctx context.Context, tree *entity.Tree) error {
	if tree.ID == "" {
		tree.ID = uuid.New().String()
	}
	m := toTreeModel(tree)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return fmt.Errorf("create tree: %w", err)
	}
	tree.CreatedAt = m.CreatedAt
	tree.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *treeRepository) GetByID(ctx context.Context, id string) (*entity.Tree, error) {
	var m model.Tree
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("get tree: %w", err)
	}
	return toTreeEntity(&m), nil
}

func (r *treeRepository) List(ctx context.Context) ([]*entity.Tree, error) {
	var models []model.Tree
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list trees: %w", err)
	}
	trees := make([]*entity.Tree, len(models))
	for i := range models {
		trees[i] = toTreeEntity(&models[i])
	}
	return trees, nil
}

func (r *treeRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&model.Tree{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete tree: %w", err)
	}
	return nil
}

func toTreeModel(e *entity.Tree) *model.Tree {
	return &model.Tree{
		ID:          e.ID,
		Title:       e.Title,
		Description: e.Description,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func toTreeEntity(m *model.Tree) *entity.Tree {
	return &entity.Tree{
		ID:          m.ID,
		Title:       m.Title,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
