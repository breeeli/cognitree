---
name: orchestrator-feature
description: Feature 工作流编排器。管理 12 阶段 Feature 开发流程的推进、状态跟踪和上下文传递。说"开始 feature""新需求""启动 feature 流程"时触发。
---

# Orchestrator Feature — Feature 工作流编排

管理 Feature 开发的 12 阶段流程：

```
brainstorm -> design-gate -> plan -> plan-gate -> implement -> code-review -> code-gate -> unit-test -> test-plan-gate -> e2e-test -> test-gate -> archive
```

## 启动

### 输入

用户提供需求描述（自然语言）。

### 初始化

1. 生成需求 ID：`feat-<YYYYMMDD>-<短描述>`
2. 创建需求目录：`requirements/in-progress/<req-id>/`
3. 创建需求 README：

```markdown
# <需求标题>

- ID: <req-id>
- 类型: Feature
- 创建时间: <YYYY-MM-DD HH:mm>
- 当前阶段: brainstorm
- 状态: in-progress

## 原始需求
<用户的需求描述>
```

4. 更新 `requirements/INDEX.md`，在"进行中"表格中添加记录
5. 创建状态文件 `requirements/in-progress/<req-id>/.status`：`brainstorm`

## 阶段执行

每个阶段遵循统一模式：

### 阶段前

1. 读取 `.status` 确认当前阶段
2. 调用 `experience-index` 检索相关知识
3. 读取前序阶段的产出物作为输入上下文

### 阶段执行

| 阶段 | 执行方式 | 产出物 |
|------|---------|--------|
| brainstorm | 调用 `ce:brainstorm`，传入需求描述 + 检索到的知识 | `brainstorm.md` |
| design-gate | 展示 brainstorm.md 摘要，等待用户确认/修改 | Gate 记录追加到 README |
| plan | 调用 `ce:plan`，传入 brainstorm.md + 检索到的知识 | `plan.md` |
| plan-gate | 展示 plan.md 摘要，等待用户确认/修改 | Gate 记录追加到 README |
| implement | 按 plan.md 的实现单元逐个编码 | `implementation.md` |
| code-review | 调用 `ce:review`，传入变更文件列表 | `code-review.md` |
| code-gate | 展示 review 结果，等待用户确认 | Gate 记录追加到 README |
| unit-test | 调用 `unit-test` skill（常规模式） | `unit-test.md` |
| test-plan-gate | 展示测试覆盖情况，等待用户确认 | Gate 记录追加到 README |
| e2e-test | 调用 `e2e-test` skill | `e2e-test.md` |
| test-gate | 展示 E2E 结果，等待用户确认 | Gate 记录追加到 README |
| archive | 调用 `archive` skill | `archive.md` |

### 阶段后

1. 更新 `.status` 为下一阶段
2. 更新 README 中的"当前阶段"
3. 通知 pipeline-server（如果运行中）

## Gate 处理

Gate 阶段是人工检查点：

1. 展示当前阶段的产出物摘要（关键信息，不超过 20 行）
2. 询问用户：
   - **通过** — 继续下一阶段
   - **修改** — 用户提出修改意见，回到前一阶段重新执行
   - **终止** — 标记需求为 cancelled

Gate 记录格式追加到 README：

```markdown
### Gate: <gate-name> — <YYYY-MM-DD HH:mm>
- 结果: 通过 / 修改 / 终止
- 备注: <用户的反馈>
```

## 阶段跳转

用户可以随时说"跳到 <阶段名>"来跳过中间阶段。编排器会：
1. 警告用户将跳过哪些阶段
2. 用户确认后直接跳转
3. 在 README 中记录跳过的阶段

## 恢复

如果会话中断，编排器可以通过读取 `.status` 文件恢复到上次的阶段继续执行。

## 约束

- 每次只推进一个阶段，不自动跳过 Gate
- Gate 必须等待用户明确响应
- 所有产出物写入 `requirements/in-progress/<req-id>/` 目录
- 产出物文件名固定，不可自定义
