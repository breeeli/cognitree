package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/cognitree/backend/internal/domain/entity"
	domainrepo "github.com/cognitree/backend/internal/domain/repository"
	"github.com/cognitree/backend/internal/infrastructure/persistence/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type summaryRepository struct {
	db *gorm.DB
}

func NewSummaryRepository(db *gorm.DB) *summaryRepository {
	return &summaryRepository{db: db}
}

func (r *summaryRepository) Create(ctx context.Context, summary *entity.Summary) error {
	if summary.ID == "" {
		summary.ID = uuid.New().String()
	}
	m := toSummaryModel(summary)
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "scope"}, {Name: "target_node_id"}, {Name: "version"}},
		DoNothing: true,
	}).Create(m)
	if result.Error != nil {
		return fmt.Errorf("create summary: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domainrepo.ErrSummaryAlreadyExists
	}
	if err := result.Error; err != nil {
		return fmt.Errorf("create summary: %w", err)
	}
	summary.CreatedAt = m.CreatedAt
	summary.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *summaryRepository) Update(ctx context.Context, summary *entity.Summary) error {
	m := toSummaryModel(summary)
	if err := r.db.WithContext(ctx).Save(m).Error; err != nil {
		return fmt.Errorf("update summary: %w", err)
	}
	summary.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *summaryRepository) GetByID(ctx context.Context, id string) (*entity.Summary, error) {
	var m model.Summary
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("get summary: %w", err)
	}
	return toSummaryEntity(&m), nil
}

func (r *summaryRepository) GetLatestByScopeAndTarget(ctx context.Context, scope entity.SummaryScope, targetNodeID string) (*entity.Summary, error) {
	var m model.Summary
	tx := r.db.WithContext(ctx).
		Where("scope = ? AND target_node_id = ?", string(scope), targetNodeID).
		Order("version DESC, created_at DESC").
		Find(&m)
	if tx.Error != nil {
		return nil, fmt.Errorf("get latest summary: %w", tx.Error)
	}
	if tx.RowsAffected == 0 {
		return nil, domainrepo.ErrSummaryNotFound
	}
	return toSummaryEntity(&m), nil
}

func (r *summaryRepository) ListByStatus(ctx context.Context, status entity.SummaryStatus) ([]*entity.Summary, error) {
	var models []model.Summary
	if err := r.db.WithContext(ctx).Where("status = ?", string(status)).Order("updated_at ASC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list summaries by status: %w", err)
	}
	summaries := make([]*entity.Summary, len(models))
	for i := range models {
		summaries[i] = toSummaryEntity(&models[i])
	}
	return summaries, nil
}

func (r *summaryRepository) ListStalePendingBefore(ctx context.Context, before time.Time) ([]*entity.Summary, error) {
	var models []model.Summary
	if err := r.db.WithContext(ctx).
		Where("status = ? AND updated_at < ?", string(entity.SummaryStatusPending), before).
		Order("updated_at ASC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list stale pending summaries: %w", err)
	}
	summaries := make([]*entity.Summary, len(models))
	for i := range models {
		summaries[i] = toSummaryEntity(&models[i])
	}
	return summaries, nil
}

func toSummaryModel(e *entity.Summary) *model.Summary {
	return &model.Summary{
		ID:           e.ID,
		TreeID:       e.TreeID,
		TargetNodeID: e.TargetNodeID,
		Scope:        string(e.Scope),
		Version:      e.Version,
		AttemptCount: e.AttemptCount,
		Status:       string(e.Status),
		Content:      e.Content,
		ErrorMessage: e.ErrorMessage,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}
}

func toSummaryEntity(m *model.Summary) *entity.Summary {
	return &entity.Summary{
		ID:           m.ID,
		TreeID:       m.TreeID,
		TargetNodeID: m.TargetNodeID,
		Scope:        entity.SummaryScope(m.Scope),
		Version:      m.Version,
		AttemptCount: m.AttemptCount,
		Status:       entity.SummaryStatus(m.Status),
		Content:      m.Content,
		ErrorMessage: m.ErrorMessage,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}
