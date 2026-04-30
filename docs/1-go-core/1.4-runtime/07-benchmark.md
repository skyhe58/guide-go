---
title: "benchmark"
module: "runtime"
difficulty: "intermediate"
interviewFrequency: "medium"
tags:
  - benchmark
  - 性能测试
  - benchstat
  - testing.B
codeExample: "01-go-core/runtime/benchmark/"
relatedEntries:
  - "/1-go-core/1.4-runtime/05-pprof"
  - "/1-go-core/1.4-runtime/08-optimization"
  - "/1-go-core/1.5-testing/01-testing"
prerequisites:
  - "/1-go-core/1.1-go-basics/06-functions"
estimatedTime: "35min"
---

# benchmark

## 概念说明

Go 内置了 benchmark 测试框架，通过 `testing.B` 类型编写性能基准测试。benchmark 能够精确测量函数的执行时间和内存分配情况，是性能优化的数据基础——"没有数据支撑的优化都是盲目的"。

## 核心原理

### 编写 benchmark

```go
// 文件名必须以 _test.go 结尾
// 函数名必须以 Benchmark 开头
// 参数类型为 *testing.B

func BenchmarkXxx(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // 被测代码
    }
}
```

`b.N` 由测试框架自动调整，确保测试运行足够长的时间以获得稳定结果。

### 运行 benchmark

```bash
# 运行所有 benchmark
go test -bench=. -benchmem

# 运行特定 benchmark
go test -bench=BenchmarkXxx -benchmem

# 指定运行次数（用于 benchstat 对比）
go test -bench=. -benchmem -count=10

# 指定运行时间
go test -bench=. -benchtime=5s

# 输出内存分配信息
go test -bench=. -benchmem
```

### 结果解读

```
BenchmarkConcat-8    5000000    300 ns/op    48 B/op    2 allocs/op
```

| 字段 | 含义 |
|------|------|
| `BenchmarkConcat-8` | 函数名-GOMAXPROCS |
| `5000000` | 运行次数（b.N） |
| `300 ns/op` | 每次操作耗时 |
| `48 B/op` | 每次操作分配的字节数 |
| `2 allocs/op` | 每次操作的内存分配次数 |

### benchstat 对比

```bash
# 安装 benchstat
go install golang.org/x/perf/cmd/benchstat@latest

# 运行优化前后的 benchmark
go test -bench=. -count=10 > old.txt
# ... 进行优化 ...
go test -bench=. -count=10 > new.txt

# 对比结果
benchstat old.txt new.txt
```

### 避免编译器优化干扰

编译器可能优化掉没有副作用的代码，导致 benchmark 结果不准确：

```go
// ❌ 错误：结果未使用，编译器可能优化掉整个调用
func BenchmarkBad(b *testing.B) {
    for i := 0; i < b.N; i++ {
        compute()  // 可能被优化掉
    }
}

// ✅ 正确：使用全局变量保存结果，防止编译器优化
var result int

func BenchmarkGood(b *testing.B) {
    var r int
    for i := 0; i < b.N; i++ {
        r = compute()
    }
    result = r  // 防止编译器优化
}
```

## 标准库方案

```go
func BenchmarkStringConcat(b *testing.B) {
    // 重置计时器（排除初始化开销）
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        s := ""
        for j := 0; j < 100; j++ {
            s += "a"
        }
    }
}

func BenchmarkStringBuilder(b *testing.B) {
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        var sb strings.Builder
        for j := 0; j < 100; j++ {
            sb.WriteString("a")
        }
        _ = sb.String()
    }
}

// 子 benchmark
func BenchmarkSliceAppend(b *testing.B) {
    sizes := []int{10, 100, 1000, 10000}
    for _, size := range sizes {
        b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                s := make([]int, 0)
                for j := 0; j < size; j++ {
                    s = append(s, j)
                }
            }
        })
    }
}
```

## 代码示例

> 💻 完整可运行代码：[code-examples/01-go-core/runtime/benchmark/](https://github.com/your-repo/code-examples/01-go-core/runtime/benchmark/)
> 🏷️ Demo 模式：`go test -bench=. -benchmem`

## 常见面试题

### Q1: 如何编写和运行 Go benchmark？

**难度**：⭐⭐ | **频率**：🔥🔥

**答题思路**：

1. 说明文件和函数命名规范
2. 解释 b.N 的作用
3. 说明 -benchmem 的含义

**标准答案**：

benchmark 函数写在 `_test.go` 文件中，函数名以 `Benchmark` 开头，参数为 `*testing.B`。核心是 `for i := 0; i < b.N; i++` 循环，b.N 由框架自动调整以获得稳定结果。使用 `go test -bench=. -benchmem` 运行，`-benchmem` 显示内存分配信息。使用 `b.ResetTimer()` 排除初始化开销，使用全局变量防止编译器优化掉被测代码。

**深入追问**：

- 如何对比优化前后的性能？（使用 benchstat 工具，`-count=10` 多次运行取统计值）
- b.ReportAllocs() 和 -benchmem 的区别？（效果相同，ReportAllocs 写在代码中，-benchmem 是命令行参数）

## 常见陷阱

1. **编译器优化干扰**：被测函数的返回值未使用，编译器可能优化掉整个调用
2. **初始化开销未排除**：benchmark 函数中的初始化代码会被计入耗时，使用 `b.ResetTimer()`
3. **单次运行不可靠**：使用 `-count=10` 多次运行，用 benchstat 做统计分析

## 参考资料

- [Go 官方文档 - testing.B](https://pkg.go.dev/testing#B)
- [benchstat 工具](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)
