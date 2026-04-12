package ai

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/cognitree/backend/internal/domain/entity"
)

func TestBuildContextIncludesTreeGoalAndAnchorEvidence(t *testing.T) {
	rootID := "root"
	anchorID := "anchor-1"

	root := &entity.Node{
		ID:       rootID,
		TreeID:   "tree-1",
		Question: "What is a thinking tree?",
		Status:   entity.NodeStatusAnswered,
	}
	child := &entity.Node{
		ID:           "child",
		TreeID:       "tree-1",
		ParentNodeID: &rootID,
		AnchorID:     &anchorID,
		Question:     "How should context remember branch intent?",
		Status:       entity.NodeStatusDraft,
	}

	builder := &contextBuilder{
		treeRepo: stubTreeRepository{
			getByID: func(context.Context, string) (*entity.Tree, error) {
				return &entity.Tree{
					ID:          "tree-1",
					Title:       "Thinking Tree Research",
					Description: "Improve context construction for tree-based exploration",
				}, nil
			},
		},
		nodeRepo: stubNodeRepository{
			getByTreeID: func(context.Context, string) ([]*entity.Node, error) {
				return []*entity.Node{root, child}, nil
			},
		},
		qaPairRepo: stubQAPairRepository{
			getByNodeID: func(_ context.Context, nodeID string) ([]*entity.QAPair, error) {
				switch nodeID {
				case rootID:
					return []*entity.QAPair{{ID: "qa-root", NodeID: rootID, Question: "What is a thinking tree?"}}, nil
				case "child":
					return []*entity.QAPair{{ID: "qa-child", NodeID: "child", Question: "How should context remember branch intent?"}}, nil
				default:
					return nil, nil
				}
			},
		},
		blockRepo: stubBlockRepository{
			getByQAPairID: func(_ context.Context, qaPairID string) ([]*entity.Block, error) {
				switch qaPairID {
				case "qa-root":
					return []*entity.Block{{ID: "block-root", QAPairID: qaPairID, Content: "A thinking tree organizes exploration by nodes and branches."}}, nil
				case "qa-child":
					return []*entity.Block{{ID: "block-child", QAPairID: qaPairID, Content: "Branch intent should stay connected to the quoted evidence."}}, nil
				default:
					return nil, nil
				}
			},
		},
		anchorRepo: stubAnchorRepository{
			getByID: func(context.Context, string) (*entity.Anchor, error) {
				return &entity.Anchor{
					ID:           anchorID,
					BlockID:      "block-root",
					SourceNodeID: rootID,
					StartOffset:  12,
					EndOffset:    35,
					QuotedText:   "the quoted evidence",
				}, nil
			},
		},
		summaryProvider: newNoopSummaryProvider(),
	}

	payload, err := builder.BuildContext(context.Background(), "tree-1", "child", "What should we optimize first?")
	if err != nil {
		t.Fatalf("BuildContext returned error: %v", err)
	}

	if payload.Degraded {
		t.Fatalf("expected non-degraded payload, got warnings: %v", payload.Warnings)
	}

	assertContains(t, payload.UserPrompt, "## Tree Goal")
	assertContains(t, payload.UserPrompt, "Thinking Tree Research")
	assertContains(t, payload.UserPrompt, "## Anchor Evidence")
	assertContains(t, payload.UserPrompt, "the quoted evidence")
	assertContains(t, payload.UserPrompt, "## Current Ask")
	assertContains(t, payload.UserPrompt, "What should we optimize first?")
}

func TestBuildContextMarksRecoverableDegradation(t *testing.T) {
	root := &entity.Node{
		ID:       "root",
		TreeID:   "tree-1",
		Question: "Root question",
		Status:   entity.NodeStatusAnswered,
	}

	builder := &contextBuilder{
		treeRepo: stubTreeRepository{
			getByID: func(context.Context, string) (*entity.Tree, error) {
				return &entity.Tree{ID: "tree-1", Title: "Goal title"}, nil
			},
		},
		nodeRepo: stubNodeRepository{
			getByTreeID: func(context.Context, string) ([]*entity.Node, error) {
				return []*entity.Node{root}, nil
			},
		},
		qaPairRepo: stubQAPairRepository{
			getByNodeID: func(context.Context, string) ([]*entity.QAPair, error) {
				return nil, errors.New("db unavailable")
			},
		},
		blockRepo:       stubBlockRepository{},
		anchorRepo:      stubAnchorRepository{},
		summaryProvider: newNoopSummaryProvider(),
	}

	payload, err := builder.BuildContext(context.Background(), "tree-1", "root", "How should degradation behave?")
	if err != nil {
		t.Fatalf("BuildContext returned error: %v", err)
	}

	if !payload.Degraded {
		t.Fatalf("expected degraded payload")
	}
	if len(payload.Warnings) == 0 {
		t.Fatalf("expected warnings to be recorded")
	}
	if !strings.Contains(payload.Warnings[0], "qa_pairs unavailable") {
		t.Fatalf("expected qa_pairs warning, got %v", payload.Warnings)
	}

	assertContains(t, payload.UserPrompt, "## Tree Goal")
	assertContains(t, payload.UserPrompt, "## Current Ask")
}

func TestBuildContextMarksTreeGoalDegradation(t *testing.T) {
	root := &entity.Node{
		ID:       "root",
		TreeID:   "tree-1",
		Question: "Root question",
		Status:   entity.NodeStatusAnswered,
	}

	builder := &contextBuilder{
		treeRepo: stubTreeRepository{
			getByID: func(context.Context, string) (*entity.Tree, error) {
				return nil, errors.New("tree unavailable")
			},
		},
		nodeRepo: stubNodeRepository{
			getByTreeID: func(context.Context, string) ([]*entity.Node, error) {
				return []*entity.Node{root}, nil
			},
		},
		qaPairRepo: stubQAPairRepository{
			getByNodeID: func(context.Context, string) ([]*entity.QAPair, error) {
				return []*entity.QAPair{}, nil
			},
		},
		blockRepo:       stubBlockRepository{},
		anchorRepo:      stubAnchorRepository{},
		summaryProvider: newNoopSummaryProvider(),
	}

	payload, err := builder.BuildContext(context.Background(), "tree-1", "root", "How should tree fallback behave?")
	if err != nil {
		t.Fatalf("BuildContext returned error: %v", err)
	}

	if !payload.Degraded {
		t.Fatalf("expected degraded payload")
	}
	found := false
	for _, warning := range payload.Warnings {
		if strings.Contains(warning, "tree goal degraded") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected tree goal degradation warning, got %v", payload.Warnings)
	}
	assertContains(t, payload.UserPrompt, "## Tree Goal")
	assertContains(t, payload.UserPrompt, "Root Question: Root question")
}

func TestBuildContextMarksAnchorDegradation(t *testing.T) {
	rootID := "root"
	anchorID := "anchor-1"
	root := &entity.Node{
		ID:       rootID,
		TreeID:   "tree-1",
		Question: "Root question",
		Status:   entity.NodeStatusAnswered,
	}
	child := &entity.Node{
		ID:           "child",
		TreeID:       "tree-1",
		ParentNodeID: &rootID,
		AnchorID:     &anchorID,
		Question:     "Child question",
		Status:       entity.NodeStatusDraft,
	}

	builder := &contextBuilder{
		treeRepo: stubTreeRepository{
			getByID: func(context.Context, string) (*entity.Tree, error) {
				return &entity.Tree{ID: "tree-1", Title: "Goal title"}, nil
			},
		},
		nodeRepo: stubNodeRepository{
			getByTreeID: func(context.Context, string) ([]*entity.Node, error) {
				return []*entity.Node{root, child}, nil
			},
		},
		qaPairRepo: stubQAPairRepository{
			getByNodeID: func(context.Context, string) ([]*entity.QAPair, error) {
				return []*entity.QAPair{}, nil
			},
		},
		blockRepo: stubBlockRepository{},
		anchorRepo: stubAnchorRepository{
			getByID: func(context.Context, string) (*entity.Anchor, error) {
				return nil, errors.New("anchor unavailable")
			},
		},
		summaryProvider: newNoopSummaryProvider(),
	}

	payload, err := builder.BuildContext(context.Background(), "tree-1", "child", "How should anchor fallback behave?")
	if err != nil {
		t.Fatalf("BuildContext returned error: %v", err)
	}

	if !payload.Degraded {
		t.Fatalf("expected degraded payload")
	}
	found := false
	for _, warning := range payload.Warnings {
		if strings.Contains(warning, "anchor evidence degraded") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected anchor degradation warning, got %v", payload.Warnings)
	}
	assertContains(t, payload.UserPrompt, "## Anchor Evidence")
}

func TestBuildContextFailsWhenCurrentNodeMissingFromTree(t *testing.T) {
	root := &entity.Node{
		ID:       "root",
		TreeID:   "tree-1",
		Question: "Root question",
		Status:   entity.NodeStatusAnswered,
	}

	builder := &contextBuilder{
		treeRepo: stubTreeRepository{},
		nodeRepo: stubNodeRepository{
			getByTreeID: func(context.Context, string) ([]*entity.Node, error) {
				return []*entity.Node{root}, nil
			},
		},
		qaPairRepo:      stubQAPairRepository{},
		blockRepo:       stubBlockRepository{},
		anchorRepo:      stubAnchorRepository{},
		summaryProvider: newNoopSummaryProvider(),
	}

	_, err := builder.BuildContext(context.Background(), "tree-1", "missing", "Question")
	if err == nil {
		t.Fatalf("expected error when current node is missing")
	}
	if !strings.Contains(err.Error(), "current node missing") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestChatMockIncludesSystemPrompt(t *testing.T) {
	client := &OpenAIClient{}

	got, err := client.Chat(context.Background(), "tree goal context", "current question")
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}

	assertContains(t, got, "System Prompt")
	assertContains(t, got, "tree goal context")
	assertContains(t, got, "current question")
}

func assertContains(t *testing.T, got string, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("expected %q to contain %q", got, want)
	}
}

type stubTreeRepository struct {
	getByID func(ctx context.Context, id string) (*entity.Tree, error)
}

func (s stubTreeRepository) Create(context.Context, *entity.Tree) error { return nil }
func (s stubTreeRepository) GetByID(ctx context.Context, id string) (*entity.Tree, error) {
	if s.getByID != nil {
		return s.getByID(ctx, id)
	}
	return nil, errors.New("not implemented")
}
func (s stubTreeRepository) List(context.Context) ([]*entity.Tree, error) { return nil, nil }
func (s stubTreeRepository) Delete(context.Context, string) error         { return nil }

type stubNodeRepository struct {
	getByID     func(ctx context.Context, id string) (*entity.Node, error)
	getByTreeID func(ctx context.Context, treeID string) ([]*entity.Node, error)
}

func (s stubNodeRepository) Create(context.Context, *entity.Node) error { return nil }
func (s stubNodeRepository) GetByID(ctx context.Context, id string) (*entity.Node, error) {
	if s.getByID != nil {
		return s.getByID(ctx, id)
	}
	return nil, errors.New("not implemented")
}
func (s stubNodeRepository) GetByTreeID(ctx context.Context, treeID string) ([]*entity.Node, error) {
	if s.getByTreeID != nil {
		return s.getByTreeID(ctx, treeID)
	}
	return nil, errors.New("not implemented")
}
func (s stubNodeRepository) GetChildren(context.Context, string) ([]*entity.Node, error) {
	return nil, nil
}
func (s stubNodeRepository) Update(context.Context, *entity.Node) error { return nil }
func (s stubNodeRepository) Delete(context.Context, string) error       { return nil }
func (s stubNodeRepository) DeleteByTreeID(context.Context, string) error {
	return nil
}

type stubQAPairRepository struct {
	getByNodeID func(ctx context.Context, nodeID string) ([]*entity.QAPair, error)
}

func (s stubQAPairRepository) Create(context.Context, *entity.QAPair) error { return nil }
func (s stubQAPairRepository) GetByID(context.Context, string) (*entity.QAPair, error) {
	return nil, nil
}
func (s stubQAPairRepository) GetByNodeID(ctx context.Context, nodeID string) ([]*entity.QAPair, error) {
	if s.getByNodeID != nil {
		return s.getByNodeID(ctx, nodeID)
	}
	return nil, nil
}
func (s stubQAPairRepository) GetLatestByNodeID(context.Context, string) (*entity.QAPair, error) {
	return nil, nil
}
func (s stubQAPairRepository) DeleteByNodeID(context.Context, string) error { return nil }

type stubBlockRepository struct {
	getByQAPairID func(ctx context.Context, qaPairID string) ([]*entity.Block, error)
}

func (s stubBlockRepository) CreateBatch(context.Context, []*entity.Block) error { return nil }
func (s stubBlockRepository) GetByID(context.Context, string) (*entity.Block, error) {
	return nil, nil
}
func (s stubBlockRepository) GetByQAPairID(ctx context.Context, qaPairID string) ([]*entity.Block, error) {
	if s.getByQAPairID != nil {
		return s.getByQAPairID(ctx, qaPairID)
	}
	return nil, nil
}
func (s stubBlockRepository) DeleteByQAPairID(context.Context, string) error { return nil }

type stubAnchorRepository struct {
	getByID func(ctx context.Context, id string) (*entity.Anchor, error)
}

func (s stubAnchorRepository) Create(context.Context, *entity.Anchor) error { return nil }
func (s stubAnchorRepository) Update(context.Context, *entity.Anchor) error { return nil }
func (s stubAnchorRepository) GetByID(ctx context.Context, id string) (*entity.Anchor, error) {
	if s.getByID != nil {
		return s.getByID(ctx, id)
	}
	return nil, errors.New("not implemented")
}
func (s stubAnchorRepository) GetByNodeID(context.Context, string) ([]*entity.Anchor, error) {
	return nil, nil
}
func (s stubAnchorRepository) GetByBlockID(context.Context, string) ([]*entity.Anchor, error) {
	return nil, nil
}
func (s stubAnchorRepository) DeleteByNodeID(context.Context, string) error { return nil }
