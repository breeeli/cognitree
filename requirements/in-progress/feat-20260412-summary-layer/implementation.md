# Implementation: summary 层

## 当前状态

- 需求已拆分为独立 pipeline
- summary 的数据模型、仓储接口和读取 provider 已开始落地
- `context-builder` 已经能把 node/path/subtree 三种 summary 作为独立 section 收集
- summary 缺失时会进入 degraded warnings，而不是静默跳过

## 已完成

- 定义 summary 领域实体和仓储接口
- 定义 summary 数据模型和 GORM repository
- 接入 summary provider 到 `context-builder`
- 把 summary 纳入 `collect -> select -> format` 的结构
- 补充 summary 相关回归测试
- 实现 summary 异步 worker
- 实现摘要生成的立即重试
- 实现失败 summary 的周期性补偿重入队
- 将聊天链路接入 summary 派发

## 待办

- 为 summary 落库和失效策略补充更细的测试
- 继续优化 summary prompt 的生成质量
- 后续再考虑人工编辑入口和更细粒度失效策略
