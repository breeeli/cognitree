package repository

import (
	"context"

	"github.com/cognitree/backend/internal/domain/entity"
)

type QAPairRepository interface {
	Create(ctx context.Context, qaPair *entity.QAPair) error
	GetByID(ctx context.Context, id string) (*entity.QAPair, error)
	GetByNodeID(ctx context.Context, nodeID string) ([]*entity.QAPair, error)
	GetLatestByNodeID(ctx context.Context, nodeID string) (*entity.QAPair, error)
	DeleteByNodeID(ctx context.Context, nodeID string) error
}
