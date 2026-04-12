# Plan: summary 层

## 目标

把 summary 从“空占位”推进成“可被 context-builder 和后续界面消费的知识压缩层”。

## 分阶段

### Phase 1: 定义 summary 层边界

- 确认 summary 的粒度
- 确认 summary 的生成时机
- 确认 summary 的消费方

### Phase 2: 设计数据路径

- 定义 summary 的存储模型
- 定义 summary 的读取接口
- 定义缺失时的降级策略

### Phase 3: 接入上下文

- 把 summary 接入 context-builder
- 让非当前路径信息优先以摘要形式进入 prompt
- 保持当前聊天链路稳定

## 验收方向

- summary 不再只是空字段
- context-builder 能读取 summary
- summary 缺失时有明确 fallback

