# 优化 context-builder

- ID: `feat-20260412-optimize-context-builder`
- 类型: `Feature`
- 创建时间: `2026-04-12 01:15`
- 当前阶段: `code-review`
- 状态: `in-progress`

## 原始需求

`/req-dev 创建需求，优化context-builder`

## 需求摘要

当前 `context-builder` 已经实现“全树结构概览 + 当前路径全量问答历史 + 当前问题”的基础上下文构建，但随着树规模增长，已经暴露出结构信息浅、上下文选择粗糙、缺少 anchor/tree goal/summary 语义、错误可观测性不足等问题。

本需求目标是在不破坏现有聊天链路的前提下，重构并增强 `context-builder`，让它更贴近 Thinking Tree 的长期目标。

## 当前产物

- `brainstorm.md`
- `plan.md`
- `implementation.md`

## Gate 记录

### Gate: design-gate - 2026-04-12 01:24

- 结果: 通过
- 备注: 优化范围本期先不做真实 summary 能力，按分阶段方式落地；context 构建失败采用显式降级；技术主线确认为 `tree goal + anchor + summary + selection`，其中 summary 本期先返回空占位。

### Gate: plan-gate - 2026-04-12 01:32

- 结果: 通过
- 备注: 按计划进入实现，当前已完成后端 builder 重构、依赖注入、降级日志和单元测试，下一阶段进入 `code-review`。
