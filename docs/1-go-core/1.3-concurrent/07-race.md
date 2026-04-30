---
title: "数据竞争检测"
module: "concurrent"
difficulty: "intermediate"
interviewFrequency: "medium"
tags:
  - race detector
  - 数据竞争
  - 并发安全
codeExample: "01-go-core/concurrent/goroutine/"
relatedEntries:
  - "/1-go-core/1.3-concurrent/03-sync"
  - "/1-go-core/1.3-concurrent/06-atomic"
prerequisites:
  - "/1-go-core/1.3-concurrent/01-goroutine"
estimatedTime: "25min"
---

# 数据竞争检测

## 概念说明

数据竞争（Data Race）是指两个或多个 goroutine 并发访问同一个变量，且至少有一个是写操作，并且没有使用同步机制。数据竞争是并发 bug 中最常见也最难排查的问题。

Go 内置了强大的 Race Detector（竞争检测器），通过 `-race` 标志即可在编译时注入检测代码，运行时自动检测数据竞争。

## 核心原理

### Race Detector 工作原理

Go 的 Race Detector 基于 Google 的 ThreadSanitizer（TSan）算法，在编译时对所有内存访问操作插桩，运行时记录每个 goroutine 对共享变量的访问历史，检测是否存在未同步的并发访问。

### 使用方式

```bash
# 编译并运行，启用竞争检测
go run -race main.go

# 测试时启用竞争检测
go test -race ./...

# 构建带竞争检测的二进制
go build -race -o myapp
```

### 检测输出示例

```
WARNING: DATA RACE
Write at 0x00c0000b4010 by goroutine 7:
  main.main.func1()
      /path/to/main.go:15 +0x38

Previous read at 0x00c0000b4010 by goroutine 6:
  main.main.func2()
      /path/to/main.go:20 +0x2c

Goroutine 7 (running) created at:
  main.main()
      /path/to/main.go:14 +0x98
```

### 性能开销

| 指标 | 开销 |
|------|------|
| 内存 | 5-10x |
| 执行速度 | 2-20x |
| 适用环境 | 开发和测试，不建议生产环境 |

## 标准库方案

Race Detector 是 Go 工具链的内置功能，无需额外依赖：

```go
// ❌ 数据竞争示例
var count int
go func() { count++ }()
go func() { count++ }()

// ✅ 修复方式 1：使用 Mutex
var mu sync.Mutex
go func() { mu.Lock(); count++; mu.Unlock() }()

// ✅ 修复方式 2：使用 atomic
var atomicCount atomic.Int64
go func() { atomicCount.Add(1) }()

// ✅ 修复方式 3：使用 channel
ch := make(chan int, 2)
go func() { ch <- 1 }()
go func() { ch <- 1 }()
count = <-ch + <-ch
```

## 代码示例

> 💻 完整可运行代码：[code-examples/01-go-core/concurrent/goroutine/](https://github.com/your-repo/code-examples/01-go-core/concurrent/goroutine/)
> 🏷️ Demo 模式：Part A（直接运行，使用 `go run -race main.go`）

## 常见面试题

### Q1: 什么是数据竞争？如何检测和修复？

**难度**：⭐⭐ | **频率**：🔥🔥

**答题思路**：

1. 定义数据竞争的三个条件
2. 介绍 `-race` 标志
3. 列举修复方法

**标准答案**：

数据竞争是指两个或多个 goroutine 并发访问同一变量，至少一个是写操作，且没有同步机制。Go 内置 Race Detector，通过 `go run -race` 或 `go test -race` 启用。修复方法：(1) 使用 `sync.Mutex` 保护临界区；(2) 使用 `sync/atomic` 进行原子操作；(3) 使用 channel 通信代替共享内存；(4) 使用 `sync.Map` 替代普通 map。

**深入追问**：
- Race Detector 能检测所有数据竞争吗？（不能，只能检测运行时实际发生的竞争，未执行的代码路径无法检测）
- 生产环境能开启 `-race` 吗？（不建议，内存和性能开销太大）

## 常见陷阱

1. **Race Detector 不是万能的**：只能检测运行时实际触发的竞争，代码覆盖率不足时可能遗漏
2. **map 并发读写**：Go 的 map 不是并发安全的，并发读写会直接 panic（不是数据竞争，是 runtime 检测）
3. **接口值的并发访问**：接口值由 type 和 data 两个字段组成，并发读写可能导致不一致
4. **slice 的并发 append**：多个 goroutine 同时 append 到同一个 slice 是数据竞争

## 参考资料

- [Go Blog - Introducing the Go Race Detector](https://go.dev/blog/race-detector)
- [Go Wiki - Data Race Detector](https://go.dev/doc/articles/race_detector)
- [ThreadSanitizer 算法](https://github.com/google/sanitizers)
