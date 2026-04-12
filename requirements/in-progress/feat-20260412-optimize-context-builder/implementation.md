# Implementation: 优化 context-builder

## 已完成内容

### 1. 重构 context-builder 为结构化流程

将原先集中在单文件中的“取数 + 选择 + 拼 prompt”逻辑重构为更清晰的流程：

- `collectTreeGoal`
- `collectTreeOverview`
- `collectThreadDetail`
- `collectAnchorEvidence`
- `collectSummarySections`
- `selectSections`
- `formatSystemPrompt`
- `formatUserPrompt`

这样后续接入真实 summary 和 token budget 时，不需要继续向单个函数堆逻辑。

### 2. 接入 tree goal

context 现在会尝试读取 `Tree.Title` 和 `Tree.Description`，并在 prompt 中增加 `Tree Goal` section。

如果 tree 元信息读取失败，会回退为仅使用 root question，同时记录 warning，属于显式降级。

### 3. 接入 anchor evidence

当当前节点存在 `AnchorID` 时，context 会读取对应 `anchor`，并在 prompt 中增加：

- 来源节点
- 引用文本
- offset 范围

这让 child node 在继续提问时能够保留“从哪段语义分叉出来”的证据链。

### 4. 接入 summary 占位 provider

新增 `summary_provider.go`，提供 `noopSummaryProvider`：

- `GetNodeSummary`
- `GetSiblingSummaries`
- `GetRelevantSummaries`

本期全部返回空，但保留了 summary 的结构位点和扩展入口。

### 5. 引入 selection 骨架

当前 context 不再直接按固定顺序拼字符串，而是先形成 section，再按优先级选择：

1. `Current Ask`
2. `Tree Goal`
3. `Anchor Evidence`
4. `Current Path`
5. `Tree Overview`
6. `Sibling Summaries`
7. `Related Summaries`

本期没有引入 token budget，但已经完成了 selection 这一步的结构化。

### 6. 引入显式降级

`ContextPayload` 新增：

- `Degraded bool`
- `Warnings []string`

builder 在遇到可恢复问题时不再静默吞掉，而是：

- 继续构建可用 context
- 将 warning 写入 payload
- 由 `ChatService` 记录 warning 日志

### 7. 更新依赖注入

`NewContextBuilder(...)` 现在注入：

- `treeRepo`
- `nodeRepo`
- `qaPairRepo`
- `blockRepo`
- `anchorRepo`

`main.go` 已同步更新。

## 变更文件

- `backend/internal/domain/service/context_builder.go`
- `backend/internal/infrastructure/ai/context_builder.go`
- `backend/internal/infrastructure/ai/summary_provider.go`
- `backend/internal/application/service/chat_service.go`
- `backend/cmd/server/main.go`
- `backend/internal/infrastructure/ai/context_builder_test.go`

## 验证

已执行：

```bash
cd backend && go test ./...
```

结果：

- 全部通过
- `internal/infrastructure/ai` 新增单元测试通过

## 当前状态

实现阶段已完成，下一步进入 `code-review`。
