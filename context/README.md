# 项目知识库

按类型组织的可复用知识，由 `experience-index` skill 自动检索，由 `knowledge-depositor` skill 和 `ce:compound` 自动沉淀。

## 目录结构

- `business/` — 业务领域知识（业务边界、流程、规则）
- `tech/` — 技术背景（架构、配置规范、服务依赖）
- `experience/` — 历史经验（踩坑记录、解决方案、最佳实践）

## 与 docs/solutions/ 的关系

- `docs/solutions/` 由 CE 插件的 `ce:compound` 管理，存放结构化的技术问题解决方案
- `context/` 存放更广泛的知识：业务领域知识、技术背景、非技术经验等
- `experience-index` skill 同时搜索两个目录
