# Plan: summary 层

## 目标

把 summary 从“空占位”推进成“可被 context-builder 和后续界面消费的知识压缩层”。

## 已确认约束

- summary 同时覆盖节点摘要、路径摘要和子树摘要
- 继续沿用 `collect -> select -> format` 的分层结构
- 先把 summary 注册到 `collect`
- summary 采用异步生成
- summary 完全由模型生成
- summary 生成失败需要重试和补偿
- summary 缺失时可以完全忽略，但要记录一次可观测事件

## 分阶段

### Phase 1: 定义 summary 数据和边界

- 确认 summary 的数据模型
- 明确节点摘要、路径摘要和子树摘要的职责
- 说明 summary 的消费边界和缺失可观测事件
- 定义 summary 元数据，至少包含 `scope`、`targetNodeID`、`treeID`、`version`、`status`
- 约定缺失时的日志事件名称和字段，确保可观测

### Phase 2: 设计异步生成链路

- 定义 summary 异步任务的触发时机
- 定义失败重试、补偿和幂等策略
- 定义模型生成结果的落库路径
- 先只使用模型生成，不开放人工编辑入口
- 为节点摘要、路径摘要、子树摘要分别定义生成入口，但允许共用 worker

### Phase 3: 接入 collect

- 把 summary 注册进 `collect`
- 让 `select` 阶段可以读取摘要
- 保持缺失时完全忽略但记录事件的策略
- 把 summary provider 从 noop 扩展为真实实现时，先按 scope 聚合返回数据
- `collect` 阶段先收集 summary 原始对象，不在这里做格式化

### Phase 4: 接入 format

- 把摘要按需格式化进 prompt
- 保持当前聊天链路稳定
- 为后续验证效果留下可观测输出
- `format` 阶段负责把节点摘要、路径摘要、子树摘要转成短文本 section
- 若某一类 summary 缺失，直接跳过，不插空占位

## 任务拆分

### Task 1: Summary model

- 补全 summary 领域/基础设施模型
- 定义三种 scope 的统一数据结构
- 确认缺失、过期、失败的状态表示

### Task 2: Summary generation worker

- 新增异步生成入口
- 接入重试和补偿逻辑
- 保证同一 summary 任务幂等

### Task 3: Summary storage and read API

- 建立 summary 持久化与查询接口
- 支持按节点、路径、子树读取
- 支持返回空结果但带可观测日志

### Task 4: ContextBuilder integration

- 在 `collect`、`select`、`format` 三个阶段接入 summary
- 统一处理缺失时的静默跳过与日志记录
- 保证现有聊天链路与降级行为不回退

### Task 5: Tests and observability

- 为三种 summary scope 补测试
- 为重试、补偿、缺失日志补测试
- 为 context-builder 的 prompt 组装补回归测试

## 验收方向

- summary 不再只是空字段
- context-builder 能读取 summary
- summary 缺失时不会影响聊天链路
- summary 缺失会留下可观测事件
- 三种 summary scope 都能在同一条 pipeline 里被验证
