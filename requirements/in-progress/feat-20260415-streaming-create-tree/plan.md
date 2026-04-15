# Plan: 流式创建思维树首问

## 目标

把“创建新树 + 首答生成”改造成一个基于 SSE 的流式编排流程，解决当前前端在后端响应回来后清空状态的问题，并让首问从一开始就以增量消息的方式呈现。

## 已确认约束

- 新树入口仍然是独立页面 `/trees/new`
- MVP 使用 SSE，不使用 WebSocket
- MVP 暂不支持取消生成
- MVP 暂不做流式结束后的自动刷新兜底
- 首答失败时，不应清空页面输入；已创建树资源应保留以便重试
- 当前代码里 `qapair` 和 `block` 是在 AI 成功返回后才落库，流式方案需要在不破坏这一安全边界的前提下引入分段输出

## 现有模式与接入点

- 新树入口已经在 `frontend/src/pages/NewTreePage.tsx`，并由 `frontend/src/App.tsx` 路由接入
- 侧边栏按钮已经跳到新树页，相关入口在 `frontend/src/components/sidebar/LeftPanel.tsx`
- 现有聊天链路是 `POST /nodes/:id/chat`，实现位于 `backend/internal/application/service/chat_service.go`
- `WorkspacePanel` 负责渲染消息与乐观态，位于 `frontend/src/components/workspace/WorkspacePanel.tsx`
- `AnswerView` 负责 markdown 渲染与消息布局，位于 `frontend/src/components/workspace/AnswerView.tsx`

## 方案

### 1. 后端新增流式编排入口

新增一个面向“首问创建”的 SSE 接口，由一个 orchestration 层负责串联现有服务：

- 创建 tree
- 创建 root node
- 开始 AI 生成
- 逐段向前端发送状态和增量内容
- 在 AI 成功完成后再落库 `qapair` 和 `block`

SSE 事件建议最少包含：

- `tree_created`
- `root_node_created`
- `answer_delta`
- `completed`
- `error`

如果首答生成失败，接口应只返回失败事件，不写入 `qapair` / `block`。这保持了当前“成功后才落库”的语义，并把部分输出仅保留在前端流式状态中。

### 2. 前端新树页改为流式状态机

`frontend/src/pages/NewTreePage.tsx` 不再只是“提交后跳回工作区”，而是成为首问流式展示页：

- 输入一个问题
- 提交后立即显示“创建中 / 生成中”
- 接收 SSE 增量并更新同一条回答内容
- 完成后再写入当前树 ID，并跳转回工作区

页面应保留以下状态：

- 当前问题文本
- 已创建树/根节点 ID
- 当前流式回答文本
- 进行中 / 完成 / 错误状态

### 3. 工作区保留现有一次性问答逻辑

这次只改“新树首问”链路，不把整个工作区聊天也改成 SSE。

原因：

- 最小化改动范围
- 先解决当前最明显的状态丢失问题
- 避免把已有 `WorkspacePanel` 的乐观消息、滚动、选择状态一并重构

### 4. 状态收口与数据一致性

新树页在流式结束后需要把最终状态收口到现有树模型：

- 成功时，把新树 ID 写入 `localStorage`
- 跳转回 `/`
- 工作区通过现有 `loadTree` 载入完整树详情

由于 MVP 不做流式结束后的自动刷新兜底，成功收口时的跳转应是唯一的“重新进入工作区”的动作。

## 任务拆分

### Task 1: 后端 SSE 编排层

实现一个新的首问流式入口，职责是：

- 串联 tree / root node 创建
- 调用 AI 流式输出
- 在完成后才持久化 `qapair` 和 `block`
- 将事件编码为 SSE

需要覆盖的文件方向：

- `backend/internal/interfaces/http/handler/*` 或新增专用 handler
- `backend/internal/interfaces/http/router/router.go`
- `backend/internal/application/service/*` 或新增 orchestration service
- `backend/internal/application/dto/*`
- `backend/internal/infrastructure/ai/*` 中与流式 AI 调用相关的实现

### Task 2: 前端新树流式页面

把 `frontend/src/pages/NewTreePage.tsx` 从“提交表单”升级为“流式生成页”：

- 建立 SSE 连接
- 逐段追加回答内容
- 展示创建/生成状态
- 成功后跳转到工作区

必要联动文件：

- `frontend/src/pages/NewTreePage.tsx`
- `frontend/src/App.tsx`
- `frontend/src/api/trees.ts` 或新增流式 API 封装

### Task 3: 流式展示组件

把流式回答状态从页面里抽出成可复用组件，避免页面变成一坨状态逻辑。

建议范围：

- 新增一个轻量的流式消息展示组件
- 复用现有 `AnswerView` 的 markdown 规范或部分样式

### Task 4: 测试与回归

补足新链路的最小验证：

- 后端：首答失败时不落库 `qapair`
- 后端：成功时按顺序返回流式事件
- 前端：流式结束前不清空展示内容
- 前端：成功后能正常跳回工作区并载入新树

## 文件级实现范围

### 后端

- `backend/internal/application/service/chat_service.go`
- `backend/internal/application/service/*` 新增 orchestration service
- `backend/internal/application/dto/*`
- `backend/internal/interfaces/http/handler/*`
- `backend/internal/interfaces/http/router/router.go`
- `backend/internal/infrastructure/ai/*`

### 前端

- `frontend/src/pages/NewTreePage.tsx`
- `frontend/src/App.tsx`
- `frontend/src/api/trees.ts`
- `frontend/src/components/workspace/AnswerView.tsx` 或新增流式展示组件
- `frontend/src/components/sidebar/LeftPanel.tsx` 仅保留路由跳转入口

## 测试场景

### 后端

- 首问提交成功时，SSE 事件顺序正确
- AI 失败时返回 error 事件且不写 `qapair`
- tree / root node 创建成功后，流式生成失败时不污染后续状态
- handler 遇到客户端断开时能安全结束，不把半成品当成完成结果

### 前端

- 新树页能在 SSE 增量到达时持续追加文本
- 流式过程中不会出现“响应回来后内容清空”
- 失败时输入内容仍保留，允许重新提交
- 成功后能写入树 ID 并回到工作区

## 风险与边界

- SSE 在浏览器和 Gin 间需要处理断线、重连和客户端中断语义，虽然 MVP 不做取消，但仍要优雅退出
- 首答内容在完成前不落库，意味着断线时可能丢失部分流式内容，这是 MVP 可接受的取舍
- 如果后端无法直接拿到 token 级流式输出，需退而求其次做 chunk 级增量输出，但前端展示模型保持不变

## 验收标准

- 新树首问在流式过程中始终保持可见
- 后端响应不再导致前端清空当前内容
- 成功路径能正确生成 tree / node / qapair / block
- 失败路径不会破坏已输入内容，也不会误写半成品数据
