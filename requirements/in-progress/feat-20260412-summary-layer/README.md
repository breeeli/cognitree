# 节点与子树 summary 层

- ID: `feat-20260412-summary-layer`
- 类型: `Feature`
- 创建时间: `2026-04-12 15:30`
- 当前阶段: `brainstorm`
- 状态: `in-progress`

## 原始需求

`/req-dev 接下来我们做summary，请你创建一个新的pipeline`

## 需求摘要

当前 `context-builder` 已经把 `summary` 作为占位层接进去了，但返回仍然是空。下一步要做的是把摘要从“预留字段”推进成“可消费的上下文层”，让系统能够把节点、路径和子树的关键信息压缩下来，而不是继续只依赖原始 QAPair 和 Block。

本 pipeline 的目标是先设计摘要层的职责、输入输出和接入时机，再决定摘要生成与存储的实现方式。重点不是一次把所有摘要都写完，而是先把“summary 应该解决什么问题”定义清楚。

## 当前产物

- `brainstorm.md`
- `plan.md`
- `implementation.md`

## 下一步建议

- 先确认 summary 的粒度：节点摘要、路径摘要、子树摘要，还是三者都要
- 再确认 summary 的生成时机：聊天后生成、后台异步生成，还是按需生成
- 最后确认 summary 的消费方式：仅用于 prompt，还是也要进入可视化和检索

