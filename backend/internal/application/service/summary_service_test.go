package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cognitree/backend/internal/domain/entity"
	domainrepo "github.com/cognitree/backend/internal/domain/repository"
)

func TestSummaryServiceRetriesUntilSuccess(t *testing.T) {
	repo := newMemorySummaryRepo()
	ai := &fakeSummaryAI{
		failuresBeforeSuccess: 2,
		success:               "node summary content",
	}

	service := &SummaryService{
		treeRepo:             stubSummaryTreeRepo{},
		nodeRepo:             stubSummaryNodeRepo{},
		qaPairRepo:           stubSummaryQAPairRepo{},
		blockRepo:            stubSummaryBlockRepo{},
		summaryRepo:          repo,
		aiClient:             ai,
		queue:                make(chan summaryJob, 1),
		maxImmediateAttempts: 3,
		maxTotalAttempts:     6,
		workerBackoffs:       []time.Duration{0, 0, 0},
	}

	job := summaryJob{treeID: "tree-1", targetNodeID: "node-1", scope: entity.SummaryScopeNode}
	service.processJob(context.Background(), job)

	summary, err := repo.GetLatestByScopeAndTarget(context.Background(), entity.SummaryScopeNode, "node-1")
	if err != nil {
		t.Fatalf("GetLatestByScopeAndTarget returned error: %v", err)
	}
	if summary.Status != entity.SummaryStatusReady {
		t.Fatalf("expected ready summary, got %s", summary.Status)
	}
	if summary.AttemptCount != 3 {
		t.Fatalf("expected attempt count 3, got %d", summary.AttemptCount)
	}
	if summary.Content != "node summary content" {
		t.Fatalf("unexpected content: %q", summary.Content)
	}
	if ai.calls != 3 {
		t.Fatalf("expected 3 ai calls, got %d", ai.calls)
	}
	if !containsPrompt(ai.lastUserPrompt, "Tree Context") {
		t.Fatalf("expected tree context in prompt, got %q", ai.lastUserPrompt)
	}
	if !containsPrompt(ai.lastUserPrompt, "Node Focus") {
		t.Fatalf("expected node focus in prompt, got %q", ai.lastUserPrompt)
	}
}

func TestSummaryServiceCompensatesFailedSummaries(t *testing.T) {
	repo := newMemorySummaryRepo()
	repo.seed(&entity.Summary{
		ID:           "summary-1",
		TreeID:       "tree-1",
		TargetNodeID: "node-1",
		Scope:        entity.SummaryScopeNode,
		Version:      1,
		AttemptCount: 3,
		Status:       entity.SummaryStatusFailed,
	})

	service := &SummaryService{
		treeRepo:             stubSummaryTreeRepo{},
		summaryRepo:          repo,
		queue:                make(chan summaryJob, 1),
		maxImmediateAttempts: 3,
		maxTotalAttempts:     6,
	}

	service.compensateFailedSummaries(context.Background())

	if len(service.queue) != 1 {
		t.Fatalf("expected one job to be re-enqueued, got %d", len(service.queue))
	}

	job := <-service.queue
	if job.treeID != "tree-1" || job.targetNodeID != "node-1" || job.scope != entity.SummaryScopeNode {
		t.Fatalf("unexpected compensator job: %+v", job)
	}
}

func TestSummaryServiceCompensatesStalePendingSummaries(t *testing.T) {
	repo := newMemorySummaryRepo()
	repo.seed(&entity.Summary{
		ID:           "summary-1",
		TreeID:       "tree-1",
		TargetNodeID: "node-1",
		Scope:        entity.SummaryScopeNode,
		Version:      1,
		AttemptCount: 2,
		Status:       entity.SummaryStatusPending,
		UpdatedAt:    time.Now().Add(-time.Minute),
	})

	service := &SummaryService{
		treeRepo:         stubSummaryTreeRepo{},
		summaryRepo:      repo,
		queue:            make(chan summaryJob, 1),
		compensateEvery:  30 * time.Second,
		maxTotalAttempts: 6,
	}

	service.compensateFailedSummaries(context.Background())

	if len(service.queue) != 1 {
		t.Fatalf("expected one stale pending job to be re-enqueued, got %d", len(service.queue))
	}

	job := <-service.queue
	if job.treeID != "tree-1" || job.targetNodeID != "node-1" || job.scope != entity.SummaryScopeNode {
		t.Fatalf("unexpected compensator job: %+v", job)
	}
}

type fakeSummaryAI struct {
	mu                    sync.Mutex
	calls                 int
	failuresBeforeSuccess int
	success               string
	lastSystemPrompt      string
	lastUserPrompt        string
}

func (f *fakeSummaryAI) Chat(_ context.Context, systemPrompt, userPrompt string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls++
	f.lastSystemPrompt = systemPrompt
	f.lastUserPrompt = userPrompt
	if f.calls <= f.failuresBeforeSuccess {
		return "", errors.New("temporary summary generation failure")
	}
	return f.success, nil
}

func (f *fakeSummaryAI) ChatStream(ctx context.Context, systemPrompt, userPrompt string, onDelta func(string) error) error {
	_, err := f.Chat(ctx, systemPrompt, userPrompt)
	return err
}

type memorySummaryRepo struct {
	mu        sync.Mutex
	summaries map[string]*entity.Summary
}

func newMemorySummaryRepo() *memorySummaryRepo {
	return &memorySummaryRepo{summaries: make(map[string]*entity.Summary)}
}

func (r *memorySummaryRepo) key(scope entity.SummaryScope, targetNodeID string) string {
	return string(scope) + ":" + targetNodeID
}

func (r *memorySummaryRepo) seed(summary *entity.Summary) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.summaries[r.key(summary.Scope, summary.TargetNodeID)] = summary
}

func (r *memorySummaryRepo) Create(ctx context.Context, summary *entity.Summary) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := r.key(summary.Scope, summary.TargetNodeID)
	if _, ok := r.summaries[key]; ok {
		return domainrepo.ErrSummaryAlreadyExists
	}
	r.summaries[key] = summary
	return nil
}

func (r *memorySummaryRepo) Update(ctx context.Context, summary *entity.Summary) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.summaries[r.key(summary.Scope, summary.TargetNodeID)] = summary
	return nil
}

func (r *memorySummaryRepo) GetByID(ctx context.Context, id string) (*entity.Summary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, summary := range r.summaries {
		if summary.ID == id {
			return summary, nil
		}
	}
	return nil, errors.New("not found")
}

func (r *memorySummaryRepo) GetLatestByScopeAndTarget(ctx context.Context, scope entity.SummaryScope, targetNodeID string) (*entity.Summary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	summary, ok := r.summaries[r.key(scope, targetNodeID)]
	if !ok {
		return nil, errors.New("not found")
	}
	return summary, nil
}

func (r *memorySummaryRepo) ListByStatus(ctx context.Context, status entity.SummaryStatus) ([]*entity.Summary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var summaries []*entity.Summary
	for _, summary := range r.summaries {
		if summary.Status == status {
			summaries = append(summaries, summary)
		}
	}
	return summaries, nil
}

func (r *memorySummaryRepo) ListStalePendingBefore(ctx context.Context, before time.Time) ([]*entity.Summary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var summaries []*entity.Summary
	for _, summary := range r.summaries {
		if summary.Status == entity.SummaryStatusPending && summary.UpdatedAt.Before(before) {
			summaries = append(summaries, summary)
		}
	}
	return summaries, nil
}

type stubSummaryNodeRepo struct{}

type stubSummaryTreeRepo struct{}

func (stubSummaryTreeRepo) Create(context.Context, *entity.Tree) error { return nil }
func (stubSummaryTreeRepo) GetByID(context.Context, string) (*entity.Tree, error) {
	return &entity.Tree{ID: "tree-1", Title: "Tree title", Description: "Tree description"}, nil
}
func (stubSummaryTreeRepo) List(context.Context) ([]*entity.Tree, error) { return nil, nil }
func (stubSummaryTreeRepo) Delete(context.Context, string) error         { return nil }

func (stubSummaryNodeRepo) Create(context.Context, *entity.Node) error { return nil }
func (stubSummaryNodeRepo) GetByID(context.Context, string) (*entity.Node, error) {
	return &entity.Node{ID: "node-1", TreeID: "tree-1", Question: "Node question", Status: entity.NodeStatusAnswered}, nil
}
func (stubSummaryNodeRepo) GetByTreeID(context.Context, string) ([]*entity.Node, error) {
	return []*entity.Node{{ID: "node-1", TreeID: "tree-1", Question: "Node question", Status: entity.NodeStatusAnswered}}, nil
}
func (stubSummaryNodeRepo) GetChildren(context.Context, string) ([]*entity.Node, error) {
	return nil, nil
}
func (stubSummaryNodeRepo) Update(context.Context, *entity.Node) error { return nil }
func (stubSummaryNodeRepo) Delete(context.Context, string) error       { return nil }
func (stubSummaryNodeRepo) DeleteByTreeID(context.Context, string) error {
	return nil
}

type stubSummaryQAPairRepo struct{}

func (stubSummaryQAPairRepo) Create(context.Context, *entity.QAPair) error { return nil }
func (stubSummaryQAPairRepo) GetByID(context.Context, string) (*entity.QAPair, error) {
	return nil, nil
}
func (stubSummaryQAPairRepo) GetByNodeID(context.Context, string) ([]*entity.QAPair, error) {
	return []*entity.QAPair{{ID: "qa-1", NodeID: "node-1", Question: "What is this?"}}, nil
}
func (stubSummaryQAPairRepo) GetLatestByNodeID(context.Context, string) (*entity.QAPair, error) {
	return nil, nil
}
func (stubSummaryQAPairRepo) DeleteByNodeID(context.Context, string) error { return nil }

type stubSummaryBlockRepo struct{}

func (stubSummaryBlockRepo) CreateBatch(context.Context, []*entity.Block) error { return nil }
func (stubSummaryBlockRepo) GetByID(context.Context, string) (*entity.Block, error) {
	return nil, nil
}
func (stubSummaryBlockRepo) GetByQAPairID(context.Context, string) ([]*entity.Block, error) {
	return []*entity.Block{{ID: "block-1", QAPairID: "qa-1", Content: "This is the answer."}}, nil
}
func (stubSummaryBlockRepo) DeleteByQAPairID(context.Context, string) error { return nil }

func containsPrompt(got string, want string) bool {
	return strings.Contains(got, want)
}
