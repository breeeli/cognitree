# 节点与子树 summary 层

- ID: `feat-20260412-summary-layer`
- 类型: `Feature`
- 创建时间: `2026-04-12 15:30`
- 当前阶段: `implementation-gate`
- 状态: `in-progress`

## 原始需求

`/req-dev 接下来我们做summary，请你创建一个新的pipeline`

## 需求摘要

当前 `context-builder` 已经把 `summary` 作为占位层接进去了，但返回仍然是空。下一步要做的是把摘要从“预留字段”推进成“可消费的上下文层”，让系统能够把节点、路径和子树的关键信息压缩下来，而不是继续只依赖原始 QAPair 和 Block。

本 pipeline 的目标是先设计摘要层的职责、输入输出和接入时机，再决定摘要生成与存储的实现方式。重点不是一次把所有摘要都写完，而是先把“summary 应该解决什么问题”定义清楚。

## 已确认方向

- summary 同时覆盖节点摘要、路径摘要和子树摘要
- 继续沿用 `collect -> select -> format` 的分层结构
- 先把 summary 注册进 `collect` 阶段，后续再一起验证效果
- summary 采用异步生成
- summary 先完全由模型生成，不做人工编辑入口
- summary 生成失败需要重试和补偿
- summary 缺失时可以完全忽略，但要记录一次可观测事件

## Gate 记录

### Gate: design-gate - 2026-04-12 15:30

- 结果: 通过
- 备注: summary 先覆盖节点摘要、路径摘要和子树摘要，按 `collect -> select -> format` 分层推进；异步生成、模型直出、失败重试补偿、缺失可观测事件作为前置约束。

### Gate: plan-gate - 2026-04-12 16:10

- 结果: 通过
- 备注: summary 已拆成数据模型、异步生成、collect 接入、format 接入和测试五个实现任务，开始进入代码落地阶段。

## 当前产物

- `brainstorm.md`
- `plan.md`
- `implementation.md`

## 下一步建议

- 先确认 summary 的粒度：节点摘要、路径摘要、子树摘要，还是三者都要
- 再确认 summary 的生成时机：聊天后生成、后台异步生成，还是按需生成
- 最后确认 summary 的消费方式：仅用于 prompt，还是也要进入可视化和检索
