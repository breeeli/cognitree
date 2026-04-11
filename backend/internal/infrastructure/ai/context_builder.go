package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/cognitree/backend/internal/domain/entity"
	"github.com/cognitree/backend/internal/domain/repository"
	domainservice "github.com/cognitree/backend/internal/domain/service"
)

type contextBuilder struct {
	nodeRepo   repository.NodeRepository
	qaPairRepo repository.QAPairRepository
	blockRepo  repository.BlockRepository
}

func NewContextBuilder(
	nodeRepo repository.NodeRepository,
	qaPairRepo repository.QAPairRepository,
	blockRepo repository.BlockRepository,
) domainservice.ContextBuilder {
	return &contextBuilder{
		nodeRepo:   nodeRepo,
		qaPairRepo: qaPairRepo,
		blockRepo:  blockRepo,
	}
}

func (b *contextBuilder) BuildContext(ctx context.Context, treeID string, currentNodeID string, newQuestion string) (*domainservice.ContextPayload, error) {
	nodes, err := b.nodeRepo.GetByTreeID(ctx, treeID)
	if err != nil {
		return nil, fmt.Errorf("get tree nodes: %w", err)
	}

	nodeMap := make(map[string]*entity.Node)
	childrenMap := make(map[string][]*entity.Node)
	var rootNode *entity.Node

	for _, n := range nodes {
		nodeMap[n.ID] = n
		if n.ParentNodeID == nil {
			rootNode = n
		} else {
			childrenMap[*n.ParentNodeID] = append(childrenMap[*n.ParentNodeID], n)
		}
	}

	if rootNode == nil {
		return nil, fmt.Errorf("root node not found")
	}

	thread := b.buildThread(nodeMap, currentNodeID)

	var treeOverview strings.Builder
	b.renderTreeOverview(&treeOverview, rootNode, childrenMap, 0, currentNodeID)

	var threadDetail strings.Builder
	for _, n := range thread {
		qaPairs, err := b.qaPairRepo.GetByNodeID(ctx, n.ID)
		if err != nil {
			continue
		}
		threadDetail.WriteString(fmt.Sprintf("\n### 节点: %s\n", n.Question))
		for _, qp := range qaPairs {
			threadDetail.WriteString(fmt.Sprintf("\n**问**: %s\n", qp.Question))
			blocks, err := b.blockRepo.GetByQAPairID(ctx, qp.ID)
			if err != nil {
				continue
			}
			for _, block := range blocks {
				threadDetail.WriteString(fmt.Sprintf("**答**: %s\n", block.Content))
			}
		}
	}

	systemPrompt := `你是一个知识探索助手，正在帮助用户在一棵"思维树"上深入探索问题。

思维树是一种树状知识结构，每个节点代表一个思考主题，节点内可以有多轮问答。用户通过在节点上提问来深入探索，也可以从回答中选取片段展开子问题。

你的回答应该：
1. 紧密围绕当前节点的主题
2. 考虑整棵树的上下文（用户的探索路径和已有知识）
3. 结构清晰，使用 Markdown 格式
4. 适当引导用户可以继续深入的方向`

	var userPrompt strings.Builder
	userPrompt.WriteString("## 思维树结构概览\n\n")
	userPrompt.WriteString(treeOverview.String())
	userPrompt.WriteString("\n\n## 当前探索路径（从根到当前节点）\n")
	userPrompt.WriteString(threadDetail.String())
	userPrompt.WriteString(fmt.Sprintf("\n\n## 当前问题\n\n%s", newQuestion))

	return &domainservice.ContextPayload{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt.String(),
	}, nil
}

func (b *contextBuilder) buildThread(nodeMap map[string]*entity.Node, currentNodeID string) []*entity.Node {
	var thread []*entity.Node
	id := currentNodeID
	for {
		node, ok := nodeMap[id]
		if !ok {
			break
		}
		thread = append([]*entity.Node{node}, thread...)
		if node.ParentNodeID == nil {
			break
		}
		id = *node.ParentNodeID
	}
	return thread
}

func (b *contextBuilder) renderTreeOverview(sb *strings.Builder, node *entity.Node, childrenMap map[string][]*entity.Node, depth int, currentID string) {
	indent := strings.Repeat("  ", depth)
	marker := ""
	if node.ID == currentID {
		marker = " ← 当前节点"
	}
	sb.WriteString(fmt.Sprintf("%s- %s [%s]%s\n", indent, node.Question, node.Status, marker))

	for _, child := range childrenMap[node.ID] {
		b.renderTreeOverview(sb, child, childrenMap, depth+1, currentID)
	}
}
