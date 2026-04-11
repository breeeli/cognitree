# 前后端初始化踩坑经验

## WSL2 端口问题
- Pipeline Server 绑定 `127.0.0.1` 时，Windows 浏览器无法访问
- 解决：改为绑定 `0.0.0.0`，通过 `localhost` 从 Windows 访问
- 如果 Windows 上有其他服务占用同端口（返回 401 Unauthorized），需换端口

## Go Proxy
- 内部镜像 `mirrors.wps.cn` 可能 TLS 超时
- 解决：使用 `GOPROXY=https://goproxy.cn,direct`

## Docker 镜像拉取
- 官方 registry `registry-1.docker.io` 可能超时
- 阿里云镜像 `registry.cn-hangzhou.aliyuncs.com/library/` 需要登录
- 解决：使用 `docker.1ms.run/library/` 镜像加速，拉取后 `docker tag` 为标准名称

## TypeScript 6.x 兼容性
- `baseUrl` 在 TS 7.0 将被移除，TS 6.x 会报弃用警告
- 解决：在 tsconfig 中添加 `"ignoreDeprecations": "6.0"`
- shadcn/ui 初始化需要根 `tsconfig.json` 中也有 `baseUrl` + `paths`

## Tailwind CSS v4
- 不再需要 `tailwind.config.js`，使用 CSS-first 配置
- shadcn/ui 会自动生成 OKLCH 色彩变量和 `@theme inline` 块
- 自定义颜色通过 `@theme { }` 块添加

## docker-compose 版本
- 旧版使用 `docker-compose`（连字符），新版使用 `docker compose`（空格）
- 检查方式：`docker compose version || docker-compose version`

## Bun 安装
- 官方安装脚本需要 `unzip`
- 替代方案：`npx --yes bun` 会自动下载到 npx 缓存
- 缓存位置：`~/.npm/_npx/*/node_modules/@oven/bun-linux-x64-baseline/bin/bun`
