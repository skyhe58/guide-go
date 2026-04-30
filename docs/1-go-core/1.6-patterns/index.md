---
title: "设计模式与工程化"
module: "design-patterns"
difficulty: "intermediate"
tags:
  - 设计模式
  - 工程化
  - Functional Options
  - 中间件
  - Pipeline
  - SOLID
  - 项目布局
  - Makefile
  - Wire
---

# 设计模式与工程化

> **前置依赖：** 请先完成 [Go 基础语法](/1-go-core/1.1-go-basics/) 和 [Go 进阶特性](/1-go-core/1.2-go-advanced/) 模块。

## 模块概述

Go 语言的设计哲学——"少即是多"（Less is More）——深刻影响了 Go 社区对设计模式的理解和应用。与 Java 中经典的 GoF 23 种设计模式不同，Go 没有类继承、没有构造函数重载、没有注解，取而代之的是接口隐式实现、函数作为一等公民、组合优于继承的设计理念。

这意味着许多经典设计模式在 Go 中要么被简化（如工厂模式变成工厂函数），要么演化出 Go 特有的形态（如 Functional Options Pattern、Middleware Pattern），要么直接被语言特性取代（如迭代器模式被 for-range 取代）。

本模块分为两大部分：

1. **设计模式**：创建型、结构型、行为型经典模式的 Go 实现，以及 Go 特有的设计模式
2. **工程化实践**：项目布局、Makefile、Wire 依赖注入、Go 版本管理、错误处理规范

## 知识点索引

| 序号 | 知识点 | 难度 | 面试频率 | 预计时间 |
|------|--------|------|---------|---------|
| 01 | [创建型模式](./01-creational.md) | ⭐⭐ | 🔥🔥🔥 | 40min |
| 02 | [结构型模式](./02-structural.md) | ⭐⭐ | 🔥🔥 | 40min |
| 03 | [行为型模式](./03-behavioral.md) | ⭐⭐ | 🔥🔥 | 35min |
| 04 | [Go 特有模式](./04-go-patterns.md) | ⭐⭐⭐ | 🔥🔥🔥 | 45min |
| 05 | [设计原则](./05-principles.md) | ⭐⭐⭐ | 🔥🔥🔥 | 40min |
| 06 | [项目布局](./06-project-layout.md) | ⭐⭐ | 🔥🔥🔥 | 30min |
| 07 | [Makefile](./07-makefile.md) | ⭐⭐ | 🔥🔥 | 25min |
| 08 | [Wire 依赖注入](./08-wire.md) | ⭐⭐⭐ | 🔥🔥 | 35min |
| 09 | [Go 版本管理](./09-go-version.md) | ⭐ | 🔥 | 20min |
| 10 | [错误处理规范](./10-error-convention.md) | ⭐⭐ | 🔥🔥🔥 | 35min |

## 代码示例

> 💻 完整可运行代码：[code-examples/01-go-core/design-patterns/](https://github.com/)
> 🏷️ Demo 模式：Part A（直接运行）

| 示例目录 | 内容 | 运行方式 |
|---------|------|---------|
| `functional-options/` | Functional Options Pattern | `go run ./functional-options/` |
| `middleware/` | 中间件模式 | `go run ./middleware/` |
| `pipeline/` | Pipeline Pattern | `go run ./pipeline/` |
| `wire-example/` | Google Wire 依赖注入概念演示 | `go run ./wire-example/` |
| `project-layout/` | 标准项目布局示例 | `go run ./project-layout/` |

## 面试指南

📝 [设计模式与工程化面试指南](./interview.md) — 覆盖 Functional Options、SOLID 原则、项目布局等高频面试题。

## 参考资料

- [Go Proverbs](https://go-proverbs.github.io/) — Rob Pike 的 Go 格言
- [Effective Go](https://go.dev/doc/effective_go) — Go 官方最佳实践
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout) — Go 项目布局参考
- [Google Wire](https://github.com/google/wire) — 编译时依赖注入
