---
name: orchestrator-bugfix
description: Bugfix 工作流编排器。管理 6 阶段 Bug 修复流程的推进、状态跟踪和上下文传递。说"修 bug""bugfix""修复问题"时触发。
---

# Orchestrator Bugfix — Bugfix 工作流编排

管理 Bug 修复的 6 阶段流程：

```
brainstorm -> reproduce-test -> test-gate -> fix -> unit-test-pass -> archive
```

## 启动

### 输入

用户提供 Bug 描述（现象、复现步骤、期望行为）。

### 初始化

1. 生成需求 ID：`fix-<YYYYMMDD>-<短描述>`
2. 创建需求目录：`requirements/in-progress/<req-id>/`
3. 创建需求 README：

```markdown
# <Bug 标题>

- ID: <req-id>
- 类型: Bugfix
- 创建时间: <YYYY-MM-DD HH:mm>
- 当前阶段: brainstorm
- 状态: in-progress

## Bug 描述
<用户的 Bug 描述>

## 复现步骤
<用户提供的复现步骤>

## 期望行为
<正确行为>

## 实际行为
<Bug 行为>
```

4. 更新 `requirements/INDEX.md`
5. 创建状态文件 `.status`：`brainstorm`

## 阶段执行

| 阶段 | 执行方式 | 产出物 |
|------|---------|--------|
| brainstorm | 调用 `ce:brainstorm`，聚焦 Bug 分析和根因假设 | `brainstorm.md` |
| reproduce-test | 调用 `unit-test` skill（复现模式），写一个应该 FAIL 的测试 | `reproduce-test.md` |
| test-gate | 展示复现测试结果，确认 Bug 被成功复现 | Gate 记录追加到 README |
| fix | 基于根因分析和复现测试，修复代码 | `fix.md` |
| unit-test-pass | 调用 `unit-test` skill（验证模式），确认复现测试 PASS | `unit-test-pass.md` |
| archive | 调用 `archive` skill | `archive.md` |

### 阶段前（统一）

1. 读取 `.status` 确认当前阶段
2. 调用 `experience-index` 检索相关知识（特别是 `docs/solutions/` 中的类似 Bug 解决方案）
3. 读取前序阶段的产出物

### 阶段后（统一）

1. 更新 `.status`
2. 更新 README 中的"当前阶段"

## Gate 处理

test-gate 是唯一的人工检查点：

1. 展示复现测试结果：
   - 测试文件路径
   - 测试是否 FAIL（FAIL = 成功复现）
   - 根因分析
2. 询问用户：
   - **通过** — Bug 已复现，继续修复
   - **修改** — 复现不准确，回到 reproduce-test
   - **终止** — 取消修复

## fix 阶段详情

fix 阶段的执行：

1. 读取 `brainstorm.md` 中的根因分析
2. 读取 `reproduce-test.md` 中的测试定位
3. 修改代码修复 Bug
4. 产出 `fix.md`：

```markdown
## 修复报告

### 根因
<确认的根因>

### 修复方案
<修复思路>

### 变更文件
- <文件路径>: <变更描述>

### 影响范围
<可能受影响的其他功能>
```

## 恢复

通过读取 `.status` 文件恢复到上次的阶段。

## 约束

- 复现测试必须在修复前 FAIL，修复后 PASS
- fix 阶段只修改必要的代码，不做无关重构
- 所有产出物写入 `requirements/in-progress/<req-id>/`
