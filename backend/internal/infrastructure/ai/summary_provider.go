package ai

import "context"

type summaryProvider interface {
	GetNodeSummary(ctx context.Context, nodeID string) (string, error)
	GetSiblingSummaries(ctx context.Context, currentNodeID string) ([]string, error)
	GetRelevantSummaries(ctx context.Context, currentNodeID string, question string) ([]string, error)
}

type noopSummaryProvider struct{}

func newNoopSummaryProvider() summaryProvider {
	return noopSummaryProvider{}
}

func (noopSummaryProvider) GetNodeSummary(context.Context, string) (string, error) {
	return "", nil
}

func (noopSummaryProvider) GetSiblingSummaries(context.Context, string) ([]string, error) {
	return nil, nil
}

func (noopSummaryProvider) GetRelevantSummaries(context.Context, string, string) ([]string, error) {
	return nil, nil
}
