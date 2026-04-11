# Cognitree — 思维树 IDE（Thinking Tree IDE）

## 项目愿景

用户不是在和 AI 聊天，而是在一棵"问题树 / 思维树"上探索、构建和沉淀知识。最终目标：**提问树 → 知识树 → 个人知识体系**。

## 问题域

当前主流 Chat 模式为"短期问答"设计，而非"知识构建"。根本矛盾：

> **人类的知识结构是树状的（Hierarchical），而大模型的交互方式是线性的（Linear Conversation）。**

这导致三个结构性问题：

1. **上下文漂移（Context Drift）** — 思考路径树状发散，但对话线性推进，讨论不断偏离主问题
2. **上下文污染（Context Noise）** — 回到主问题时，上下文已充满弱相关内容，占据 token，降低理解质量
3. **上下文遗忘（Context Window Limit）** — 对话过长时，早期重要信息被截断，模型表现"遗忘"

Chat 模式的本质局限：知识无法结构化沉淀、上下文非结构化、AI 不理解长期目标、历史对话难以复用、难以形成知识体系。

## 核心概念

**思维树（Thinking Tree）**：用户的知识探索以树状结构组织，每个节点是一个问题或主题，子节点是对父节点的深入探索。

**AI 的角色**：
- 帮助用户沿问题树逐层深入
- 基于"整棵问题树"构造上下文（而非线性历史对话）
- 对已有节点进行总结、归纳和结构化整理
- 将"提问过程"逐渐沉淀为"知识结构"

## 技术栈

| 层 | 技术 | 说明 |
|----|------|------|
| 前端 | React + TypeScript + Vite + Tailwind CSS + shadcn/ui | SPA，树状可视化交互 |
| 后端 | Go + Gin + DDD 分层架构 | 领域驱动设计 |
| 数据库 | PostgreSQL + GORM | JSONB 存储灵活数据结构 |
| 基础设施 | Docker Compose / Viper / Zap / Swagger | 开发环境 + 配置 + 日志 + API 文档 |

## 目录结构

```
cognitree/
├── frontend/               # 前端 — React SPA
├── backend/                # 后端 — Go DDD
├── docker-compose.yml      # 开发环境编排
├── requirements/           # 需求管理（工作流产出物）
├── context/                # 项目知识库
├── docs/                   # 文档（brainstorm / plan / solution）
└── .cursor/                # Cursor 工作流配置
```

### 后端 DDD 分层

```
backend/
├── cmd/server/             # 应用入口
├── internal/
│   ├── domain/             # 领域层 — 实体、值对象、仓储接口、领域服务
│   ├── application/        # 应用层 — 用例编排（command / query / dto）
│   ├── infrastructure/     # 基础设施层 — 仓储实现、配置、中间件
│   └── interfaces/         # 接口层 — HTTP handler / router
├── pkg/                    # 可复用公共包
├── go.mod
└── go.sum
```

**依赖方向**：interfaces → application → domain ← infrastructure（依赖反转）

## 编码规范

### Go 后端
- 遵循 Go 官方代码风格（gofmt / goimports）
- domain 层零外部依赖，只定义接口
- infrastructure 实现 domain 定义的接口
- application 层编排领域对象，不包含业务逻辑
- 错误处理使用自定义错误类型，不裸用 `error`

### 前端
- 组件使用函数式组件 + Hooks
- 样式使用 Tailwind CSS，不写自定义 CSS（除非必要）
- API 调用统一通过 `api/` 层，不在组件中直接 fetch
- TypeScript strict 模式
