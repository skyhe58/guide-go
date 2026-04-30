---
title: "Go 基础语法"
module: "go-basics"
difficulty: "beginner"
tags:
  - Go 基础
  - 入门必修
---

# Go 基础语法

> ⚠️ **本模块为所有后续模块的必修前置条件。** 请务必完成本模块全部内容后再进入其他模块学习。

## 模块概述

本模块系统讲解 Go 语言的基础语法和核心概念，是整个知识库的第一模块。Go 语言以简洁、高效、天生并发著称，其设计哲学"少即是多"（Less is More）贯穿语言的方方面面。

## 知识点索引

| 序号 | 知识点 | 难度 | 面试频率 | 预计时间 |
|------|--------|------|---------|---------|
| 01 | [环境搭建](./01-environment.md) | ⭐ | 🔥 | 20min |
| 02 | [数据类型](./02-data-types.md) | ⭐⭐ | 🔥🔥 | 30min |
| 03 | [变量、常量与作用域](./03-variables.md) | ⭐⭐ | 🔥🔥 | 30min |
| 04 | [运算符](./04-operators.md) | ⭐ | 🔥 | 20min |
| 05 | [控制流](./05-control-flow.md) | ⭐⭐ | 🔥🔥 | 30min |
| 06 | [函数](./06-functions.md) | ⭐⭐⭐ | 🔥🔥🔥 | 45min |
| 07 | [错误处理](./07-error-handling.md) | ⭐⭐⭐ | 🔥🔥🔥 | 40min |
| 08 | [结构体与方法](./08-struct-method.md) | ⭐⭐⭐ | 🔥🔥🔥 | 40min |
| 09 | [数组与切片](./09-slice.md) | ⭐⭐⭐ | 🔥🔥🔥 | 45min |
| 10 | [Map](./10-map.md) | ⭐⭐ | 🔥🔥🔥 | 30min |
| 11 | [指针](./11-pointer.md) | ⭐⭐ | 🔥🔥🔥 | 30min |
| 12 | [包管理与 Go Module](./12-module.md) | ⭐⭐ | 🔥🔥 | 30min |
| 13 | [字符串处理](./13-string.md) | ⭐⭐ | 🔥🔥 | 30min |
| 📝 | [面试指南](./interview.md) | - | 🔥🔥🔥 | 60min |

## 代码示例

> 💻 完整可运行代码：[code-examples/01-go-core/go-basics/](https://github.com/your-repo/code-examples/01-go-core/go-basics/)

| 示例目录 | 对应知识点 | 运行方式 |
|---------|-----------|---------|
| `datatypes/` | 数据类型、零值、类型转换 | `go run main.go` |
| `variables/` | 变量声明、iota、作用域 | `go run main.go` |
| `operators/` | 各类运算符 | `go run main.go` |
| `controlflow/` | if/for/switch | `go run main.go` |
| `functions/` | 多返回值/闭包/defer | `go run main.go` |
| `errorhandling/` | error 接口/panic-recover | `go run main.go` |
| `structs/` | 结构体/方法 | `go run main.go` |
| `slice/` | 切片操作/扩容 | `go run main.go` |
| `maps/` | Map 操作 | `go run main.go` |
| `pointers/` | 指针/new vs make | `go run main.go` |
| `modules/` | Go Module | `go run main.go` |
| `strings/` | 字符串/rune vs byte | `go run main.go` |
| `project-todo-cli/` | 综合练习：TODO CLI | `go run main.go` |

## 学习建议

1. **按顺序学习**：知识点之间有依赖关系，建议按序号顺序学习
2. **动手实践**：每个知识点都有对应的代码示例，务必亲手运行并修改
3. **关注面试**：标注 🔥🔥🔥 的知识点是面试高频考点，需重点掌握
4. **综合练习**：完成所有知识点后，尝试独立完成 `project-todo-cli` 综合项目

## 前置条件

- 已安装 Go 1.22+
- 基本的编程概念（变量、循环、函数等）
- 一个趁手的编辑器（推荐 VS Code + Go 插件 或 GoLand）
