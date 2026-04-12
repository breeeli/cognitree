# Context Builder 优化沉淀

## 背景

Cognitree 的核心不是“聊天历史”，而是“围绕问题树逐步沉淀知识”。因此，`context-builder` 的目标也不该是简单拼接消息，而是把整棵树的结构、当前探索路径和关键证据组织成可控的模型输入。

这份沉淀记录的是 2026-04-12 之后的上下文构建思路，作为后续继续接入 summary、检索裁剪和更细粒度 anchor 语义的基础。

## 当前主线

现在的上下文组织主线可以概括为：

`tree goal` + `anchor evidence` + `current path` + `selection` + `current ask`

其中：

- `tree goal` 负责说明这棵树整体在解决什么问题
- `anchor evidence` 负责保留子问题从父回答中切出的依据
- `current path` 负责保留当前探索链路上的完整问答历史
- `selection` 负责把信息分成“必须进 prompt”和“只做结构提示”的两类
- `current ask` 作为本次新问题放在最后，避免过早抢占模型注意力

## 设计原则

### 1. 先保结构，再给问题

Prompt 不是线性聊天记录，而是树状知识结构的压缩视图。模型先看到树目标、树结构和证据，再看到本次提问，回答会更稳定。

### 2. 显式降级，不静默丢失

上下文构建失败时，不再悄悄跳过节点或证据。

当前策略是：

- 能保留节点标题，就保留标题
- 能保留占位说明，就保留占位说明
- 只有当前节点本身不可定位时，才让构建失败

这样做的好处是，模型虽然看到的是“降级后的上下文”，但仍然知道有哪些信息缺失。

### 3. summary 先占位，后接入

本期 summary 入口已经预留，但返回空。

这样做是为了先把上下文结构和调用链打通，等 summary 生成策略稳定后，再把节点摘要、子树摘要和全局摘要逐层接进去，而不是一次性把生成与消费同时改掉。

### 4. mock 路径也要验证系统提示词

本地无 API key 时，mock 输出不能只看 user prompt。

因为这次优化的关键变化之一是 system prompt 也承担了更多上下文约束，所以 mock 需要把 system prompt 一并暴露出来，避免“本地测通过、真实模型行为变化”这种假阳性。

## 现在的落点

当前实现已经把上下文拆成了几个明确的 section：

- `Tree Goal`
- `Tree Overview`
- `Current Path`
- `Anchor Evidence`
- `Summary`
- `Current Ask`

其中 `Summary` 目前是空占位，后续再按节点摘要和子树摘要逐步接入。

## 后续演进建议

- 先把 summary 接成可配置 provider，再决定摘要生成时机
- 再引入 token budget 和更细的 selection 规则
- 最后考虑把兄弟分支的高相关摘要也纳入 prompt，而不是只展示结构标题

