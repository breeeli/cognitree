package repository

import (
	"context"

	"github.com/cognitree/backend/internal/domain/entity"
)

type BlockRepository interface {
	CreateBatch(ctx context.Context, blocks []*entity.Block) error
	GetByID(ctx context.Context, id string) (*entity.Block, error)
	GetByQAPairID(ctx context.Context, qaPairID string) ([]*entity.Block, error)
	DeleteByQAPairID(ctx context.Context, qaPairID string) error
}
