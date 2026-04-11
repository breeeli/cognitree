# Project Memory

## 工作流

- **Feature**: brainstorm -> design-gate -> plan -> plan-gate -> implement -> code-review -> code-gate -> unit-test -> test-plan-gate -> e2e-test -> test-gate -> archive
- **Bugfix**: brainstorm -> reproduce-test -> test-gate -> fix -> unit-test-pass -> archive

每个阶段都有明确的产出文档，存放在 `requirements/<req-id>/` 目录下。

## 目录指针

- 需求管理: `requirements/INDEX.md`
- 知识库（业务/技术/经验）: `context/`
- 技术解决方案: `docs/solutions/`（由 ce:compound 管理）
- 工作流定义: `.cursor/pipelines/`
- Skill 定义: `.cursor/skills/`
- 命令入口: `.cursor/commands/`

## CE 插件集成

本项目使用 compound-engineering 插件提供基础能力：

- `ce:brainstorm` — 需求探索和文档化（产出到 `docs/brainstorms/`）
- `ce:plan` — 技术方案规划（产出到 `docs/plans/`）
- `ce:review` — 多 persona 代码审查
- `ce:compound` — 知识沉淀（产出到 `docs/solutions/`）
- `reproduce-bug` — Bug 复现

## 通用规范

- 中文沟通，技术术语保留英文
- 每个阶段必须产出对应文档到 `requirements/<req-id>/`
- 执行任何阶段前先调用 `experience-index` 检索 `context/` 和 `docs/solutions/`
- 项目特定技术栈和约束见 `context/tech/`
- 知识沉淀：技术问题用 `ce:compound`，业务/经验知识用 `knowledge-depositor`
