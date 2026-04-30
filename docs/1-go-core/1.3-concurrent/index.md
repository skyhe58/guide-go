---
title: "并发编程"
module: "concurrent"
difficulty: "intermediate"
tags:
  - 并发编程
  - goroutine
  - channel
  - sync
  - context
---

# 并发编程

> **前置依赖：** 请先完成 [Go 基础语法](/1-go-core/1.1-go-basics/) 模块的全部内容。

## 模块概述

并发编程是 Go 语言的**核心竞争力**。Go 从语言层面内置了轻量级并发原语——goroutine 和 channel，配合 `sync`、`context`、`atomic` 等标准库包，让开发者能够轻松编写高并发程序。

Go 的并发哲学源自 CSP（Communicating Sequential Processes）模型：

> **"Don't communicate by sharing memory; share memory by communicating."**
> ——不要通过共享内存来通信，而要通过通信来共享内存。

这意味着 Go 推荐使用 channel 在 goroutine 之间传递数据，而非使用锁来保护共享变量。当然，在性能敏感的场景下，`sync.Mutex` 和 `atomic` 操作仍然是必要的工具。

## 知识点索引

| 序号 | 知识点 | 难度 | 面试频率 | 预计时间 |
|------|--------|------|---------|---------|
| 01 | [goroutine](./01-goroutine.md) | ⭐⭐ | 🔥🔥🔥 | 40min |
| 02 | [channel](./02-channel.md) | ⭐⭐⭐ | 🔥🔥🔥 | 50min |
| 03 | [sync 包](./03-sync.md) | ⭐⭐⭐ | 🔥🔥🔥 | 45min |
| 04 | [context 包](./04-context.md) | ⭐⭐⭐ | 🔥🔥🔥 | 40min |
| 05 | [并发模式](./05-patterns.md) | ⭐⭐⭐ | 🔥🔥 | 60min |
| 06 | [原子操作](./06-atomic.md) | ⭐⭐ | 🔥🔥 | 30min |
| 07 | [数据竞争检测](./07-race.md) | ⭐⭐ | 🔥🔥 | 25min |
| 08 | [errgroup](./08-errgroup.md) | ⭐⭐ | 🔥🔥 | 30min |
| 09 | [semaphore](./09-semaphore.md) | ⭐⭐ | 🔥 | 25min |
| 📝 | [面试指南](./interview.md) | - | 🔥🔥🔥 | 60min |

## 代码示例

> 💻 完整可运行代码：[code-examples/01-go-core/concurrent/](https://github.com/your-repo/code-examples/01-go-core/concurrent/)

| 示例目录 | 对应知识点 | 运行方式 |
|---------|-----------|---------|
| `goroutine/` | goroutine 创建/泄漏检测 | `go run main.go` |
| `channel/` | channel 各种用法/select | `go run main.go` |
| `sync/` | Mutex/WaitGroup/Once/Pool | `go run main.go` |
| `context/` | context 传播/超时控制 | `go run main.go` |
| `patterns/` | fan-in/fan-out/pipeline/worker-pool | `go run main.go` |
| `errgroup/` | errgroup 并发错误处理 | `go run main.go` |

## 学习建议

1. **goroutine 是基础**：理解 goroutine 的创建、调度和生命周期管理是一切并发编程的前提
2. **channel 是核心**：掌握 channel 的各种用法和关闭语义，这是 Go 并发的灵魂
3. **sync 包是补充**：当 channel 不适合时（如保护共享状态），使用 sync 包中的同步原语
4. **context 是规范**：所有涉及超时、取消的并发操作都应使用 context，这是 Go 社区的标准实践
5. **并发模式要熟练**：fan-in/fan-out/pipeline/worker-pool 是面试和实战中的高频模式
6. **race detector 常开**：开发和测试阶段始终使用 `-race` 标志检测数据竞争

## 前置条件

- 已完成 [Go 基础语法](/1-go-core/1.1-go-basics/) 模块
- 熟悉函数、闭包、错误处理等基础概念
- Go 1.22+ 环境
