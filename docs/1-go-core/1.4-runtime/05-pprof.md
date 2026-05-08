---
title: "pprof 性能分析"
module: "runtime"
difficulty: "intermediate"
interviewFrequency: "high"
tags:
  - pprof
  - 性能分析
  - CPU profiling
  - 内存分析
  - 线上排查
codeExample: "01-go-core/runtime/pprof/"
relatedEntries:
  - "/1-go-core/1.4-runtime/06-trace"
  - "/1-go-core/1.4-runtime/07-benchmark"
  - "/1-go-core/1.4-runtime/08-optimization"
prerequisites:
  - "/1-go-core/1.1-go-basics/06-functions"
estimatedTime: "45min"
---

# pprof 性能分析

## 概念说明

pprof 是 Go 内置的性能分析工具，可以采集 CPU、内存、goroutine、阻塞、互斥锁等多维度的性能数据，是 Go 开发者排查线上性能问题的核心利器。

Go 提供两种使用 pprof 的方式：
- `runtime/pprof`：适用于命令行工具，手动开始/停止采集
- `net/http/pprof`：适用于 HTTP 服务，通过 HTTP 端点暴露性能数据

## 核心原理

### Profile 类型

| Profile 类型 | 说明 | 常见场景 |
|-------------|------|---------|
| **CPU** | 采样 CPU 使用情况，找出热点函数 | CPU 飙高排查 |
| **Heap（内存）** | 堆内存分配情况 | 内存泄漏排查 |
| **Goroutine** | 所有 goroutine 的栈信息 | goroutine 泄漏排查 |
| **Block** | 阻塞操作（channel/mutex 等待） | 并发瓶颈分析 |
| **Mutex** | 互斥锁竞争情况 | 锁竞争排查 |
| **Allocs** | 内存分配采样 | 分配热点分析 |
| **Threadcreate** | OS 线程创建情况 | 线程泄漏排查 |

### net/http/pprof 集成

```go
import _ "net/http/pprof"

// 启动 HTTP 服务后，访问以下端点：
// http://localhost:6060/debug/pprof/           — 索引页
// http://localhost:6060/debug/pprof/profile    — CPU profile（默认 30s）
// http://localhost:6060/debug/pprof/heap       — 堆内存
// http://localhost:6060/debug/pprof/goroutine  — goroutine
// http://localhost:6060/debug/pprof/block      — 阻塞
// http://localhost:6060/debug/pprof/mutex      — 互斥锁
```

### go tool pprof 交互式分析

```bash
# 采集 CPU profile（30 秒）
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 采集堆内存 profile
go tool pprof http://localhost:6060/debug/pprof/heap

# 采集 goroutine profile
go tool pprof http://localhost:6060/debug/pprof/goroutine

# 常用交互命令
# top       — 显示资源消耗最多的函数
# list func — 显示函数的逐行分析
# web       — 生成 SVG 调用图（需要 graphviz）
# png       — 生成 PNG 调用图
```

## 标准库方案

```go
package main

import (
    "log"
    "net/http"
    _ "net/http/pprof" // 自动注册 /debug/pprof/ 路由
    "os"
    "runtime/pprof"
)

func main() {
    // 方式一：net/http/pprof（HTTP 服务）
    go func() {
        log.Println(http.ListenAndServe(":6060", nil))
    }()

    // 方式二：runtime/pprof（命令行工具）
    f, _ := os.Create("cpu.prof")
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    // 你的业务逻辑...

    // 写入堆内存 profile
    heapFile, _ := os.Create("heap.prof")
    defer heapFile.Close()
    pprof.WriteHeapProfile(heapFile)
}
```

## 代码示例

> 💻 完整可运行代码：[code-examples/01-go-core/runtime/pprof/](https://github.com/skyhe58/guide-go/tree/main/code-examples/01-go-core/runtime/pprof/)
> 🏷️ Demo 模式：Part A（直接运行）

## 常见面试题

### Q1: 如何排查 Go 服务的 CPU 飙高问题？

**难度**：⭐⭐⭐ | **频率**：🔥🔥🔥

**答题思路**：

1. 使用 pprof 采集 CPU profile
2. 分析热点函数
3. 定位具体代码行

**标准答案**：

1. 确保服务集成了 `net/http/pprof`
2. 使用 `go tool pprof http://host:port/debug/pprof/profile?seconds=30` 采集 30 秒 CPU profile
3. 在交互式界面使用 `top` 查看 CPU 消耗最多的函数
4. 使用 `list funcName` 查看具体代码行的 CPU 消耗
5. 使用 `web` 生成调用图，直观查看调用链

**深入追问**：

- 生产环境如何安全地暴露 pprof？（单独端口、内网访问、鉴权中间件）
- pprof 采集对性能有多大影响？（CPU profile 约 5% 开销，heap profile 几乎无开销）

## 常见陷阱

1. **生产环境暴露 pprof**：不要在公网端口暴露 `/debug/pprof/`，应使用独立端口并限制访问
2. **忘记开启 block/mutex profile**：需要手动调用 `runtime.SetBlockProfileRate` 和 `runtime.SetMutexProfileFraction`
3. **只看 top 不看调用图**：top 只显示直接消耗，调用图能展示完整的调用链路

## 参考资料

- [Go 官方文档 - pprof](https://pkg.go.dev/runtime/pprof)
- [Go 官方博客 - Profiling Go Programs](https://go.dev/blog/pprof)
- [net/http/pprof 文档](https://pkg.go.dev/net/http/pprof)
