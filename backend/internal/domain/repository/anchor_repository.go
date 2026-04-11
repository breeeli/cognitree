package repository

import (
	"context"

	"github.com/cognitree/backend/internal/domain/entity"
)

type AnchorRepository interface {
	Create(ctx context.Context, anchor *entity.Anchor) error
	Update(ctx context.Context, anchor *entity.Anchor) error
	GetByID(ctx context.Context, id string) (*entity.Anchor, error)
	GetByNodeID(ctx context.Context, sourceNodeID string) ([]*entity.Anchor, error)
	GetByBlockID(ctx context.Context, blockID string) ([]*entity.Anchor, error)
	DeleteByNodeID(ctx context.Context, sourceNodeID string) error
}
