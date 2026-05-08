---
title: "testing 包"
module: "testing-tools"
difficulty: "intermediate"
interviewFrequency: "high"
tags:
  - testing
  - 单元测试
  - 表驱动测试
  - t.Run
  - TestMain
  - t.Parallel
  - t.Helper
codeExample: "01-go-core/testing-tools/tabledriven/"
relatedEntries:
  - "/1-go-core/1.5-testing/02-benchmark"
  - "/1-go-core/1.5-testing/08-best-practices"
prerequisites:
  - "/1-go-core/1.1-go-basics/06-functions"
  - "/1-go-core/1.1-go-basics/08-struct-method"
estimatedTime: "45min"
---

# testing 包

## 概念说明

Go 内置的 `testing` 包是 Go 测试体系的基石。与 Java 需要引入 JUnit 不同，Go 的测试框架是标准库的一部分，通过 `go test` 命令即可运行。Go 测试遵循"约定优于配置"的原则：文件名以 `_test.go` 结尾、函数名以 `Test` 开头、参数为 `*testing.T`。

## 核心原理

### 基本单元测试

```go
// math_test.go
package math

import "testing"

func TestAdd(t *testing.T) {
    got := Add(1, 2)
    want := 3
    if got != want {
        t.Errorf("Add(1, 2) = %d, want %d", got, want)
    }
}
```

### 表驱动测试（Table-Driven Tests）

表驱动测试是 Go 社区最推崇的测试模式，将测试用例组织为表格结构，避免重复代码：

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"正数相加", 1, 2, 3},
        {"负数相加", -1, -2, -3},
        {"零值", 0, 0, 0},
        {"正负混合", 1, -1, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.want {
                t.Errorf("Add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

### 子测试 t.Run

`t.Run` 创建子测试，支持：
- 独立命名，失败时精确定位
- 可单独运行：`go test -run TestAdd/正数相加`
- 与 `t.Parallel()` 配合实现并行测试

### TestMain

`TestMain` 是测试的入口函数，用于全局的 setup/teardown：

```go
func TestMain(m *testing.M) {
    // setup：初始化数据库连接、启动服务等
    setup()

    code := m.Run() // 运行所有测试

    // teardown：清理资源
    teardown()
    os.Exit(code)
}
```

### t.Parallel 并行测试

```go
func TestParallel(t *testing.T) {
    tests := []struct {
        name  string
        input int
    }{
        {"case1", 1},
        {"case2", 2},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel() // 标记为可并行执行
            // 注意：Go 1.22+ for-range 变量语义已修复
            // 不再需要 tt := tt 的 shadow 技巧
            result := Process(tt.input)
            if result != tt.input*2 {
                t.Errorf("got %d, want %d", result, tt.input*2)
            }
        })
    }
}
```

### t.Helper

`t.Helper()` 标记辅助函数，使错误报告指向调用者而非辅助函数内部：

```go
func assertEqual(t *testing.T, got, want int) {
    t.Helper() // 错误信息会指向调用 assertEqual 的那一行
    if got != want {
        t.Errorf("got %d, want %d", got, want)
    }
}
```

## 标准库方案

Go 的 `testing` 包是标准库方案，无需任何第三方依赖。常用 API：

| 方法 | 说明 |
|------|------|
| `t.Error/t.Errorf` | 报告错误但继续执行 |
| `t.Fatal/t.Fatalf` | 报告错误并立即终止当前测试 |
| `t.Skip/t.Skipf` | 跳过当前测试 |
| `t.Log/t.Logf` | 输出日志（仅 `-v` 模式可见） |
| `t.Run` | 创建子测试 |
| `t.Parallel` | 标记并行测试 |
| `t.Helper` | 标记辅助函数 |
| `t.Cleanup` | 注册清理函数（LIFO 顺序执行） |

## 第三方库方案

- **testify**：提供 `assert`/`require` 断言库，简化断言语法
- **go-cmp**：Google 出品，提供更强大的深度比较功能（`cmp.Diff`）

```go
// testify 风格
import "github.com/stretchr/testify/assert"

func TestAdd(t *testing.T) {
    assert.Equal(t, 3, Add(1, 2))
}

// go-cmp 风格
import "github.com/google/go-cmp/cmp"

func TestStruct(t *testing.T) {
    got := GetUser()
    want := User{Name: "Alice", Age: 30}
    if diff := cmp.Diff(want, got); diff != "" {
        t.Errorf("mismatch (-want +got):\n%s", diff)
    }
}
```

## 代码示例

> 💻 完整可运行代码：[code-examples/01-go-core/testing-tools/tabledriven/](https://github.com/skyhe58/guide-go/tree/main/code-examples/01-go-core/testing-tools/tabledriven/)
> 🏷️ Demo 模式：Part A（直接运行）

## 常见面试题

### Q1: 什么是表驱动测试？为什么 Go 推荐这种模式？

**难度**：⭐⭐ | **频率**：🔥🔥🔥

**答题思路**：

1. 表驱动测试的定义和结构
2. 与传统测试的对比优势
3. 配合 t.Run 的好处

**标准答案**：

表驱动测试将测试用例组织为结构体切片，每个元素包含输入和期望输出，通过循环遍历执行。优势：减少重复代码、新增用例只需加一行、配合 `t.Run` 可独立运行和命名、失败时精确定位到具体用例。Go 标准库自身大量使用这种模式。

**深入追问**：

- t.Parallel 在表驱动测试中需要注意什么？
- Go 1.22 之前的 for-range 变量捕获陷阱是什么？

### Q2: TestMain 的使用场景是什么？

**难度**：⭐⭐ | **频率**：🔥🔥

**答题思路**：

1. TestMain 的签名和执行时机
2. 典型使用场景
3. 注意事项

**标准答案**：

`TestMain(m *testing.M)` 是包级别的测试入口，在所有测试函数之前执行。典型场景：初始化数据库连接、启动测试容器、设置环境变量、全局 setup/teardown。必须调用 `m.Run()` 运行测试并通过 `os.Exit(code)` 退出。

**深入追问**：

- 一个包可以有多个 TestMain 吗？（不可以，每个包最多一个）

## 常见陷阱

1. **忘记调用 os.Exit**：TestMain 中必须调用 `os.Exit(m.Run())`，否则测试结果不会正确反映
2. **并行测试中的变量捕获**：Go 1.22 之前，for-range 中的循环变量在闭包中需要 shadow（`tt := tt`）
3. **t.Fatal 在 goroutine 中使用**：`t.Fatal` 只能在测试 goroutine 中调用，不能在子 goroutine 中使用

## 参考资料

- [Go 官方 testing 包文档](https://pkg.go.dev/testing)
- [Go Wiki: Table Driven Tests](https://go.dev/wiki/TableDrivenTests)
- [Go Blog: Using Subtests and Sub-benchmarks](https://go.dev/blog/subtests)
