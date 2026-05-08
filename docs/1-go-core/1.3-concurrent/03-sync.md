---
title: "sync 包"
module: "concurrent"
difficulty: "intermediate"
interviewFrequency: "high"
tags:
  - sync
  - Mutex
  - WaitGroup
  - Once
  - Pool
  - 面试高频
codeExample: "01-go-core/concurrent/sync/"
relatedEntries:
  - "/1-go-core/1.3-concurrent/06-atomic"
  - "/1-go-core/1.3-concurrent/01-goroutine"
prerequisites:
  - "/1-go-core/1.3-concurrent/01-goroutine"
estimatedTime: "45min"
---

# sync 包

## 概念说明

`sync` 包提供了 Go 标准库中的基本同步原语，用于在多个 goroutine 之间协调对共享资源的访问。虽然 Go 推荐使用 channel 进行通信，但在保护共享状态、等待一组操作完成等场景下，`sync` 包中的工具更加直接高效。

sync 包解决的核心问题：**多 goroutine 对共享资源的安全访问和协调**。

## 核心原理

### Mutex（互斥锁）

`sync.Mutex` 是最基本的互斥锁，同一时刻只允许一个 goroutine 持有锁：

```go
var mu sync.Mutex
var count int

mu.Lock()
count++
mu.Unlock()
```

**注意**：Mutex 不可重入（同一 goroutine 重复 Lock 会死锁），也不可复制。

### RWMutex（读写锁）

`sync.RWMutex` 允许多个读操作并发执行，但写操作是排他的：

```go
var rw sync.RWMutex

// 读操作 —— 多个 goroutine 可同时持有读锁
rw.RLock()
val := sharedData
rw.RUnlock()

// 写操作 —— 排他，阻塞所有读写
rw.Lock()
sharedData = newVal
rw.Unlock()
```

适用于**读多写少**的场景，如配置缓存、路由表等。

### WaitGroup（等待组）

`sync.WaitGroup` 用于等待一组 goroutine 完成：

```go
var wg sync.WaitGroup
for i := 0; i < 5; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        // 执行任务...
    }(i)
}
wg.Wait() // 阻塞直到所有 goroutine 调用 Done()
```

### Once（单次执行）

`sync.Once` 确保某个操作只执行一次，常用于单例初始化：

```go
var once sync.Once
var instance *Config

func GetConfig() *Config {
    once.Do(func() {
        instance = loadConfig()
    })
    return instance
}
```

### Pool（对象池）

`sync.Pool` 是一个临时对象池，用于缓存和复用临时对象，减少 GC 压力：

```go
var bufPool = sync.Pool{
    New: func() any {
        return new(bytes.Buffer)
    },
}

buf := bufPool.Get().(*bytes.Buffer)
buf.Reset()
// 使用 buf...
bufPool.Put(buf) // 归还到池中
```

**注意**：Pool 中的对象可能在任何时候被 GC 回收，不适合存储持久数据。

### Map（并发安全 Map）

`sync.Map` 是并发安全的 map，适用于**读多写少**或**不同 goroutine 操作不同 key** 的场景：

```go
var m sync.Map
m.Store("key", "value")
val, ok := m.Load("key")
m.Delete("key")
m.Range(func(key, value any) bool {
    fmt.Println(key, value)
    return true // 返回 false 停止遍历
})
```

### Cond（条件变量）

`sync.Cond` 用于 goroutine 之间的条件等待和通知：

```go
var mu sync.Mutex
cond := sync.NewCond(&mu)

// 等待方
mu.Lock()
for !condition {
    cond.Wait() // 释放锁并等待通知
}
// 条件满足，继续执行
mu.Unlock()

// 通知方
mu.Lock()
condition = true
cond.Signal()   // 唤醒一个等待者
// cond.Broadcast() // 唤醒所有等待者
mu.Unlock()
```

## 标准库方案

sync 包本身就是标准库，上述所有类型都可以直接使用，无需第三方依赖。

## 代码示例

> 💻 完整可运行代码：[code-examples/01-go-core/concurrent/sync/](https://github.com/skyhe58/guide-go/tree/main/code-examples/01-go-core/concurrent/sync/)
> 🏷️ Demo 模式：Part A（直接运行）

## 常见面试题

### Q1: sync.Pool 的原理和适用场景？

**难度**：⭐⭐⭐ | **频率**：🔥🔥🔥

**答题思路**：

1. 解释 Pool 的作用——临时对象缓存，减少 GC 压力
2. 说明底层实现——每个 P 有私有缓存和共享缓存
3. 强调 GC 时 Pool 中的对象会被清除

**标准答案**：

`sync.Pool` 是一个临时对象池，每个 P（逻辑处理器）维护一个私有对象和一个共享链表。Get 时先从私有对象获取，再从共享链表获取，最后从其他 P 的共享链表偷取，都没有则调用 New 创建。Put 时优先放入私有对象。GC 时 Pool 中的所有对象会被清除。适用于频繁创建和销毁的临时对象（如 bytes.Buffer、JSON encoder），不适合存储持久数据。标准库 `fmt` 包内部就使用了 sync.Pool。

**深入追问**：
- sync.Pool 为什么不适合做连接池？（GC 时对象会被清除，连接池需要持久管理）
- sync.Pool 的 Get/Put 是否需要加锁？（私有对象无锁，共享链表使用 CAS 操作）

### Q2: Mutex 和 RWMutex 如何选择？

**难度**：⭐⭐ | **频率**：🔥🔥

**标准答案**：

如果读写比例接近或写操作频繁，使用 Mutex；如果读操作远多于写操作（如配置缓存、路由表），使用 RWMutex 可以提高并发读性能。RWMutex 的开销比 Mutex 略高，因为需要维护读计数器，所以在写多读少的场景下反而不如 Mutex。

## 常见陷阱

1. **Mutex 不可复制**：将 Mutex 作为值传递会导致锁失效，应使用指针或嵌入结构体
2. **Lock/Unlock 不配对**：忘记 Unlock 导致死锁，建议使用 `defer mu.Unlock()`
3. **WaitGroup Add 位置错误**：`wg.Add(1)` 必须在启动 goroutine 之前调用，否则可能在 Add 之前 Wait 就返回了
4. **sync.Pool 存储持久数据**：Pool 中的对象会被 GC 回收，不能用于连接池等需要持久管理的场景

## 参考资料

- [Go 标准库 sync 包](https://pkg.go.dev/sync)
- [Go Blog - Share Memory By Communicating](https://go.dev/blog/codelab-share)
- [Go 源码 - sync/mutex.go](https://github.com/golang/go/blob/master/src/sync/mutex.go)
