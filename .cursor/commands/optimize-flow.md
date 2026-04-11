---
name: optimize-flow
description: 沉淀知识。从当前上下文中提取值得记录的知识，写入知识库。
---

# /optimize-flow — 知识沉淀

从当前上下文中提取值得记录的知识，写入对应的知识库。

## 使用方式

```
/optimize-flow [描述要沉淀的知识]
```

不带参数时，自动扫描当前需求的产出物提取知识。

## 执行逻辑

### 带参数

直接调用 `knowledge-depositor` skill，传入用户描述。

### 不带参数

1. 查找当前进行中的需求（`requirements/in-progress/` 下最新的目录）
2. 扫描各阶段产出物，提取值得沉淀的知识：
   - brainstorm.md 中的业务发现
   - 实现过程中的技术决策
   - 测试中发现的边界条件
   - 代码审查中的改进建议
3. 对于技术问题解决方案，调用 `ce:compound`
4. 对于业务/经验知识，调用 `knowledge-depositor`
5. 汇总沉淀结果
