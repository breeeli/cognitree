# MVP 前端开发踩坑经验

## 背景

Cognitree MVP 前端开发过程中，在文本选择交互、CORS 配置、状态持久化等方面遇到的问题和解决方案。

## 内容

### 文本选择 + 浮窗交互的竞态问题

**问题**：用户在 AI 回答中选中文本后弹出 QuotePopover，但在 popover 的 input 中打字或点击时，`mouseup` 事件触发 `useTextSelection` 的 handler，检测到当前 selection 不在 `[data-answer-content]` 区域内，于是 `setSelection(null)`，导致：
1. 浮窗消失（selection 被清空）
2. 即使浮窗还在，点击"创建子问题"时 `handleCreateAnchorChild` 读到 `selection === null`，直接 return

**解决**：三层防护机制：
- **lockSelection()**：selection 出现后立即锁定，后续 mouseup 不再修改 selection state
- **selectionRef**：用 `useRef` 保存 selection 快照，异步回调通过 ref 读取而非 state
- **clearSelection()**：同时解锁 + 清空，只在用户主动关闭浮窗或操作成功后调用
- **最小文本长度**：选中文本少于 2 字符时忽略，避免单击误触发
- **严格 blockId 校验**：blockId 为 falsy 时不设置 selection

**关键代码模式**：
```
// Hook 内部
const lockedRef = useRef(false);
const handleMouseUp = () => { if (lockedRef.current) return; ... };
const lockSelection = () => { lockedRef.current = true; };
const clearSelection = () => { lockedRef.current = false; setSelection(null); };

// 组件内部
const selectionRef = useRef(selection);
selectionRef.current = selection;
useEffect(() => { if (selection) lockSelection(); }, [selection]);
// 异步回调中用 selectionRef.current 而非 selection
```

### CORS + Vite 端口漂移

**问题**：Vite dev server 在端口被占用时自动递增（5173 → 5174 → 5175），但后端 CORS 白名单是硬编码的端口列表，导致新端口的请求返回 403 Forbidden。

**解决**：用 `AllowOriginFunc` 动态匹配，而非 `AllowOrigins` 静态列表：
```go
AllowOriginFunc: func(origin string) bool {
    return strings.HasPrefix(origin, "http://localhost:") ||
        strings.HasPrefix(origin, "http://127.0.0.1:")
},
```

**教训**：开发环境的 CORS 配置不要硬编码端口号，Vite/Webpack 等工具的端口是不确定的。

### 刷新丢失状态

**问题**：选择了思维树后刷新页面，状态全部丢失，需要重新选择。

**解决**：关键状态用 `localStorage` 持久化：
- `useState` 初始化时从 localStorage 读取
- 状态变更时同步写入 localStorage
- 页面加载时 useEffect 自动恢复

**注意**：只持久化 ID 类的轻量数据（如 `currentTreeId`），不要持久化完整的数据对象，启动时通过 ID 重新从 API 加载。

### 布局迭代经验

**问题**：初版布局 Toolbar 在顶部 + 左右分栏，与设计图（左侧三栏面板 + 右侧工作区）差异大。

**教训**：
- 前端布局重构时，新增组件文件后 Vite HMR 可能无法自动加载，需要手动刷新浏览器
- 工作区内容在宽屏上过于分散时，用 `max-w-3xl mx-auto` 居中约束
- 聊天气泡布局（AI 左 / 用户右）比传统 Q&A 上下排列更直观

## 适用场景

- 开发涉及文本选择 + 浮窗交互的功能时
- 配置前后端分离项目的 CORS 时
- 需要跨刷新持久化前端状态时
- 前端布局大幅重构时

## 关联

- `frontend/src/hooks/useTextSelection.ts` — 文本选择 Hook
- `frontend/src/components/workspace/QuotePopover.tsx` — 引用浮窗
- `frontend/src/components/workspace/WorkspacePanel.tsx` — 工作区（selection 集成）
- `backend/internal/interfaces/http/middleware/cors.go` — CORS 配置
- `context/experience/init-fullstack-lessons.md` — 初始化阶段踩坑经验
