# MVP 架构决策记录

## 背景

Cognitree MVP 开发过程中做出的关键技术决策及其原因，供后续迭代参考。

## 内容

### 数据模型：完整设计 + 简化实现

**决策**：数据库表结构按完整设计建表（6 张表：trees/nodes/qa_pairs/blocks/anchors/summaries），但 MVP 只实现主链路。

**原因**：
- 表结构变更成本高（数据迁移），代码变更成本低
- Block 表预留了 type 字段（paragraph/list/code/quote/heading），MVP 阶段 Answer 存为单个 markdown Block
- Summary 表预留但 MVP 不写入
- 后续迭代可以在不改表结构的前提下逐步增强

### AI 上下文构建：全树收集

**决策**：MVP 阶段 ContextBuilder 收集整棵树所有 Node + QAPair，格式化为结构化 Prompt，不做 Select/Token Budget。

**原因**：MVP 数据量小，全树收集足够。后续需要加 Token Budget 和优先级筛选。

**风险**：树节点过多时会超 token 限制。缓解：后续加 Token Budget。

### 前端状态管理：useState + props

**决策**：不引入状态管理库（Redux/Zustand），用 `useState` + props 传递。

**原因**：MVP 规模下组件层级浅（TreePage → LeftPanel/WorkspacePanel），prop drilling 可控。关键状态（currentTreeId）用 localStorage 持久化。

**后续**：如果组件层级加深或跨组件通信增多，考虑引入 Zustand。

### 前端布局：两列固定宽度

**决策**：左侧面板固定 `w-72`（288px），右侧工作区 `flex-1`，内容区 `max-w-3xl` 居中。

**原因**：
- 问题树需要固定宽度保证可读性
- 工作区内容居中避免宽屏下阅读困难
- 参考了 ChatGPT / Notion 的布局模式

### Anchor 机制：字符偏移量

**决策**：Anchor 使用 `start_offset` / `end_offset` 基于 Block content 的字符偏移量定位。

**风险**：Markdown 渲染后 DOM 结构与原始文本的偏移量可能不一致。MVP 阶段用 `textContent` 计算偏移量，基本可用。

**后续**：如果偏移量不准确，考虑改为基于 DOM 路径的定位方式。

## 适用场景

- 规划后续迭代时参考已有决策
- 评估技术债务和重构优先级
- 新成员了解项目技术选型背景

## 关联

- `requirements/in-progress/feat-20260409-mvp-thinking-tree/plan.md` — 详细实现计划
- `requirements/in-progress/feat-20260409-mvp-thinking-tree/brainstorm.md` — 需求分析
- `context/tech/project-stack.md` — 技术栈信息
