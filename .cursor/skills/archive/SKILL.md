---
name: archive
description: 需求归档。检查产出物完整性，调用 ce:compound 沉淀技术知识，提取业务/经验知识到 context/，移动需求到 completed/，更新 INDEX.md。说"归档""archive""完成需求"时触发。
---

# Archive — 需求归档

需求完成后的归档流程，确保产出物齐全、知识被沉淀、状态被更新。

## 执行流程

### Step 1: 检查产出物完整性

读取 `requirements/<req-id>/README.md`，检查当前工作流类型（Feature/Bugfix）所需的产出物是否齐全。

**Feature 工作流必需产出物**：
- `brainstorm.md`
- `plan.md`
- `implementation.md`
- `code-review.md`
- `unit-test.md`
- `e2e-test.md`

**Bugfix 工作流必需产出物**：
- `brainstorm.md`
- `reproduce-test.md`
- `fix.md`
- `unit-test-pass.md`

如有缺失，列出缺失项并询问用户是否继续归档。

### Step 2: 生成归档总结

创建 `requirements/<req-id>/archive.md`：

```markdown
## 归档总结

### 基本信息
- 需求 ID: <req-id>
- 标题: <标题>
- 类型: Feature / Bugfix
- 创建时间: <时间>
- 完成时间: <时间>

### 各阶段产出物
- [x] brainstorm.md
- [x] plan.md
- ...

### 变更摘要
<从 implementation.md 或 fix.md 提取>

### 关键决策
<从各阶段 Gate 记录中提取>
```

### Step 3: 知识沉淀

#### 3a: 技术知识 -> docs/solutions/（通过 ce:compound）

如果需求涉及技术问题解决（Bug 修复、性能优化、架构决策等），调用 `ce:compound` 将解决方案沉淀到 `docs/solutions/`。

#### 3b: 业务/经验知识 -> context/

扫描各阶段产出物，提取值得沉淀的业务知识或经验：
- brainstorm.md 中的业务边界发现
- 实现过程中的踩坑经验
- 跨团队协作中的约束发现

如果发现值得沉淀的内容，调用 `knowledge-depositor` 写入 `context/`。

### Step 4: 移动需求目录

```bash
mv requirements/in-progress/<req-id> requirements/completed/<req-id>
```

### Step 5: 更新 INDEX.md

将需求从"进行中"移到"已完成"，更新 `requirements/INDEX.md`。

## 约束

- 归档前必须检查产出物完整性
- 知识沉淀是自动触发的，不需要用户额外操作
- 归档后需求目录不可修改（如需修改，创建新需求）
