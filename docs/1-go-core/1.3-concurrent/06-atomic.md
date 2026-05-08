---
title: "原子操作"
module: "concurrent"
difficulty: "intermediate"
interviewFrequency: "medium"
tags:
  - atomic
  - 原子操作
  - 无锁编程
codeExample: "01-go-core/concurrent/sync/"
relatedEntries:
  - "/1-go-core/1.3-concurrent/03-sync"
  - "/1-go-core/1.3-concurrent/07-race"
prerequisites:
  - "/1-go-core/1.3-concurrent/01-goroutine"
estimatedTime: "30min"
---

# 原子操作

## 概念说明

`sync/atomic` 包提供了底层的原子内存操作，用于实现无锁并发。原子操作直接映射到 CPU 指令（如 CAS），比互斥锁开销更小，适用于简单的计数器、标志位等场景。

原子操作解决的核心问题：**在不使用锁的情况下安全地读写共享变量**。

## 核心原理

### 基本原子操作

| 操作 | 函数 | 说明 |
|------|------|------|
| 加载 | `atomic.LoadInt64(&v)` | 原子读取 |
| 存储 | `atomic.StoreInt64(&v, new)` | 原子写入 |
| 加法 | `atomic.AddInt64(&v, delta)` | 原子加减 |
| 交换 | `atomic.SwapInt64(&v, new)` | 原子交换，返回旧值 |
| CAS | `atomic.CompareAndSwapInt64(&v, old, new)` | 比较并交换 |

### atomic.Value

`atomic.Value` 可以原子地存储和加载任意类型的值：

```go
var config atomic.Value

// 存储
config.Store(Config{Debug: true})

// 加载
cfg := config.Load().(Config)
```

适用于配置热更新等场景。

### Go 1.19+ 泛型原子类型

Go 1.19 引入了类型安全的原子类型：

```go
var counter atomic.Int64
counter.Add(1)
counter.Store(100)
val := counter.Load()

var flag atomic.Bool
flag.Store(true)
```

### Mutex vs Atomic 性能对比

| 场景 | Mutex | Atomic |
|------|-------|--------|
| 简单计数器 | ~30ns/op | ~5ns/op |
| 适用复杂度 | 任意复杂操作 | 单个变量的简单操作 |
| 可组合性 | 可保护多个变量 | 只能操作单个变量 |

## 标准库方案

```go
// Go 1.19+ 推荐方式
var counter atomic.Int64
counter.Add(1)
fmt.Println(counter.Load())

// 传统方式
var count int64
atomic.AddInt64(&count, 1)
fmt.Println(atomic.LoadInt64(&count))

// atomic.Value 存储任意类型
var val atomic.Value
val.Store("hello")
s := val.Load().(string)
```

## 代码示例

> 💻 完整可运行代码：[code-examples/01-go-core/concurrent/sync/](https://github.com/skyhe58/guide-go/tree/main/code-examples/01-go-core/concurrent/sync/)
> 🏷️ Demo 模式：Part A（直接运行）

## 常见面试题

### Q1: atomic 和 Mutex 如何选择？

**难度**：⭐⭐ | **频率**：🔥🔥

**答题思路**：

1. 从性能和适用场景两个维度对比
2. 强调 atomic 只适合简单操作

**标准答案**：

atomic 操作直接映射到 CPU 指令，性能比 Mutex 高 5-6 倍，但只能操作单个变量的简单操作（加减、读写、CAS）。Mutex 可以保护任意复杂的临界区（多个变量、复杂逻辑）。简单计数器、标志位用 atomic；涉及多个变量或复杂逻辑用 Mutex。

**深入追问**：
- CAS 操作的 ABA 问题是什么？（值从 A 变为 B 再变回 A，CAS 无法检测中间变化）
- atomic.Value 的 Store 有什么限制？（首次 Store 后，后续 Store 的值类型必须相同）

## 常见陷阱

1. **atomic 操作不可组合**：两个 atomic 操作之间不是原子的，需要多个变量的原子操作应使用 Mutex
2. **atomic.Value 类型一致性**：首次 Store 后，后续 Store 的值类型必须与首次相同，否则 panic
3. **误用 atomic 替代 Mutex**：复杂的读-改-写操作不能简单用 atomic 替代，可能导致数据不一致
4. **忽略内存对齐**：在 32 位系统上，64 位原子操作要求变量 8 字节对齐

## 参考资料

- [Go 标准库 sync/atomic 包](https://pkg.go.dev/sync/atomic)
- [Go Memory Model](https://go.dev/ref/mem)
