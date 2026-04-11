---
name: req-dev
description: 启动需求开发。根据需求类型（Feature/Bugfix）启动对应的工作流编排器。
---

# /req-dev — 启动需求开发

根据用户输入判断需求类型，启动对应的工作流编排器。

## 使用方式

```
/req-dev <需求描述>
```

## 执行逻辑

### Step 1: 判断需求类型

根据用户描述判断：

- 包含"bug""修复""fix""报错""异常""不正常"等关键词 -> **Bugfix**
- 其他 -> **Feature**

如果不确定，询问用户。

### Step 2: 启动编排器

- Feature -> 读取并执行 `orchestrator-feature` skill
- Bugfix -> 读取并执行 `orchestrator-bugfix` skill

### Step 3: 传递上下文

将用户的需求描述完整传递给编排器。
