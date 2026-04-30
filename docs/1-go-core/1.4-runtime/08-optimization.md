---
title: "常见优化技巧"
module: "runtime"
difficulty: "intermediate"
interviewFrequency: "high"
tags:
  - 性能优化
  - sync.Pool
  - strings.Builder
  - 预分配
  - 减少逃逸
codeExample: "01-go-core/runtime/benchmark/"
relatedEntries:
  - "/1-go-core/1.4-runtime/07-benchmark"
  - "/1-go-core/1.4-runtime/09-escape"
  - "/1-go-core/1.4-runtime/02-gc"
prerequisites:
  - "/1-go-core/1.4-runtime/03-memory"
estimatedTime: "40min"
---

# 常见优化技巧

## 概念说明

Go 性能优化的核心思路是**减少内存分配**和**降低 GC 压力**。大部分性能问题都可以归结为：不必要的堆分配、频繁的小对象创建、低效的数据结构使用。本节介绍 Go 开发中最常用的优化技巧。

## 核心原理

### 1. sync.Pool 对象复用

`sync.Pool` 是一个临时对象池，用于缓存和复用临时对象，减少内存分配和 GC 压力。

```go
var bufPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func process() {
    buf := bufPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufPool.Put(buf)
    }()
    // 使用 buf...
}
```

**注意：** Pool 中的对象可能在任意时刻被 GC 回收，不要存放需要持久化的数据。

### 2. strings.Builder 字符串拼接

```go
// ❌ 低效：每次 += 都会分配新字符串
func concatBad(n int) string {
    s := ""
    for i := 0; i < n; i++ {
        s += "a"
    }
    return s
}

// ✅ 高效：strings.Builder 内部使用 []byte，减少分配
func concatGood(n int) string {
    var sb strings.Builder
    sb.Grow(n) // 预分配容量
    for i := 0; i < n; i++ {
        sb.WriteString("a")
    }
    return sb.String()
}
```

### 3. slice 预分配

```go
// ❌ 低效：频繁扩容，多次内存分配和复制
func sliceBad(n int) []int {
    s := make([]int, 0)
    for i := 0; i < n; i++ {
        s = append(s, i)
    }
    return s
}

// ✅ 高效：预分配容量，一次分配
func sliceGood(n int) []int {
    s := make([]int, 0, n)
    for i := 0; i < n; i++ {
        s = append(s, i)
    }
    return s
}
```

### 4. map 预分配

```go
// ❌ 低效：频繁扩容和 rehash
m := make(map[string]int)

// ✅ 高效：预分配容量
m := make(map[string]int, 1000)
```

### 5. 减少内存逃逸

```go
// ❌ 逃逸到堆：返回指针
func newUser() *User {
    u := User{Name: "test"}
    return &u  // u 逃逸到堆
}

// ✅ 栈分配：传入指针填充
func fillUser(u *User) {
    u.Name = "test"  // u 由调用方管理
}

// ❌ 逃逸：interface{} 参数
fmt.Println(x)  // x 逃逸（fmt.Println 接受 interface{}）

// ✅ 减少逃逸：使用具体类型
fmt.Printf("%d\n", x)  // 某些情况下编译器可以优化
```

### 优化技巧速查表

| 技巧 | 场景 | 效果 |
|------|------|------|
| `sync.Pool` | 频繁创建/销毁的临时对象 | 减少 GC 压力 |
| `strings.Builder` | 大量字符串拼接 | 减少内存分配 |
| slice 预分配 | 已知大小的 slice | 避免扩容开销 |
| map 预分配 | 已知大小的 map | 避免 rehash |
| 减少逃逸 | 热路径上的对象分配 | 栈分配替代堆分配 |
| 避免 `[]byte` ↔ `string` 转换 | 高频转换场景 | 减少内存复制 |
| 使用 `strconv` 替代 `fmt.Sprintf` | 简单类型转字符串 | 减少反射开销 |

## 标准库方案

```go
package main

import (
    "bytes"
    "fmt"
    "strings"
    "sync"
)

// sync.Pool 示例
var pool = sync.Pool{
    New: func() interface{} {
        return bytes.NewBuffer(make([]byte, 0, 1024))
    },
}

func main() {
    // strings.Builder
    var sb strings.Builder
    sb.Grow(100)
    for i := 0; i < 100; i++ {
        sb.WriteByte('a')
    }
    fmt.Println(sb.Len())

    // sync.Pool
    buf := pool.Get().(*bytes.Buffer)
    buf.WriteString("hello")
    fmt.Println(buf.String())
    buf.Reset()
    pool.Put(buf)
}
```

## 代码示例

> 💻 完整可运行代码：[code-examples/01-go-core/runtime/benchmark/](https://github.com/your-repo/code-examples/01-go-core/runtime/benchmark/)
> 🏷️ Demo 模式：`go test -bench=. -benchmem`

## 常见面试题

### Q1: Go 中有哪些常见的性能优化手段？

**难度**：⭐⭐⭐ | **频率**：🔥🔥🔥

**答题思路**：

1. 从减少内存分配的角度列举
2. 每个技巧给出具体场景
3. 强调"先 benchmark 再优化"

**标准答案**：

核心思路是减少堆分配和 GC 压力：(1) sync.Pool 复用临时对象；(2) strings.Builder 替代字符串 += 拼接；(3) slice/map 预分配容量；(4) 减少逃逸（避免不必要的指针返回、减少 interface{} 使用）；(5) 使用 strconv 替代 fmt.Sprintf；(6) 避免频繁的 []byte 和 string 转换。优化前必须先用 benchmark 和 pprof 定位瓶颈，避免盲目优化。

**深入追问**：

- sync.Pool 的对象什么时候会被回收？（每次 GC 时清空 Pool）
- 为什么 strings.Builder 比 += 快？（内部使用 []byte，避免每次拼接创建新字符串）

## 常见陷阱

1. **过早优化**：先保证正确性，再用 benchmark 定位瓶颈，最后针对性优化
2. **sync.Pool 存放持久对象**：Pool 中的对象随时可能被 GC 回收
3. **预分配过大**：预分配容量过大浪费内存，应根据实际数据量估算

## 参考资料

- [Go 官方文档 - sync.Pool](https://pkg.go.dev/sync#Pool)
- [Go 官方文档 - strings.Builder](https://pkg.go.dev/strings#Builder)
- [High Performance Go Workshop - Dave Cheney](https://dave.cheney.net/high-performance-go-workshop/dotgo-paris.html)
