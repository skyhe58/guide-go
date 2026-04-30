---
title: "Go 进阶特性"
module: "go-advanced"
difficulty: "intermediate"
tags:
  - Go 进阶
  - 接口
  - 反射
  - 泛型
---

# Go 进阶特性

> **前置依赖：** 请先完成 [Go 基础语法](/1-go-core/1.1-go-basics/) 模块的全部内容。

## 模块概述

本模块深入讲解 Go 语言的高级特性，包括接口、组合与嵌入、反射、泛型、unsafe 包、代码生成和构建标签。这些特性是编写优雅、高效 Go 代码的关键，也是面试中的高频考点。

Go 的进阶特性设计充分体现了"组合优于继承"和"显式优于隐式"的哲学。接口的隐式实现让代码解耦更自然，泛型的引入（Go 1.18+）在保持简洁的同时提供了类型安全的抽象能力。

## 知识点索引

| 序号 | 知识点 | 难度 | 面试频率 | 预计时间 |
|------|--------|------|---------|---------|
| 01 | [接口](./01-interfaces.md) | ⭐⭐⭐ | 🔥🔥🔥 | 45min |
| 02 | [组合与嵌入](./02-composition.md) | ⭐⭐ | 🔥🔥🔥 | 30min |
| 03 | [反射](./03-reflection.md) | ⭐⭐⭐ | 🔥🔥 | 40min |
| 04 | [泛型](./04-generics.md) | ⭐⭐⭐ | 🔥🔥 | 40min |
| 05 | [unsafe 包](./05-unsafe.md) | ⭐⭐⭐ | 🔥🔥 | 35min |
| 06 | [代码生成](./06-codegen.md) | ⭐⭐ | 🔥 | 30min |
| 07 | [构建标签](./07-build-tags.md) | ⭐⭐ | 🔥 | 25min |
| 📝 | [面试指南](./interview.md) | - | 🔥🔥🔥 | 60min |

## 代码示例

> 💻 完整可运行代码：[code-examples/01-go-core/go-advanced/](https://github.com/your-repo/code-examples/01-go-core/go-advanced/)

| 示例目录 | 对应知识点 | 运行方式 |
|---------|-----------|---------|
| `interfaces/` | 接口实现/类型断言/标准库接口 | `go run main.go` |
| `reflection/` | 反射操作/标签解析 | `go run main.go` |
| `generics/` | 泛型函数/泛型类型 | `go run main.go` |
| `unsafe/` | unsafe 指针/内存布局 | `go run main.go` |
| `codegen/` | go generate 示例 | `go run main.go` |

## 学习建议

1. **接口是核心**：Go 的接口是最重要的抽象机制，务必深入理解隐式实现和接口组合
2. **反射慎用**：反射功能强大但性能开销大，理解原理后在实际项目中谨慎使用
3. **泛型适度**：Go 1.18 引入泛型，但不要滥用——能用接口解决的问题不必用泛型
4. **unsafe 了解即可**：除非做底层库开发或性能极致优化，日常开发中几乎不需要 unsafe

## 前置条件

- 已完成 [Go 基础语法](/1-go-core/1.1-go-basics/) 模块
- 熟悉结构体、方法、指针等基础概念
- Go 1.22+ 环境（泛型需要 Go 1.18+）
