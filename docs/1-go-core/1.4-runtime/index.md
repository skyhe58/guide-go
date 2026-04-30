---
title: "运行时与性能"
module: "runtime"
difficulty: "advanced"
tags:
  - 运行时
  - GMP
  - GC
  - 内存管理
  - pprof
  - benchmark
  - 性能优化
---

# 运行时与性能

> **前置依赖：** 请先完成 [Go 基础语法](/1-go-core/1.1-go-basics/) 和 [并发编程](/1-go-core/1.3-concurrent/) 模块。

## 模块概述

Go 运行时（runtime）是 Go 程序的"隐形引擎"，负责 goroutine 调度、内存分配、垃圾回收、栈管理等核心工作。理解运行时原理是编写高性能 Go 代码和排查线上问题的基础。

本模块分为两大部分：

1. **运行时原理**：GMP 调度模型、垃圾回收、内存管理、栈管理——理解 Go 程序"为什么快"
2. **性能分析与优化**：pprof、trace、benchmark、逃逸分析——掌握"如何让代码更快"

## 知识点索引

| 序号 | 知识点 | 难度 | 面试频率 | 预计时间 |
|------|--------|------|---------|---------|
| 01 | [GMP 调度模型](./01-gmp.md) | ⭐⭐⭐ | 🔥🔥🔥 | 60min |
| 02 | [垃圾回收](./02-gc.md) | ⭐⭐⭐ | 🔥🔥🔥 | 50min |
| 03 | [内存管理](./03-memory.md) | ⭐⭐⭐ | 🔥🔥 | 45min |
| 04 | [栈管理](./04-stack.md) | ⭐⭐⭐ | 🔥🔥 | 30min |
| 05 | [pprof 性能分析](./05-pprof.md) | ⭐⭐ | 🔥🔥🔥 | 45min |
| 06 | [trace 工具](./06-trace.md) | ⭐⭐⭐ | 🔥🔥 | 35min |
| 07 | [benchmark](./07-benchmark.md) | ⭐⭐ | 🔥🔥 | 35min |
| 08 | [常见优化技巧](./08-optimization.md) | ⭐⭐ | 🔥🔥🔥 | 40min |
| 09 | [逃逸分析实战](./09-escape.md) | ⭐⭐⭐ | 🔥🔥🔥 | 35min |
| 📝 | [面试指南](./interview.md) | - | 🔥🔥🔥 | 60min |

## 代码示例

> 💻 完整可运行代码：[code-examples/01-go-core/runtime/](https://github.com/your-repo/code-examples/01-go-core/runtime/)

| 示例目录 | 对应知识点 | 运行方式 |
|---------|-----------|---------|
| `gmp/` | GMP 调度模拟（Part A 纯内存模拟） | `go run main.go` |
| `gc/` | GC 三色标记模拟 | `go run main.go` |
| `pprof/` | pprof 分析示例 | `go run main.go` |
| `benchmark/` | benchmark 示例 | `go test -bench=. -benchmem` |

## 学习建议

1. **GMP 是面试必考**：理解 G/M/P 三者关系、调度策略、抢占式调度是面试的重中之重
2. **GC 原理要清晰**：三色标记法、写屏障、GC 触发条件是高频考点
3. **pprof 是实战利器**：线上排查 CPU 飙高、内存泄漏、goroutine 泄漏都离不开 pprof
4. **benchmark 要会写**：性能优化必须有数据支撑，benchmark 是 Go 内置的性能测试工具
5. **逃逸分析要理解**：减少堆分配是 Go 性能优化的核心手段之一

## 前置条件

- 已完成 [Go 基础语法](/1-go-core/1.1-go-basics/) 模块
- 已完成 [并发编程](/1-go-core/1.3-concurrent/) 模块
- 理解 goroutine、channel 等并发基础概念
- Go 1.22+ 环境
