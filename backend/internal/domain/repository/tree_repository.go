package repository

import (
	"context"

	"github.com/cognitree/backend/internal/domain/entity"
)

type TreeRepository interface {
	Create(ctx context.Context, tree *entity.Tree) error
	GetByID(ctx context.Context, id string) (*entity.Tree, error)
	List(ctx context.Context) ([]*entity.Tree, error)
	Delete(ctx context.Context, id string) error
}
