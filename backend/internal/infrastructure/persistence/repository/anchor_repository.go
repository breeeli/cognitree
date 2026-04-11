package repository

import (
	"context"
	"fmt"

	"github.com/cognitree/backend/internal/domain/entity"
	"github.com/cognitree/backend/internal/infrastructure/persistence/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type anchorRepository struct {
	db *gorm.DB
}

func NewAnchorRepository(db *gorm.DB) *anchorRepository {
	return &anchorRepository{db: db}
}

func (r *anchorRepository) Create(ctx context.Context, anchor *entity.Anchor) error {
	if anchor.ID == "" {
		anchor.ID = uuid.New().String()
	}
	m := toAnchorModel(anchor)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return fmt.Errorf("create anchor: %w", err)
	}
	anchor.CreatedAt = m.CreatedAt
	return nil
}

func (r *anchorRepository) Update(ctx context.Context, anchor *entity.Anchor) error {
	m := toAnchorModel(anchor)
	if err := r.db.WithContext(ctx).Save(m).Error; err != nil {
		return fmt.Errorf("update anchor: %w", err)
	}
	return nil
}

func (r *anchorRepository) GetByID(ctx context.Context, id string) (*entity.Anchor, error) {
	var m model.Anchor
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("get anchor: %w", err)
	}
	return toAnchorEntity(&m), nil
}

func (r *anchorRepository) GetByNodeID(ctx context.Context, sourceNodeID string) ([]*entity.Anchor, error) {
	var models []model.Anchor
	if err := r.db.WithContext(ctx).Where("source_node_id = ?", sourceNodeID).Order("created_at ASC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("get anchors by node: %w", err)
	}
	anchors := make([]*entity.Anchor, len(models))
	for i := range models {
		anchors[i] = toAnchorEntity(&models[i])
	}
	return anchors, nil
}

func (r *anchorRepository) GetByBlockID(ctx context.Context, blockID string) ([]*entity.Anchor, error) {
	var models []model.Anchor
	if err := r.db.WithContext(ctx).Where("block_id = ?", blockID).Order("created_at ASC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("get anchors by block: %w", err)
	}
	anchors := make([]*entity.Anchor, len(models))
	for i := range models {
		anchors[i] = toAnchorEntity(&models[i])
	}
	return anchors, nil
}

func (r *anchorRepository) DeleteByNodeID(ctx context.Context, sourceNodeID string) error {
	if err := r.db.WithContext(ctx).Where("source_node_id = ?", sourceNodeID).Delete(&model.Anchor{}).Error; err != nil {
		return fmt.Errorf("delete anchors by node: %w", err)
	}
	return nil
}

func toAnchorModel(e *entity.Anchor) *model.Anchor {
	return &model.Anchor{
		ID:           e.ID,
		BlockID:      e.BlockID,
		SourceNodeID: e.SourceNodeID,
		StartOffset:  e.StartOffset,
		EndOffset:    e.EndOffset,
		QuotedText:   e.QuotedText,
		ChildNodeID:  e.ChildNodeID,
		CreatedAt:    e.CreatedAt,
	}
}

func toAnchorEntity(m *model.Anchor) *entity.Anchor {
	return &entity.Anchor{
		ID:           m.ID,
		BlockID:      m.BlockID,
		SourceNodeID: m.SourceNodeID,
		StartOffset:  m.StartOffset,
		EndOffset:    m.EndOffset,
		QuotedText:   m.QuotedText,
		ChildNodeID:  m.ChildNodeID,
		CreatedAt:    m.CreatedAt,
	}
}
