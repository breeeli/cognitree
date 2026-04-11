package repository

import (
	"context"
	"fmt"

	"github.com/cognitree/backend/internal/domain/entity"
	"github.com/cognitree/backend/internal/infrastructure/persistence/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type blockRepository struct {
	db *gorm.DB
}

func NewBlockRepository(db *gorm.DB) *blockRepository {
	return &blockRepository{db: db}
}

func (r *blockRepository) CreateBatch(ctx context.Context, blocks []*entity.Block) error {
	if len(blocks) == 0 {
		return nil
	}
	models := make([]model.Block, len(blocks))
	for i, b := range blocks {
		if b.ID == "" {
			b.ID = uuid.New().String()
		}
		models[i] = model.Block{
			ID:        b.ID,
			QAPairID:  b.QAPairID,
			Type:      string(b.Type),
			Content:   b.Content,
			CreatedAt: b.CreatedAt,
		}
	}
	if err := r.db.WithContext(ctx).Create(&models).Error; err != nil {
		return fmt.Errorf("create blocks: %w", err)
	}
	for i := range blocks {
		blocks[i].CreatedAt = models[i].CreatedAt
	}
	return nil
}

func (r *blockRepository) GetByID(ctx context.Context, id string) (*entity.Block, error) {
	var m model.Block
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("get block: %w", err)
	}
	return toBlockEntity(&m), nil
}

func (r *blockRepository) GetByQAPairID(ctx context.Context, qaPairID string) ([]*entity.Block, error) {
	var models []model.Block
	if err := r.db.WithContext(ctx).Where("qa_pair_id = ?", qaPairID).Order("created_at ASC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("get blocks by qa_pair: %w", err)
	}
	blocks := make([]*entity.Block, len(models))
	for i := range models {
		blocks[i] = toBlockEntity(&models[i])
	}
	return blocks, nil
}

func (r *blockRepository) DeleteByQAPairID(ctx context.Context, qaPairID string) error {
	if err := r.db.WithContext(ctx).Where("qa_pair_id = ?", qaPairID).Delete(&model.Block{}).Error; err != nil {
		return fmt.Errorf("delete blocks by qa_pair: %w", err)
	}
	return nil
}

func toBlockEntity(m *model.Block) *entity.Block {
	return &entity.Block{
		ID:        m.ID,
		QAPairID:  m.QAPairID,
		Type:      entity.BlockType(m.Type),
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}
}
