# 项目技术栈

## 前端
- **框架**: React + TypeScript (TS 6.x)
- **构建**: Vite v8
- **样式**: Tailwind CSS v4 (CSS-first 配置，无 tailwind.config.js)
- **组件库**: shadcn/ui (OKLCH 色彩系统)
- **路由**: React Router v7
- **路径别名**: `@/` → `src/`（需要在 tsconfig.json 和 tsconfig.app.json 中都配置 `ignoreDeprecations: "6.0"` + `baseUrl` + `paths`）

## 后端
- **语言**: Go 1.26+
- **架构**: DDD 分层（domain / application / infrastructure / interfaces）
- **HTTP**: Gin v1.12
- **ORM**: GORM v1.31 + PostgreSQL driver
- **配置**: Viper (config.yaml + 环境变量)
- **日志**: Zap (结构化日志)
- **依赖方向**: interfaces → application → domain ← infrastructure

## 数据库
- PostgreSQL 16 (Docker Alpine)
- 连接信息: cognitree/cognitree@localhost:5432/cognitree

## 开发环境
- Docker Compose 启动 PostgreSQL
- Go proxy: `GOPROXY=https://goproxy.cn,direct`（内部镜像 mirrors.wps.cn 可能超时）
- Docker 镜像: 官方 registry 可能超时，使用 `docker.1ms.run` 镜像加速
- Pipeline Server: 端口 19091 (默认 19090 在 WSL2 环境可能被占用)

## 端口分配
| 服务 | 端口 |
|------|------|
| 前端 (Vite dev) | 5173 |
| 后端 (Gin) | 8085 |
| PostgreSQL | 5432 |
| Pipeline Server | 19091 |
