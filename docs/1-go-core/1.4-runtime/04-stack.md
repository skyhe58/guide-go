---
title: "栈管理"
module: "runtime"
difficulty: "advanced"
interviewFrequency: "medium"
tags:
  - 栈管理
  - 连续栈
  - goroutine
  - 栈增长
codeExample: "01-go-core/runtime/gmp/"
relatedEntries:
  - "/1-go-core/1.4-runtime/01-gmp"
  - "/1-go-core/1.4-runtime/03-memory"
prerequisites:
  - "/1-go-core/1.3-concurrent/01-goroutine"
estimatedTime: "30min"
---

# 栈管理

## 概念说明

每个 goroutine 都有自己独立的栈空间。Go 使用**连续栈**（Contiguous Stack）方案管理 goroutine 栈，支持动态增长和收缩，初始栈大小仅 2KB（Go 1.4+），最大可增长到 1GB。这是 Go 能够创建百万级 goroutine 的关键因素之一。

## 核心原理

### 栈管理方案演进

| Go 版本 | 方案 | 说明 |
|---------|------|------|
| Go 1.3 之前 | 分段栈（Segmented Stack） | 栈不够时分配新段，通过链表连接。缺点：频繁的栈分裂/合并导致"热分裂"问题 |
| Go 1.4+ | 连续栈（Contiguous Stack） | 栈不够时分配一个 2 倍大小的新栈，将旧栈内容复制过去 |

### 连续栈工作原理

**栈增长：**
1. 编译器在函数入口插入栈检查代码（stack check prologue）
2. 如果当前栈空间不足，触发 `runtime.morestack`
3. 分配一个 2 倍大小的新栈
4. 将旧栈内容复制到新栈
5. 调整所有指向旧栈的指针
6. 释放旧栈

**栈收缩：**
- GC 扫描阶段检查 goroutine 栈使用率
- 如果栈使用率低于 1/4，将栈缩小为当前的 1/2
- 收缩操作在 GC 期间完成

### goroutine 栈大小

| 参数 | 值 | 说明 |
|------|-----|------|
| 初始大小 | 2KB（Go 1.4+） | 之前版本为 4KB 或 8KB |
| 最大大小 | 1GB（64 位系统） | 超过后 panic: stack overflow |
| 增长倍数 | 2x | 每次增长为当前的 2 倍 |
| 收缩阈值 | 使用率 < 1/4 | GC 时检查并收缩 |

### 为什么不用固定大小栈？

| 方案 | 优点 | 缺点 |
|------|------|------|
| 固定大小（如 OS 线程 1-8MB） | 简单 | 百万 goroutine × 1MB = 1TB 内存，不可行 |
| 动态增长（Go 方案） | 初始 2KB，按需增长 | 栈复制有开销，但摊销后可忽略 |

## 标准库方案

```go
package main

import (
    "fmt"
    "runtime"
)

func main() {
    // 查看当前 goroutine 栈使用情况
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    fmt.Printf("栈使用量: %d KB\n", m.StackInuse/1024)
    fmt.Printf("栈系统分配: %d KB\n", m.StackSys/1024)

    // 获取当前 goroutine 的栈信息
    buf := make([]byte, 4096)
    n := runtime.Stack(buf, false)
    fmt.Printf("当前 goroutine 栈:\n%s\n", buf[:n])
}
```

## 代码示例

> 💻 完整可运行代码：[code-examples/01-go-core/runtime/gmp/](https://github.com/skyhe58/guide-go/tree/main/code-examples/01-go-core/runtime/gmp/)
> 🏷️ Demo 模式：Part A（直接运行）

## 常见面试题

### Q1: goroutine 的栈是如何管理的？

**难度**：⭐⭐⭐ | **频率**：🔥🔥

**答题思路**：

1. 说明初始大小和动态增长机制
2. 解释连续栈方案
3. 对比 OS 线程栈

**标准答案**：

goroutine 初始栈大小为 2KB，使用连续栈方案动态管理。当栈空间不足时，运行时分配一个 2 倍大小的新栈，将旧栈内容复制过去并调整指针。GC 时如果栈使用率低于 1/4 则收缩为 1/2。最大栈大小为 1GB。相比 OS 线程固定 1-8MB 的栈，goroutine 的动态栈使得百万级并发成为可能。

**深入追问**：

- 栈复制时如何调整指针？（运行时遍历栈帧，将所有指向旧栈的指针偏移到新栈地址）
- 什么是"热分裂"问题？（分段栈方案中，函数调用恰好在栈边界时频繁分配/释放栈段）

## 常见陷阱

1. **深度递归**：无限递归会导致栈持续增长直到 1GB 上限后 panic
2. **CGO 调用**：CGO 调用会切换到 OS 线程栈（默认 8MB），不受 goroutine 栈管理
3. **大数组局部变量**：在函数中声明大数组会快速消耗栈空间，考虑使用 slice（堆分配）

## 参考资料

- [Go 连续栈设计文档](https://docs.google.com/document/d/1wAaf1rYoM4namerELAssN_v-lv4eSPQDSOXiDGAges0)
- [Go 运行时源码 - stack.go](https://go.dev/src/runtime/stack.go)
