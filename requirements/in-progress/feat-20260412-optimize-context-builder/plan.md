# Plan: 优化 context-builder

## 设计前提

本计划基于 design-gate 已确认的约束：

1. 本期不实现真实 summary 生成与持久化链路。
2. 采用分阶段落地，避免一次性把聊天链路复杂化。
3. context 构建失败采用显式降级，而不是静默丢失。
4. 技术主线按 `tree goal + anchor + summary + selection` 组织。
5. summary 在本期保留结构位点，但返回空结果。

## 本期目标

本期目标不是一次做完最终形态的智能上下文系统，而是把当前的 `context_builder` 从“单文件拼 prompt”重构为“可扩展的结构化构建流程”，并优先补齐两个最关键的语义缺口：

- tree goal
- anchor evidence

同时，为后续 summary 和 selection 演进建立稳定骨架。

## 非目标

本期不做以下内容：

- summary 生成逻辑
- summary 持久化与管理界面
- 向量检索
- 外部知识库/RAG
- 前端聊天交互改版
- 多模型调度

## 目标架构

### 当前问题

当前 `backend/internal/infrastructure/ai/context_builder.go` 负责：

- 加载数据
- 组织优先级
- 拼接结构
- 格式化 prompt
- 静默吞掉部分错误

这些职责混在一起，导致后续引入 tree goal、anchor、summary、selection 时会继续向单个文件堆逻辑。

### 本期目标结构

建议把 context 构建拆成四步：

1. `collect`
2. `select`
3. `format`
4. `report degradation`

其中：

- `collect` 负责取数与组装原始候选上下文
- `select` 负责按优先级决定哪些内容进入 prompt
- `format` 负责将结构化上下文转成 prompt
- `report degradation` 负责显式记录本次上下文缺失了什么

## 分阶段落地方案

### Phase 1: 当前实现范围

本期实际交付：

1. 引入 tree goal
2. 引入 anchor evidence
3. 引入 selection 骨架
4. 引入 summary 占位接口，默认返回空
5. 引入显式降级机制

### Phase 2: 后续扩展点

后续再补：

1. summary 的真实 repository / provider
2. 兄弟分支和远支节点的摘要进入上下文
3. token budget
4. 更细粒度的 relevance selection

## 计划中的上下文结构

本期统一按以下结构组装候选上下文：

1. `tree goal`
2. `tree overview`
3. `current path detail`
4. `anchor evidence`
5. `sibling summaries`
6. `relevant remote summaries`
7. `current ask`

其中本期实际落地说明：

- `tree goal`: 实现
- `tree overview`: 保留
- `current path detail`: 保留
- `anchor evidence`: 实现
- `sibling summaries`: 占位，当前为空
- `relevant remote summaries`: 占位，当前为空
- `current ask`: 保留

## 选择策略

### 本期选择优先级

本期先实现基础选择顺序，而不是复杂预算模型。优先级建议如下：

1. `current ask`
2. `tree goal`
3. `anchor evidence`
4. `current path detail`
5. `tree overview`
6. `sibling summaries`
7. `relevant remote summaries`

说明：

- `current ask`、`tree goal` 为强制进入
- `anchor evidence` 在当前节点存在 `AnchorID` 时进入
- `current path detail` 保留为当前主要语义密度来源
- `tree overview` 保留整棵树结构感
- `summary` 相关 section 先允许为空，但保留结构与扩展位点

### 为什么这样排

当前目标不是压缩到极致，而是先把“全局目标”和“分叉依据”补齐。当前路径仍然是最主要的知识来源，因此优先级仍高于其他分支信息。

## 显式降级策略

### 基本原则

恢复性失败不再静默 `continue` 后无痕结束，而是要做到：

1. builder 内部记录降级原因
2. payload 对外暴露降级信息
3. ChatService 记录日志
4. prompt 内可选加入简短的上下文说明

### 故障分类

#### 致命失败

以下情况直接返回 error：

- 当前节点不存在
- tree 不存在
- tree 节点集合无法加载
- 当前路径无法构建

这些场景意味着本次聊天已无法可靠构建基本上下文。

#### 可恢复降级

以下情况进入显式降级：

- tree 元信息加载失败
- anchor 读取失败
- 某个节点的 qa_pairs 读取失败
- 某个 qa_pair 的 blocks 读取失败
- summary provider 不可用或返回空

这些场景不阻断主流程，但必须记录 warning。

### 表达方式

建议扩展 `ContextPayload`，新增：

- `Warnings []string`
- `Degraded bool`

这样 `ChatService` 可以：

1. 在日志中输出本次降级原因
2. 在需要时将调试信息暴露给开发环境
3. 保持 AIClient 仍只消费 `systemPrompt/userPrompt`

## summary 占位方案

### 本期方案

本期不直接引入完整 `SummaryRepository` 落地链路，而是先在 `infrastructure/ai` 内抽出一个最小 `summaryProvider` 抽象，默认实现返回空：

- `GetNodeSummary(nodeID) -> empty`
- `GetSiblingSummaries(nodeID) -> empty`
- `GetRelevantSummaries(nodeID, question) -> empty`

这能满足两件事：

1. prompt 结构可以提前稳定
2. 后续引入真实 summary 时，不必再重写 builder 主流程

### 为什么不现在做真实 summary

用户已明确本期范围先不做 summary，本期最重要的是先完成 builder 架构重构和关键语义补齐，而不是把 scope 扩到摘要生产链路。

## 代码层改动计划

### 1. 扩展依赖注入

计划改动：

- `backend/cmd/server/main.go`

变更内容：

- `NewContextBuilder(...)` 增加 `treeRepo`
- `NewContextBuilder(...)` 增加 `anchorRepo`
- 注入一个默认空实现的 `summaryProvider`

### 2. 扩展 ContextBuilder 输出

计划改动：

- `backend/internal/domain/service/context_builder.go`

变更内容：

- 为 `ContextPayload` 增加 `Warnings`
- 为 `ContextPayload` 增加 `Degraded`

保持对 `ChatService` 的主调用方式不变。

### 3. 重构 ai/context_builder.go

计划改动：

- `backend/internal/infrastructure/ai/context_builder.go`

建议拆分职责：

- `collectTreeGoal`
- `collectThread`
- `collectAnchorEvidence`
- `collectSummarySections`
- `selectSections`
- `formatSystemPrompt`
- `formatUserPrompt`
- `appendWarning`

这样能让 tree goal、anchor、summary、selection 各自成为独立扩展点。

### 4. 引入 summary provider 占位实现

计划新增文件：

- `backend/internal/infrastructure/ai/summary_provider.go`

职责：

- 定义最小接口
- 提供 `noopSummaryProvider`
- 为未来真实 summary 接入预留统一入口

### 5. 在 ChatService 中消费降级信息

计划改动：

- `backend/internal/application/service/chat_service.go`

变更内容：

- 调用 `BuildContext` 后检查 `payload.Degraded`
- 使用现有 logger 记录 warning
- 不改变对 AIClient 的主调用方式

### 6. 保持前端零改动

本期不改：

- `frontend/src/api/chat.ts`
- `frontend/src/components/workspace/WorkspacePanel.tsx`

上下文仍由后端负责构建。

## Prompt 结构计划

### System Prompt

保持“Thinking Tree 助手”的角色设定，但语言上补充：

- 需要优先理解树的整体目标
- 在 anchor 存在时，应将其视为当前问题的语义证据

### User Prompt

建议改为分块式结构：

1. `## Tree Goal`
2. `## Tree Overview`
3. `## Current Path`
4. `## Anchor Evidence`
5. `## Related Summaries`
6. `## Current Ask`

其中：

- `Anchor Evidence` 在无 anchor 时省略
- `Related Summaries` 本期允许为空，或在开发态写明 `暂无摘要`
- 若发生降级，可加一个简短的 `## Context Build Notes`

## 测试计划

### 单元测试

建议为 builder 补这些测试：

1. 能把 tree goal 正确拼入 prompt
2. 当前节点存在 `AnchorID` 时能拼入 `quoted_text`
3. 无 anchor 时不会错误注入 anchor section
4. summary provider 返回空时 prompt 仍可正常生成
5. 可恢复失败时 `Degraded=true` 且 `Warnings` 非空
6. 致命失败时返回 error

### 集成验证

建议验证以下场景：

1. 根节点继续提问
2. 普通子节点继续提问
3. anchor 派生子节点继续提问
4. 某个可选 section 读取失败时仍能返回回答

## 风险与应对

### 风险 1: builder 重构后 prompt 行为波动

应对：

- 保留当前路径详情作为主内容
- 通过测试锁住 tree goal / anchor 的新增行为

### 风险 2: summary 占位过早抽象

应对：

- 接口保持最小
- 只在 ai builder 内部使用
- 不提前扩散到前端与应用层

### 风险 3: 降级信息过多影响 prompt 质量

应对：

- warning 默认用于日志
- prompt 中只保留简短说明
- 开发态和生产态可区分输出强度

## 实施顺序

1. 扩展 `ContextPayload`
2. 扩展 `NewContextBuilder` 依赖
3. 为 summary 增加空 provider
4. 重构 builder 为 collect/select/format 流程
5. 接入 tree goal
6. 接入 anchor evidence
7. 接入显式降级与 warning
8. 补测试

## 产出判断

完成本计划后，应达到以下状态：

1. context-builder 代码结构明显比当前更清晰
2. tree goal 与 anchor 已进入 prompt
3. summary 已有稳定占位扩展点，但当前不产生真实内容
4. selection 已成为独立步骤，而不是散落在字符串拼接中
5. 可恢复失败会显式降级，不再无痕丢失上下文

## 下一步

当前已完成 `plan`，进入 `plan-gate`。待确认后再进入实现阶段。
