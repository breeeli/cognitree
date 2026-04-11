package repository

import (
	"context"

	"github.com/cognitree/backend/internal/domain/entity"
)

type NodeRepository interface {
	Create(ctx context.Context, node *entity.Node) error
	GetByID(ctx context.Context, id string) (*entity.Node, error)
	GetByTreeID(ctx context.Context, treeID string) ([]*entity.Node, error)
	GetChildren(ctx context.Context, parentNodeID string) ([]*entity.Node, error)
	Update(ctx context.Context, node *entity.Node) error
	Delete(ctx context.Context, id string) error
	DeleteByTreeID(ctx context.Context, treeID string) error
}
