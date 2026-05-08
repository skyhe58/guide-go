---
title: "trace 工具"
module: "runtime"
difficulty: "advanced"
interviewFrequency: "medium"
tags:
  - trace
  - 执行追踪
  - 调度延迟
  - 性能分析
codeExample: "01-go-core/runtime/pprof/"
relatedEntries:
  - "/1-go-core/1.4-runtime/05-pprof"
  - "/1-go-core/1.4-runtime/01-gmp"
prerequisites:
  - "/1-go-core/1.4-runtime/01-gmp"
  - "/1-go-core/1.4-runtime/05-pprof"
estimatedTime: "35min"
---

# trace 工具

## 概念说明

`go tool trace` 是 Go 内置的执行追踪工具，能够记录程序运行期间的所有运行时事件（goroutine 创建/阻塞/唤醒、GC、系统调用、网络 I/O 等），并以时间线的形式可视化展示。与 pprof 的采样分析不同，trace 记录的是**精确的事件序列**，适合分析调度延迟、GC 影响、并发瓶颈等问题。

## 核心原理

### trace vs pprof

| 维度 | pprof | trace |
|------|-------|-------|
| 分析方式 | 采样（统计） | 事件记录（精确） |
| 时间粒度 | 粗（采样间隔） | 细（纳秒级事件） |
| 适用场景 | CPU/内存热点分析 | 调度延迟/GC 影响/并发分析 |
| 开销 | 低（CPU ~5%） | 较高（建议短时间采集） |
| 输出 | 文本/SVG | 交互式时间线 |

### trace 能看到什么

- **goroutine 调度**：创建、运行、阻塞、唤醒的完整时间线
- **GC 事件**：GC 开始/结束、STW 时间、辅助标记
- **系统调用**：syscall 阻塞时间
- **网络 I/O**：网络轮询事件
- **P 的利用率**：每个 P 在做什么（运行 G / GC / 空闲）

### 使用方式

```bash
# 方式一：通过 net/http/pprof 采集（HTTP 服务）
curl -o trace.out http://localhost:6060/debug/pprof/trace?seconds=5
go tool trace trace.out

# 方式二：通过代码采集（命令行工具）
# 在代码中使用 runtime/trace 包

# 方式三：通过 go test 采集
go test -trace=trace.out ./...
go tool trace trace.out
```

## 标准库方案

```go
package main

import (
    "fmt"
    "os"
    "runtime/trace"
)

func main() {
    // 创建 trace 输出文件
    f, err := os.Create("trace.out")
    if err != nil {
        panic(err)
    }
    defer f.Close()

    // 开始追踪
    if err := trace.Start(f); err != nil {
        panic(err)
    }
    defer trace.Stop()

    // 你的业务逻辑...
    fmt.Println("正在追踪...")

    // 自定义 task 和 region（Go 1.11+）
    ctx, task := trace.NewTask(nil, "myTask")
    defer task.End()

    trace.WithRegion(ctx, "myRegion", func() {
        // 被追踪的代码区域
        fmt.Println("执行关键操作")
    })
}
// 运行后：go tool trace trace.out
```

## 代码示例

> 💻 完整可运行代码：[code-examples/01-go-core/runtime/pprof/](https://github.com/skyhe58/guide-go/tree/main/code-examples/01-go-core/runtime/pprof/)
> 🏷️ Demo 模式：Part A（直接运行）

## 常见面试题

### Q1: go tool trace 和 pprof 有什么区别？

**难度**：⭐⭐⭐ | **频率**：🔥🔥

**答题思路**：

1. 对比分析方式（采样 vs 事件记录）
2. 说明各自适用场景
3. 举例说明何时用 trace

**标准答案**：

pprof 是采样分析，适合找 CPU/内存热点函数；trace 是事件记录，记录所有运行时事件的精确时间线。当需要分析 goroutine 调度延迟、GC 对延迟的影响、并发瓶颈时用 trace；当需要找出 CPU 消耗最多的函数或内存分配热点时用 pprof。trace 开销较大，建议短时间采集（几秒）。

**深入追问**：

- trace 中如何分析 GC 对延迟的影响？（查看 GC 事件的 STW 时间和辅助标记占比）
- 如何自定义 trace 中的 task 和 region？（使用 `trace.NewTask` 和 `trace.WithRegion`）

## 常见陷阱

1. **长时间采集 trace**：trace 开销较大，采集时间过长会产生巨大的 trace 文件，建议 5-10 秒
2. **忽略调度延迟**：goroutine 从就绪到实际运行的延迟可能是性能瓶颈，trace 能直观展示
3. **只用 pprof 不用 trace**：pprof 看不到时间维度的信息，某些问题只有 trace 能发现

## 参考资料

- [Go 官方文档 - runtime/trace](https://pkg.go.dev/runtime/trace)
- [Go 执行追踪器设计文档](https://docs.google.com/document/d/1FP5apqzBgr7ahCCgFO-yoVhk4YZrNIDNf9RybNBa4wI)
