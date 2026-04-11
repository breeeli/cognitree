---
name: unit-test
description: 单元测试。支持三种模式：常规（feature 流程的单测）、复现（bugfix 流程写一个应该 FAIL 的测试）、验证（bugfix 流程确认修复后测试 PASS）。说"写单测""跑测试""单元测试"时触发。
---

# Unit Test — 单元测试

通用的单元测试 skill，支持三种模式适配不同工作流阶段。

## 模式

### 常规模式（默认）

用于 Feature 工作流的 unit-test 阶段。

**流程**：
1. 读取 `requirements/<req-id>/implementation.md` 了解代码变更
2. 分析变更涉及的模块和函数
3. 编写单元测试，覆盖：
   - 正常路径（happy path）
   - 边界条件（boundary cases）
   - 错误处理（error paths）
4. 运行测试并收集结果
5. 产出测试报告

**产出**：`requirements/<req-id>/unit-test.md`

### 复现模式（mode: reproduce）

用于 Bugfix 工作流的 reproduce-test 阶段。目标是写一个**应该 FAIL** 的测试来复现 Bug。

**流程**：
1. 读取 `requirements/<req-id>/brainstorm.md` 了解 Bug 描述和根因假设
2. 定位 Bug 涉及的代码路径
3. 编写一个精确复现 Bug 的测试用例：
   - 测试应该在当前代码上 **FAIL**
   - 测试描述清楚说明期望行为和实际行为
4. 运行测试，确认它确实 FAIL
5. 如果测试 PASS（说明 Bug 未被复现），分析原因并调整

**产出**：`requirements/<req-id>/reproduce-test.md`

```markdown
## Bug 复现测试报告

### 测试文件
- <测试文件路径>

### 测试用例
- <用例名称>: FAIL（符合预期，成功复现 Bug）

### Bug 复现确认
- 输入: <触发 Bug 的输入>
- 期望行为: <正确行为>
- 实际行为: <Bug 行为>

### 根因定位
- <基于测试结果的根因分析>
```

### 验证模式（mode: verify）

用于 Bugfix 工作流的 unit-test-pass 阶段。目标是确认修复后测试 **PASS**。

**流程**：
1. 读取 `requirements/<req-id>/reproduce-test.md` 找到复现测试
2. 运行复现测试，确认它现在 PASS
3. 如果仍然 FAIL，报告修复不完整
4. 额外运行相关的回归测试，确保修复没有引入新问题

**产出**：`requirements/<req-id>/unit-test-pass.md`

```markdown
## 修复验证报告

### 复现测试
- <测试名称>: PASS（Bug 已修复）

### 回归测试
- <相关测试>: PASS
- <相关测试>: PASS

### 结论
Bug 已修复，回归测试通过。
```

## 通用约束

- 测试框架由项目决定，读取项目配置（package.json、go.mod、Makefile 等）确定
- 测试文件放在项目约定的测试目录下
- 不修改业务代码，只写测试代码
- 测试命名清晰，描述被测试的行为
