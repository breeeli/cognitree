# Cognitree

AI 工程工作流框架 — 通过结构化流程和知识沉淀，让每一次工程实践都比上一次更高效。

## 核心概念

**AI 工程工作流**: 将需求开发拆分为多个阶段，每个阶段有明确的产出物和质量门禁（Gate），由 AI 辅助执行、人工把关决策。

**Compound Engineering**: 集成 [compound-engineering](https://github.com/EveryInc/compound-engineering-plugin) 插件，复用其 brainstorm、plan、review、compound 能力，让工程知识持续积累而非每次从零开始。

**知识沉淀**: 每次需求完成后自动提取值得记录的知识（业务规则、技术决策、踩坑经验），写入知识库供后续需求自动检索。

## 目录结构

```
cognitree/
├── requirements/           # 需求管理
│   ├── INDEX.md            # 需求索引
│   ├── in-progress/        # 进行中的需求
│   └── completed/          # 已完成的需求
├── context/                # 项目知识库
│   ├── business/           # 业务领域知识
│   ├── tech/               # 技术背景
│   └── experience/         # 历史经验
├── docs/
│   ├── brainstorms/        # CE 插件 brainstorm 产出
│   ├── plans/              # CE 插件 plan 产出
│   └── solutions/          # CE 插件 compound 产出（技术解决方案）
└── .cursor/
    ├── AGENTS.md            # 项目记忆（工作流、目录指针、规范）
    ├── commands/            # 命令入口
    ├── skills/              # 自定义 Skill
    ├── plugins/             # CE 插件
    ├── pipelines/           # 工作流 YAML 定义
    ├── hooks/               # Hook 脚本 + Pipeline Server
    └── rules/               # 项目规则
```

## 快速开始

### 启动需求开发

在 Cursor Chat 中输入：

```
/req-dev <需求描述>
```

系统会自动判断需求类型（Feature / Bugfix）并启动对应的工作流编排器。

### 沉淀知识

在 Cursor Chat 中输入：

```
/optimize-flow [要沉淀的知识描述]
```

不带参数时，自动扫描当前需求的产出物提取知识。

## 工作流

### Feature（12 阶段）

```
brainstorm -> design-gate -> plan -> plan-gate -> implement -> code-review -> code-gate -> unit-test -> test-plan-gate -> e2e-test -> test-gate -> archive
```

| 阶段 | 说明 | 产出物 |
|------|------|--------|
| brainstorm | 需求探索和文档化 | brainstorm.md |
| design-gate | 人工确认需求方向 | Gate 记录 |
| plan | 技术方案规划 | plan.md |
| plan-gate | 人工确认技术方案 | Gate 记录 |
| implement | 代码实现 | implementation.md |
| code-review | 多 persona 代码审查 | code-review.md |
| code-gate | 人工确认审查结果 | Gate 记录 |
| unit-test | 单元测试 | unit-test.md |
| test-plan-gate | 人工确认测试覆盖 | Gate 记录 |
| e2e-test | 端到端测试 | e2e-test.md |
| test-gate | 人工确认测试结果 | Gate 记录 |
| archive | 归档 + 知识沉淀 | archive.md |

### Bugfix（6 阶段）

```
brainstorm -> reproduce-test -> test-gate -> fix -> unit-test-pass -> archive
```

| 阶段 | 说明 | 产出物 |
|------|------|--------|
| brainstorm | Bug 分析和根因假设 | brainstorm.md |
| reproduce-test | 写一个应该 FAIL 的测试复现 Bug | reproduce-test.md |
| test-gate | 人工确认 Bug 已复现 | Gate 记录 |
| fix | 修复代码 | fix.md |
| unit-test-pass | 验证复现测试 PASS | unit-test-pass.md |
| archive | 归档 + 知识沉淀 | archive.md |

## 技术栈

- **Cursor IDE** — AI 编程环境 + Hooks 生命周期管理
- **Compound Engineering Plugin** — brainstorm / plan / review / compound 核心能力
- **Bun + TypeScript** — Pipeline Server（流水线状态追踪和可视化看板）
- **Bash** — Hook 脚本（需求状态管理 + 上下文注入）
