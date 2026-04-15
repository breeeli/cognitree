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

func TestChatServiceStreamPersistsCompletedAnswer(t *testing.T) {
	nodeRepo := &chatStreamNodeRepo{
		node: &entity.Node{ID: "node-1", TreeID: "tree-1"},
	}
	qaPairRepo := &chatStreamQAPairRepo{}
	blockRepo := &chatStreamBlockRepo{}
	builder := &chatStreamContextBuilder{
		payload: &domainservice.ContextPayload{
			SystemPrompt: "system prompt",
			UserPrompt:   "user prompt",
		},
	}
	aiClient := &chatStreamAIClient{
		chunks: []string{"Hello ", "world"},
	}

	service := NewChatService(nodeRepo, qaPairRepo, blockRepo, builder, aiClient, nil)

	var events []dto.ChatStreamEvent
	err := service.StreamChat(context.Background(), "node-1", dto.ChatRequest{Question: "What is this?"}, func(event dto.ChatStreamEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatalf("StreamChat returned error: %v", err)
	}

	if len(qaPairRepo.created) != 1 {
		t.Fatalf("expected one qa_pair to be created, got %d", len(qaPairRepo.created))
	}
	if len(blockRepo.created) != 1 {
		t.Fatalf("expected one block to be created, got %d", len(blockRepo.created))
	}
	if got, want := blockRepo.created[0].Content, "Hello world"; got != want {
		t.Fatalf("unexpected block content: got %q want %q", got, want)
	}

	types := make([]string, 0, len(events))
	for _, event := range events {
		types = append(types, event.Type)
	}
	if got := strings.Join(types, ","); got != "answer_delta,answer_delta,completed" {
		t.Fatalf("unexpected event sequence: %s", got)
	}
}

func TestChatServiceStreamDoesNotPersistQAPairOnStreamFailure(t *testing.T) {
	nodeRepo := &chatStreamNodeRepo{
		node: &entity.Node{ID: "node-1", TreeID: "tree-1"},
	}
	qaPairRepo := &chatStreamQAPairRepo{}
	blockRepo := &chatStreamBlockRepo{}
	builder := &chatStreamContextBuilder{
		payload: &domainservice.ContextPayload{
			SystemPrompt: "system prompt",
			UserPrompt:   "user prompt",
		},
	}
	aiClient := &chatStreamAIClient{
		chunks: []string{"partial "},
		err:    errors.New("stream broken"),
	}

	service := NewChatService(nodeRepo, qaPairRepo, blockRepo, builder, aiClient, nil)

	var events []dto.ChatStreamEvent
	err := service.StreamChat(context.Background(), "node-1", dto.ChatRequest{Question: "What is this?"}, func(event dto.ChatStreamEvent) error {
		events = append(events, event)
		return nil
	})
	if err == nil {
		t.Fatal("expected error from StreamChat")
	}

	if len(qaPairRepo.created) != 0 {
		t.Fatalf("expected no qa_pair to be created, got %d", len(qaPairRepo.created))
	}
	if len(blockRepo.created) != 0 {
		t.Fatalf("expected no block to be created, got %d", len(blockRepo.created))
	}
	if len(events) == 0 || events[len(events)-1].Type != "error" {
		t.Fatalf("expected final error event, got %#v", events)
	}
}

type chatStreamNodeRepo struct {
	node *entity.Node
}

func (r *chatStreamNodeRepo) Create(ctx context.Context, node *entity.Node) error {
	return nil
}

func (r *chatStreamNodeRepo) GetByID(ctx context.Context, id string) (*entity.Node, error) {
	if r.node != nil && r.node.ID == id {
		return r.node, nil
	}
	return nil, errors.New("node not found")
}

func (r *chatStreamNodeRepo) GetByTreeID(ctx context.Context, treeID string) ([]*entity.Node, error) {
	return []*entity.Node{r.node}, nil
}

func (r *chatStreamNodeRepo) GetChildren(context.Context, string) ([]*entity.Node, error) {
	return nil, nil
}

func (r *chatStreamNodeRepo) Update(ctx context.Context, node *entity.Node) error {
	r.node = node
	return nil
}

func (r *chatStreamNodeRepo) Delete(context.Context, string) error { return nil }
func (r *chatStreamNodeRepo) DeleteByTreeID(context.Context, string) error {
	return nil
}

type chatStreamQAPairRepo struct {
	created []*entity.QAPair
}

func (r *chatStreamQAPairRepo) Create(ctx context.Context, qaPair *entity.QAPair) error {
	if qaPair.ID == "" {
		qaPair.ID = "qa-1"
	}
	qaPair.CreatedAt = time.Now()
	r.created = append(r.created, qaPair)
	return nil
}

func (r *chatStreamQAPairRepo) GetByID(context.Context, string) (*entity.QAPair, error) {
	return nil, nil
}
func (r *chatStreamQAPairRepo) GetByNodeID(context.Context, string) ([]*entity.QAPair, error) {
	return nil, nil
}
func (r *chatStreamQAPairRepo) GetLatestByNodeID(context.Context, string) (*entity.QAPair, error) {
	return nil, nil
}
func (r *chatStreamQAPairRepo) DeleteByNodeID(context.Context, string) error { return nil }

type chatStreamBlockRepo struct {
	created []*entity.Block
}

func (r *chatStreamBlockRepo) CreateBatch(ctx context.Context, blocks []*entity.Block) error {
	for _, block := range blocks {
		if block.ID == "" {
			block.ID = "block-1"
		}
		r.created = append(r.created, block)
	}
	return nil
}

func (r *chatStreamBlockRepo) GetByID(context.Context, string) (*entity.Block, error) {
	return nil, nil
}
func (r *chatStreamBlockRepo) GetByQAPairID(context.Context, string) ([]*entity.Block, error) {
	return nil, nil
}
func (r *chatStreamBlockRepo) DeleteByQAPairID(context.Context, string) error { return nil }

type chatStreamContextBuilder struct {
	payload *domainservice.ContextPayload
}

func (b *chatStreamContextBuilder) BuildContext(ctx context.Context, treeID string, currentNodeID string, newQuestion string) (*domainservice.ContextPayload, error) {
	return b.payload, nil
}

type chatStreamAIClient struct {
	chunks []string
	err    error
}

func (c *chatStreamAIClient) Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	return strings.Join(c.chunks, ""), c.err
}

func (c *chatStreamAIClient) ChatStream(ctx context.Context, systemPrompt, userPrompt string, onDelta func(string) error) error {
	for _, chunk := range c.chunks {
		if err := onDelta(chunk); err != nil {
			return err
		}
	}
	return c.err
}
