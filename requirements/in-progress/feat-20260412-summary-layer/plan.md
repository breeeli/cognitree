# Plan: summary 层

## 目标

把 summary 从“空占位”推进成“可被 context-builder 和后续界面消费的知识压缩层”。

## 已确认约束

- summary 先覆盖节点摘要和子树摘要
- 继续沿用 `collect -> select -> format` 的分层结构
- 先把 summary 注册到 `collect`
- summary 采用异步生成
- summary 完全由模型生成
- summary 缺失时可以完全忽略，不做 prompt 占位

## 分阶段

### Phase 1: 定义 summary 数据和边界

- 确认 summary 的数据模型
- 明确节点摘要和子树摘要的职责
- 说明 summary 的消费边界

### Phase 2: 设计异步生成链路

- 定义 summary 异步任务的触发时机
- 定义失败重试和幂等策略
- 定义模型生成结果的落库路径

### Phase 3: 接入 collect

- 把 summary 注册进 `collect`
- 让 `select` 阶段可以读取摘要
- 保持缺失时完全忽略的策略

### Phase 4: 接入 format

- 把摘要按需格式化进 prompt
- 保持当前聊天链路稳定
- 为后续验证效果留下可观测输出

## 验收方向

- summary 不再只是空字段
- context-builder 能读取 summary
- summary 缺失时不会影响聊天链路
