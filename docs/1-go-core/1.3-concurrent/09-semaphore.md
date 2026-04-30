---
title: "semaphore"
module: "concurrent"
difficulty: "intermediate"
interviewFrequency: "low"
tags:
  - semaphore
  - 信号量
  - 并发控制
  - golang.org/x/sync
codeExample: "01-go-core/concurrent/errgroup/"
relatedEntries:
  - "/1-go-core/1.3-concurrent/08-errgroup"
  - "/1-go-core/1.3-concurrent/03-sync"
prerequisites:
  - "/1-go-core/1.3-concurrent/01-goroutine"
  - "/1-go-core/1.3-concurrent/04-context"
estimatedTime: "25min"
---

# semaphore

## 概念说明

`golang.org/x/sync/semaphore` 提供了加权信号量（Weighted Semaphore）的实现，用于控制对共享资源的并发访问数量。与简单的 channel 信号量不同，加权信号量支持每次获取不同数量的资源，适用于更复杂的并发控制场景。

信号量解决的核心问题：**限制同时访问某个资源的 goroutine 数量**。

## 核心原理

### 信号量 vs 其他并发控制

| 方式 | 适用场景 | 灵活性 |
|------|---------|--------|
| `make(chan struct{}, n)` | 简单的固定并发限制 | 低 |
| `errgroup.SetLimit(n)` | errgroup 内的并发限制 | 中 |
| `semaphore.Weighted` | 加权并发控制、支持 context | 高 |

### 基本 API

```go
sem := semaphore.NewWeighted(int64(maxConcurrency))

// 获取资源（阻塞直到有足够资源）
if err := sem.Acquire(ctx, 1); err != nil {
    return err // context 被取消
}
defer sem.Release(1)

// 尝试获取（非阻塞）
if sem.TryAcquire(1) {
    defer sem.Release(1)
    // 获取成功
}
```

### 加权信号量的应用

```go
// 场景：数据库连接池，最多 10 个连接
// 普通查询占 1 个连接，批量导入占 3 个连接
sem := semaphore.NewWeighted(10)

// 普通查询
sem.Acquire(ctx, 1)
defer sem.Release(1)

// 批量导入（占用更多资源）
sem.Acquire(ctx, 3)
defer sem.Release(3)
```

## 标准库方案

信号量属于 `golang.org/x/sync` 扩展库：

```bash
go get golang.org/x/sync
```

简单场景可以用 buffered channel 实现信号量：

```go
// 简单信号量：最多 3 个并发
sem := make(chan struct{}, 3)

sem <- struct{}{} // 获取
// 执行操作...
<-sem             // 释放
```

## 代码示例

> 💻 完整可运行代码：[code-examples/01-go-core/concurrent/errgroup/](https://github.com/your-repo/code-examples/01-go-core/concurrent/errgroup/)
> 🏷️ Demo 模式：Part A（直接运行）

## 常见面试题

### Q1: 如何在 Go 中实现信号量？

**难度**：⭐⭐ | **频率**：🔥

**标准答案**：

Go 没有内置信号量，但有多种实现方式：(1) 使用 buffered channel，容量即为信号量大小；(2) 使用 `golang.org/x/sync/semaphore` 包的加权信号量，支持 context 取消和不同权重的资源获取；(3) 使用 `errgroup.SetLimit` 在 errgroup 内限制并发。简单场景用 channel，复杂场景用 semaphore 包。

## 常见陷阱

1. **Acquire 和 Release 不配对**：忘记 Release 会导致信号量永远被占用，使用 `defer sem.Release(n)`
2. **忽略 Acquire 的错误**：当 context 被取消时 Acquire 返回错误，必须检查
3. **Release 超过 Acquire**：Release 的总量超过 Acquire 的总量会 panic

## 参考资料

- [golang.org/x/sync/semaphore 文档](https://pkg.go.dev/golang.org/x/sync/semaphore)
- [Go Wiki - Semaphore](https://go.dev/wiki/Semaphore)
