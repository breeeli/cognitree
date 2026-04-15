package repository

import (
	"context"
	"errors"
	"time"

	"github.com/cognitree/backend/internal/domain/entity"
)

var ErrSummaryAlreadyExists = errors.New("summary already exists")
var ErrSummaryNotFound = errors.New("summary not found")

type SummaryRepository interface {
	Create(ctx context.Context, summary *entity.Summary) error
	Update(ctx context.Context, summary *entity.Summary) error
	GetByID(ctx context.Context, id string) (*entity.Summary, error)
	GetLatestByScopeAndTarget(ctx context.Context, scope entity.SummaryScope, targetNodeID string) (*entity.Summary, error)
	ListByStatus(ctx context.Context, status entity.SummaryStatus) ([]*entity.Summary, error)
	ListStalePendingBefore(ctx context.Context, before time.Time) ([]*entity.Summary, error)
}
