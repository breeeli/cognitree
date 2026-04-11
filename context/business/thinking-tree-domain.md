# 思维树领域模型

## 背景

Cognitree 的核心问题域：人类知识结构是树状的，但大模型交互是线性的。这导致上下文漂移、上下文污染、上下文遗忘三个结构性问题。思维树 IDE 通过树状结构组织知识探索过程来解决这些问题。

## 内容

### 核心概念

**思维树（Thinking Tree）**：用户的知识探索以树状结构组织。不是在和 AI 聊天，而是在一棵问题树上探索、构建和沉淀知识。最终目标：提问树 → 知识树 → 个人知识体系。

**与 Chat 的根本区别**：
- 数据模型：消息流 → 结构化思考对象
- AI 输入：线性历史 → 可控的结构化上下文（整棵树）
- 交互结果：一次性回答 → 可持续沉淀的知识资产
- 系统目标：回答问题 → 帮助用户构建知识体系

### 五领域划分

| 领域 | 职责 | MVP 状态 |
|------|------|---------|
| **Node Domain** | Node CRUD、QAPair 追加、Block 存储、Anchor 创建、Summary 维护 | 已实现核心 |
| **Tree Domain** | 父子关联、展开/折叠、Thread 回溯、子树查询 | 已实现核心 |
| **Context Domain** | 收集候选上下文、按优先级筛选、生成结构化 Prompt | MVP 简化为全树收集 |
| **AI Domain** | 调用 LLM、解析为 Block、异步生成 Summary | 已实现调用，未做解析和异步 |
| **Knowledge Domain** | 从提问树抽象为知识树 | 未实现 |

### 核心对象模型

**Node** — 思考主题容器，不是一次问答
- 内部包含多轮 QAPair（多次提问和回答）
- 通过 Anchor 支持从文本局部展开子问题
- 状态：draft → answered → summarized

**QAPair** — 节点内一轮问答
- 属于 Node，按创建时间排序
- 包含多个 Block（回答的结构化文本块）

**Block** — 回答的结构化文本块
- 类型：paragraph / list / code / quote / heading
- MVP 阶段 Answer 存为单个 markdown Block，表结构已预留多 Block

**Anchor** — 回答中被选中的文本片段，子节点挂载点
- 核心差异化特性：「问题从哪里展开」具有明确语义
- 通过 start_offset / end_offset 定位在 Block 中的位置
- 一个 Anchor 对应一个 Child Node

**Thread** — 从当前节点到根的路径视图
- 运行时计算，非持久化实体
- 用于构建 AI 上下文中的「当前思考路径」

### 关键链路

```
用户提问
  → Application Service
  → Node Domain 持久化 Question
  → Context Domain 收集整棵树上下文
  → AI Domain 调用 LLM
  → 解析为 QAPair + Blocks
  → Node Domain 持久化回答
  → 前端展示，允许从任意文本展开子问题
```

### 两个关键设计点

1. **Answer 不是字符串，而是 Block 集合**：支持富文本渲染和文本级引用。即使 MVP 只存单个 Block，表结构已预留多 Block 能力。

2. **Child Node 挂载在 Anchor 上，而非简单的 parent_id**：传统树结构只有「A 是 B 的子节点」，思维树有「A 是 B 的哪段内容引出的子节点」。这让知识关联有了精确的语义锚点。

### 上下文构建策略

MVP 策略（全树收集）：
1. 加载整棵树所有 Node + QAPair
2. 构建树状结构文本：系统提示词 + 树结构概览 + 当前 Thread 完整 QA + 用户新问题
3. 调用 LLM，获取回答

后续演进方向：
- Collect → Select → Format 三步 + Token Budget
- 按优先级筛选：当前 Thread > 兄弟节点 > 远亲节点
- Summary 压缩：对远离当前节点的内容用摘要替代全文
- 向量检索：相似节点召回

## 适用场景

- 新功能设计时理解领域边界
- API 设计时确认对象关系
- 前端交互设计时理解数据流
- 后续迭代规划时参考演进路径

## 关联

- `backend/internal/domain/` — 领域层实现
- `backend/internal/application/service/` — 应用服务
- `requirements/in-progress/feat-20260409-mvp-thinking-tree/brainstorm.md` — 完整需求分析
- `context/tech/mvp-architecture-decisions.md` — 技术决策记录
