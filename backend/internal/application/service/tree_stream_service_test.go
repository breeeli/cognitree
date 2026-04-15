package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/cognitree/backend/internal/application/dto"
	"github.com/cognitree/backend/internal/domain/entity"
	domainservice "github.com/cognitree/backend/internal/domain/service"
)

func TestTreeStreamServiceCreatesTreeAndPersistsCompletedAnswer(t *testing.T) {
	treeRepo := &streamTreeRepo{}
	nodeRepo := &streamNodeRepo{}
	qaPairRepo := &streamQAPairRepo{}
	blockRepo := &streamBlockRepo{}
	builder := &streamContextBuilder{
		payload: &domainservice.ContextPayload{
			SystemPrompt: "system prompt",
			UserPrompt:   "user prompt",
		},
	}
	aiClient := &streamAIClient{
		chunks: []string{"Hello ", "world"},
	}

	service := NewTreeStreamService(treeRepo, nodeRepo, qaPairRepo, blockRepo, builder, aiClient, nil)

	var events []dto.TreeStreamEvent
	err := service.StreamFirstQuestion(context.Background(), dto.CreateTreeStreamRequest{
		Question: "What is a thinking tree?",
	}, func(event dto.TreeStreamEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatalf("StreamFirstQuestion returned error: %v", err)
	}

	if len(treeRepo.created) != 1 {
		t.Fatalf("expected one tree to be created, got %d", len(treeRepo.created))
	}
	if len(nodeRepo.created) != 1 {
		t.Fatalf("expected one root node to be created, got %d", len(nodeRepo.created))
	}
	if len(qaPairRepo.created) != 1 {
		t.Fatalf("expected one qa_pair to be created, got %d", len(qaPairRepo.created))
	}
	if len(blockRepo.created) != 1 {
		t.Fatalf("expected one block to be created, got %d", len(blockRepo.created))
	}

	if got, want := qaPairRepo.created[0].Question, "What is a thinking tree?"; got != want {
		t.Fatalf("unexpected qa_pair question: got %q want %q", got, want)
	}
	if got, want := blockRepo.created[0].Content, "Hello world"; got != want {
		t.Fatalf("unexpected block content: got %q want %q", got, want)
	}
	if nodeRepo.updated == nil || nodeRepo.updated.Status != entity.NodeStatusAnswered {
		t.Fatalf("expected root node status to be answered, got %#v", nodeRepo.updated)
	}

	types := make([]string, 0, len(events))
	for _, event := range events {
		types = append(types, event.Type)
	}
	gotTypes := strings.Join(types, ",")
	if gotTypes != "tree_ready,root_node_ready,answer_delta,answer_delta,completed" {
		t.Fatalf("unexpected event sequence: %s", gotTypes)
	}
}

func TestTreeStreamServiceDoesNotPersistQAPairOnStreamFailure(t *testing.T) {
	treeRepo := &streamTreeRepo{}
	nodeRepo := &streamNodeRepo{}
	qaPairRepo := &streamQAPairRepo{}
	blockRepo := &streamBlockRepo{}
	builder := &streamContextBuilder{
		payload: &domainservice.ContextPayload{
			SystemPrompt: "system prompt",
			UserPrompt:   "user prompt",
		},
	}
	aiClient := &streamAIClient{
		chunks: []string{"partial "},
		err:    errors.New("stream broken"),
	}

	service := NewTreeStreamService(treeRepo, nodeRepo, qaPairRepo, blockRepo, builder, aiClient, nil)

	var events []dto.TreeStreamEvent
	err := service.StreamFirstQuestion(context.Background(), dto.CreateTreeStreamRequest{
		Question: "Why does this fail?",
	}, func(event dto.TreeStreamEvent) error {
		events = append(events, event)
		return nil
	})
	if err == nil {
		t.Fatal("expected error from StreamFirstQuestion")
	}

	if len(treeRepo.created) != 1 {
		t.Fatalf("expected one tree to be created, got %d", len(treeRepo.created))
	}
	if len(nodeRepo.created) != 1 {
		t.Fatalf("expected one root node to be created, got %d", len(nodeRepo.created))
	}
	if len(qaPairRepo.created) != 0 {
		t.Fatalf("expected no qa_pair to be created, got %d", len(qaPairRepo.created))
	}
	if len(blockRepo.created) != 0 {
		t.Fatalf("expected no block to be created, got %d", len(blockRepo.created))
	}
	if nodeRepo.updated != nil {
		t.Fatalf("expected root node to remain draft, got %#v", nodeRepo.updated)
	}

	if len(events) == 0 || events[len(events)-1].Type != "error" {
		t.Fatalf("expected final error event, got %#v", events)
	}
}

type streamTreeRepo struct {
	created []*entity.Tree
	byID    map[string]*entity.Tree
}

func (r *streamTreeRepo) Create(ctx context.Context, tree *entity.Tree) error {
	if tree.ID == "" {
		tree.ID = "tree-1"
	}
	tree.CreatedAt = now()
	tree.UpdatedAt = tree.CreatedAt
	r.created = append(r.created, tree)
	if r.byID == nil {
		r.byID = map[string]*entity.Tree{}
	}
	r.byID[tree.ID] = tree
	return nil
}

func (r *streamTreeRepo) GetByID(ctx context.Context, id string) (*entity.Tree, error) {
	if r.byID != nil {
		if tree, ok := r.byID[id]; ok {
			return tree, nil
		}
	}
	return nil, errors.New("tree not found")
}

func (r *streamTreeRepo) List(context.Context) ([]*entity.Tree, error) { return nil, nil }
func (r *streamTreeRepo) Delete(context.Context, string) error         { return nil }

type streamNodeRepo struct {
	created []*entity.Node
	updated *entity.Node
	byID    map[string]*entity.Node
}

func (r *streamNodeRepo) Create(ctx context.Context, node *entity.Node) error {
	if node.ID == "" {
		node.ID = "node-1"
	}
	node.CreatedAt = now()
	node.UpdatedAt = node.CreatedAt
	r.created = append(r.created, node)
	if r.byID == nil {
		r.byID = map[string]*entity.Node{}
	}
	r.byID[node.ID] = node
	return nil
}

func (r *streamNodeRepo) GetByID(ctx context.Context, id string) (*entity.Node, error) {
	if r.byID != nil {
		if node, ok := r.byID[id]; ok {
			return node, nil
		}
	}
	return nil, errors.New("node not found")
}

func (r *streamNodeRepo) GetByTreeID(ctx context.Context, treeID string) ([]*entity.Node, error) {
	var nodes []*entity.Node
	for _, node := range r.created {
		if node.TreeID == treeID {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

func (r *streamNodeRepo) GetChildren(context.Context, string) ([]*entity.Node, error) {
	return nil, nil
}
func (r *streamNodeRepo) Update(ctx context.Context, node *entity.Node) error {
	r.updated = node
	return nil
}
func (r *streamNodeRepo) Delete(context.Context, string) error { return nil }
func (r *streamNodeRepo) DeleteByTreeID(context.Context, string) error {
	return nil
}

type streamQAPairRepo struct {
	created []*entity.QAPair
}

func (r *streamQAPairRepo) Create(ctx context.Context, qaPair *entity.QAPair) error {
	if qaPair.ID == "" {
		qaPair.ID = "qa-1"
	}
	qaPair.CreatedAt = now()
	r.created = append(r.created, qaPair)
	return nil
}

func (r *streamQAPairRepo) GetByID(context.Context, string) (*entity.QAPair, error) {
	return nil, nil
}
func (r *streamQAPairRepo) GetByNodeID(context.Context, string) ([]*entity.QAPair, error) {
	return nil, nil
}
func (r *streamQAPairRepo) GetLatestByNodeID(context.Context, string) (*entity.QAPair, error) {
	return nil, nil
}
func (r *streamQAPairRepo) DeleteByNodeID(context.Context, string) error { return nil }

type streamBlockRepo struct {
	created []*entity.Block
}

func (r *streamBlockRepo) CreateBatch(ctx context.Context, blocks []*entity.Block) error {
	for _, block := range blocks {
		if block.ID == "" {
			block.ID = "block-1"
		}
		r.created = append(r.created, block)
	}
	return nil
}

func (r *streamBlockRepo) GetByID(context.Context, string) (*entity.Block, error) {
	return nil, nil
}
func (r *streamBlockRepo) GetByQAPairID(context.Context, string) ([]*entity.Block, error) {
	return nil, nil
}
func (r *streamBlockRepo) DeleteByQAPairID(context.Context, string) error { return nil }

type streamContextBuilder struct {
	payload *domainservice.ContextPayload
}

func (b *streamContextBuilder) BuildContext(ctx context.Context, treeID string, currentNodeID string, newQuestion string) (*domainservice.ContextPayload, error) {
	return b.payload, nil
}

type streamAIClient struct {
	chunks []string
	err    error
}

func (c *streamAIClient) ChatStream(ctx context.Context, systemPrompt, userPrompt string, onDelta func(string) error) error {
	for _, chunk := range c.chunks {
		if err := onDelta(chunk); err != nil {
			return err
		}
	}
	return c.err
}

func now() time.Time {
	return time.Unix(1700000000, 0)
}
