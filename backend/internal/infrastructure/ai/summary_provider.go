package ai

import (
	"context"
	"errors"
	"fmt"

	"github.com/cognitree/backend/internal/domain/entity"
	"github.com/cognitree/backend/internal/domain/repository"
)

type summaryProvider interface {
	GetNodeSummary(ctx context.Context, nodeID string) (string, bool, error)
	GetPathSummary(ctx context.Context, nodeID string) (string, bool, error)
	GetSubtreeSummary(ctx context.Context, nodeID string) (string, bool, error)
}

type repositorySummaryProvider struct {
	summaryRepo repository.SummaryRepository
}

func newRepositorySummaryProvider(summaryRepo repository.SummaryRepository) summaryProvider {
	return repositorySummaryProvider{summaryRepo: summaryRepo}
}

type noopSummaryProvider struct{}

func newNoopSummaryProvider() summaryProvider {
	return noopSummaryProvider{}
}

func (p repositorySummaryProvider) GetNodeSummary(ctx context.Context, nodeID string) (string, bool, error) {
	return p.getSummary(ctx, entity.SummaryScopeNode, nodeID)
}

func (p repositorySummaryProvider) GetPathSummary(ctx context.Context, nodeID string) (string, bool, error) {
	return p.getSummary(ctx, entity.SummaryScopePath, nodeID)
}

func (p repositorySummaryProvider) GetSubtreeSummary(ctx context.Context, nodeID string) (string, bool, error) {
	return p.getSummary(ctx, entity.SummaryScopeSubtree, nodeID)
}

func (p repositorySummaryProvider) getSummary(ctx context.Context, scope entity.SummaryScope, nodeID string) (string, bool, error) {
	summary, err := p.summaryRepo.GetLatestByScopeAndTarget(ctx, scope, nodeID)
	if err != nil {
		if errors.Is(err, repository.ErrSummaryNotFound) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("%s summary unavailable: %w", scope, err)
	}
	if summary.Status != entity.SummaryStatusReady {
		return "", false, nil
	}
	return summary.Content, true, nil
}

func (noopSummaryProvider) GetNodeSummary(context.Context, string) (string, bool, error) {
	return "", false, nil
}

func (noopSummaryProvider) GetPathSummary(context.Context, string) (string, bool, error) {
	return "", false, nil
}

func (noopSummaryProvider) GetSubtreeSummary(context.Context, string) (string, bool, error) {
	return "", false, nil
}
