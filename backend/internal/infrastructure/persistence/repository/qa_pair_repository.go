package repository

import (
	"context"
	"fmt"

	"github.com/cognitree/backend/internal/domain/entity"
	"github.com/cognitree/backend/internal/infrastructure/persistence/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type qaPairRepository struct {
	db *gorm.DB
}

func NewQAPairRepository(db *gorm.DB) *qaPairRepository {
	return &qaPairRepository{db: db}
}

func (r *qaPairRepository) Create(ctx context.Context, qaPair *entity.QAPair) error {
	if qaPair.ID == "" {
		qaPair.ID = uuid.New().String()
	}
	m := toQAPairModel(qaPair)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return fmt.Errorf("create qa_pair: %w", err)
	}
	qaPair.CreatedAt = m.CreatedAt
	return nil
}

func (r *qaPairRepository) GetByID(ctx context.Context, id string) (*entity.QAPair, error) {
	var m model.QAPair
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("get qa_pair: %w", err)
	}
	return toQAPairEntity(&m), nil
}

func (r *qaPairRepository) GetByNodeID(ctx context.Context, nodeID string) ([]*entity.QAPair, error) {
	var models []model.QAPair
	if err := r.db.WithContext(ctx).Where("node_id = ?", nodeID).Order("created_at ASC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("get qa_pairs by node: %w", err)
	}
	pairs := make([]*entity.QAPair, len(models))
	for i := range models {
		pairs[i] = toQAPairEntity(&models[i])
	}
	return pairs, nil
}

func (r *qaPairRepository) GetLatestByNodeID(ctx context.Context, nodeID string) (*entity.QAPair, error) {
	var m model.QAPair
	if err := r.db.WithContext(ctx).Where("node_id = ?", nodeID).Order("created_at DESC").First(&m).Error; err != nil {
		return nil, fmt.Errorf("get latest qa_pair: %w", err)
	}
	return toQAPairEntity(&m), nil
}

func (r *qaPairRepository) DeleteByNodeID(ctx context.Context, nodeID string) error {
	if err := r.db.WithContext(ctx).Where("node_id = ?", nodeID).Delete(&model.QAPair{}).Error; err != nil {
		return fmt.Errorf("delete qa_pairs by node: %w", err)
	}
	return nil
}

func toQAPairModel(e *entity.QAPair) *model.QAPair {
	return &model.QAPair{
		ID:        e.ID,
		NodeID:    e.NodeID,
		Question:  e.Question,
		CreatedAt: e.CreatedAt,
	}
}

func toQAPairEntity(m *model.QAPair) *entity.QAPair {
	return &entity.QAPair{
		ID:        m.ID,
		NodeID:    m.NodeID,
		Question:  m.Question,
		CreatedAt: m.CreatedAt,
	}
}
