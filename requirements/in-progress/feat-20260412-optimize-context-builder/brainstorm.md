# Brainstorm: 优化 context-builder

## 背景

当前聊天链路中的 context 构建由后端 `backend/internal/infrastructure/ai/context_builder.go` 完成。现状已经沉淀在 `context/tech/current-context-construction.md` 中，可概括为：

> 全树结构概览 + 当前路径全量问答历史 + 当前问题

这个实现适合 MVP，但随着树变深、分支变多，开始与 Thinking Tree 的目标产生明显偏差。

## 当前主要问题

### 1. 全树收集过粗，缺少选择策略

当前实现先把整棵树全部节点加载出来，再展开当前路径详情。它没有：

- token budget
- relevance ranking
- priority selection
- fallback summary

结果是上下文增长方式基本线性，树稍大就会重新遇到 context noise 和 context window limit。

### 2. 整棵树“看见了结构”，但没“看见知识”

非当前路径分支只进入树概览，不进入内容层。模型知道有哪些节点，却不知道这些分支已经沉淀了什么答案和结论，跨分支复用能力很弱。

### 3. anchor 的语义价值没有进入 prompt

当前系统支持从回答中划词创建子问题，但 child node 聊天时不会携带：

- `anchor`
- `quoted_text`
- 引文位置

这导致“为什么从这里分叉”这条核心语义链在聊天时丢失。

### 4. tree 级目标没有进入 prompt

`Tree.Title` 和 `Tree.Description` 没有进入 `BuildContext`，模型只能看到节点问题，难以理解整棵树的长期目标和边界。

### 5. summary 机制尚未启用

仓库已有 `Summary` model，但当前 context-builder 没有接入 summary：

- 无节点摘要
- 无子树摘要
- 无远距离内容压缩

这意味着系统还停留在“提问历史堆叠”，尚未过渡到“知识结构沉淀”。

### 6. 上下文缺失不会显式暴露

当前拼接 thread 详情时，某些节点的 `qa_pairs` 或 `blocks` 查询失败会直接 `continue`。请求可能成功，但模型拿到的是不完整 context，接口层无法感知。

## 用户价值

优化完成后，用户应该获得这些直接收益：

1. 在树变大后仍能得到稳定、聚焦、上下文连续的回答。
2. 从父回答中分叉出来的子问题，能够保留明确的语义来源。
3. AI 不仅知道“现在问什么”，还知道“整棵树在探索什么”。
4. 已经探索过的分支知识能以摘要或相关信息的形式被复用，而不是只留在历史记录里。

## 需求目标

### 核心目标

1. 将当前 context-builder 从“全树收集器”升级为“结构化上下文构建器”。
2. 在 prompt 中补齐 tree goal、anchor evidence、summary 等高价值语义信息。
3. 为后续 token budget 和 context selection 打下可扩展的构建框架。

### 体验目标

1. 不改变用户当前的基本提问方式。
2. 不要求前端主动拼接复杂上下文，仍由后端负责构建。
3. 在功能增强的同时，保持聊天链路可观测、可调试。

## 范围

### In Scope

- 重构 `context_builder` 的输入、组装结构和输出格式
- 在构建中接入 tree 级信息
- 在构建中接入当前节点的 anchor 信息
- 设计并接入 summary 的读取位点
- 引入基础的上下文优先级策略
- 提升上下文构建过程的错误显式性和可观测性

### Out of Scope

- 前端聊天交互大改
- 向量检索或外部 RAG 系统
- 完整知识树抽象层
- 多模型路由
- 一次性做完最终形态的智能检索系统

## 建议的演进分期

### Phase 1: 补齐最关键语义

- 将 `Tree.Title/Description` 带入 context
- 若当前节点由 anchor 派生，则将 `quoted_text` 和来源信息带入 context
- 将当前“静默 continue”的数据缺失改为显式记录或失败策略

这一步优先解决“目标缺失”和“分叉依据缺失”。

### Phase 2: 引入摘要层

- 为节点和子树接入 summary 读取
- 让非当前路径分支从“只有标题”升级为“标题 + 摘要”
- 为远距离历史提供压缩替代

这一步优先解决“知识无法沉淀复用”。

### Phase 3: 引入选择策略

- 明确上下文优先级
- 根据 token budget 做裁剪
- 区分 full detail、summary、overview 三种信息密度

这一步优先解决“树变大后 prompt 失控”。

## 候选上下文结构

建议下一版 context 至少拆成这些块：

1. `tree goal`
2. `tree overview`
3. `current path detail`
4. `current node anchor evidence`
5. `sibling summaries`
6. `relevant remote summaries`
7. `current ask`

这样可以把“结构、证据、历史、当前提问”分层表达，而不是全部揉进一段文本里。

## 功能性要求

1. 系统必须在 context 中包含 tree 级目标信息。
2. 系统必须在当前节点由 anchor 派生时包含 anchor 引文信息。
3. 系统必须支持从 summary 层读取上下文，而不是只能依赖原始 QAPair/Block。
4. 系统必须定义明确的上下文优先级顺序。
5. 系统必须避免在上下文构建失败时静默丢失关键片段。

## 非功能性要求

1. 设计应保持后端 DDD 分层边界清晰。
2. context-builder 的能力扩展不应把复杂业务逻辑塞进 application 层。
3. 输出结构应便于后续继续演进到 token budget、检索和 summary 策略。
4. 调试时应能定位“本次 prompt 为何包含这些内容，忽略了哪些内容”。

## 验收标准

### 功能验收

1. 当前节点聊天时，prompt 中可看到 tree 级目标信息。
2. anchor 派生节点聊天时，prompt 中可看到 anchor 引文信息。
3. 非当前路径的重要分支不再只有标题，至少能以摘要形式进入上下文。
4. context-builder 具备明确的信息优先级顺序。

### 质量验收

1. 上下文构建失败不再被静默吞掉。
2. 代码结构比当前实现更易扩展，而不是继续向单个 builder 堆逻辑。
3. 为后续 summary/token budget 演进保留明确扩展点。

## 风险与开放问题

### 风险

1. 如果一步到位引入过多策略，可能把当前简单链路复杂化。
2. 如果 summary 生成机制尚未准备好，summary 接口设计可能先行于真实数据。
3. prompt 结构变复杂后，需要同步考虑调试与日志输出方式。

### 开放问题

1. 本次是否只优化 builder 的读取逻辑，还是同时落 summary 生成链路？
2. 非当前路径分支进入上下文时，优先使用“兄弟节点摘要”还是“相关节点摘要”？
3. 构建失败时更适合“显式降级”还是“直接失败返回”？

## 关联资料

- `context/tech/current-context-construction.md`
- `context/tech/mvp-architecture-decisions.md`
- `context/business/thinking-tree-domain.md`

## 建议的下一阶段

当前需求已完成 brainstorm，下一步进入 `design-gate`：

- 确认优化范围是否包含 summary
- 确认是否接受分阶段落地
- 确认以“tree goal + anchor + summary + selection”作为主线进入 plan
