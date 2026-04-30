---
title: "errgroup"
module: "concurrent"
difficulty: "intermediate"
interviewFrequency: "medium"
tags:
  - errgroup
  - 并发错误处理
  - golang.org/x/sync
codeExample: "01-go-core/concurrent/errgroup/"
relatedEntries:
  - "/1-go-core/1.3-concurrent/04-context"
  - "/1-go-core/1.3-concurrent/09-semaphore"
prerequisites:
  - "/1-go-core/1.3-concurrent/01-goroutine"
  - "/1-go-core/1.3-concurrent/04-context"
estimatedTime: "30min"
---

# errgroup

## 概念说明

`golang.org/x/sync/errgroup` 是 Go 官方扩展库中的并发错误处理工具，它在 `sync.WaitGroup` 的基础上增加了错误传播和 context 取消功能。当你需要并发执行多个任务并收集第一个错误时，errgroup 是最佳选择。

errgroup 解决的核心问题：**并发执行多个任务，任一任务失败时取消其余任务并返回错误**。

## 核心原理

### errgroup vs WaitGroup

| 特性 | sync.WaitGroup | errgroup.Group |
|------|---------------|----------------|
| 等待完成 | ✅ | ✅ |
| 错误收集 | ❌ 需手动处理 | ✅ 自动收集第一个错误 |
| context 取消 | ❌ 需手动实现 | ✅ `WithContext` 自动取消 |
| 并发限制 | ❌ | ✅ `SetLimit(n)` |
| 使用复杂度 | 需要 Add/Done/Wait | 只需 Go/Wait |

### 基本用法

```go
g, ctx := errgroup.WithContext(context.Background())

g.Go(func() error {
    // 任务 1
    return fetchURL(ctx, "https://example.com")
})

g.Go(func() error {
    // 任务 2
    return fetchURL(ctx, "https://example.org")
})

// 等待所有任务完成，返回第一个错误
if err := g.Wait(); err != nil {
    log.Fatal(err)
}
```

### 并发限制

```go
g := new(errgroup.Group)
g.SetLimit(3) // 最多 3 个 goroutine 并发执行

for _, url := range urls {
    url := url
    g.Go(func() error {
        return fetchURL(context.Background(), url)
    })
}
```

### TryGo

`TryGo` 尝试启动一个 goroutine，如果已达到并发限制则返回 false：

```go
g := new(errgroup.Group)
g.SetLimit(2)

ok := g.TryGo(func() error {
    // 如果并发数未达上限，执行此任务
    return nil
})
// ok == false 表示并发数已满
```

## 标准库方案

errgroup 属于 `golang.org/x/sync` 扩展库，需要额外安装：

```bash
go get golang.org/x/sync
```

## 第三方库方案

如果需要更丰富的功能（如收集所有错误、自定义错误处理），可以考虑：
- `github.com/hashicorp/go-multierror`：收集多个错误
- `github.com/sourcegraph/conc`：更现代的并发工具库

## 代码示例

> 💻 完整可运行代码：[code-examples/01-go-core/concurrent/errgroup/](https://github.com/your-repo/code-examples/01-go-core/concurrent/errgroup/)
> 🏷️ Demo 模式：Part A（直接运行）

## 常见面试题

### Q1: errgroup 和 WaitGroup 的区别？

**难度**：⭐⭐ | **频率**：🔥🔥

**标准答案**：

errgroup 在 WaitGroup 基础上增加了三个能力：(1) 自动收集第一个错误并通过 Wait() 返回；(2) 通过 WithContext 创建时，任一 goroutine 返回错误会自动取消 context，通知其他 goroutine 退出；(3) SetLimit 可以限制并发 goroutine 数量。WaitGroup 只提供等待功能，错误处理和取消需要手动实现。

**深入追问**：
- errgroup 只返回第一个错误，如何收集所有错误？（使用 channel 或 go-multierror 库）
- errgroup 的 SetLimit 底层如何实现？（使用 semaphore channel 控制并发数）

## 常见陷阱

1. **忽略 context 取消**：使用 `WithContext` 时，goroutine 内部应检查 `ctx.Done()` 以响应取消信号
2. **只返回第一个错误**：errgroup.Wait() 只返回第一个非 nil 错误，如果需要所有错误需要额外处理
3. **闭包变量捕获**：在循环中使用 `g.Go` 时注意闭包变量捕获问题（Go 1.22 已修复 for 循环变量）

## 参考资料

- [golang.org/x/sync/errgroup 文档](https://pkg.go.dev/golang.org/x/sync/errgroup)
- [Go Blog - Go Concurrency Patterns: Context](https://go.dev/blog/context)
