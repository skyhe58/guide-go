---
title: "go vet 静态分析"
module: "testing-tools"
difficulty: "beginner"
interviewFrequency: "low"
tags:
  - go vet
  - 静态分析
  - 代码检查
codeExample: "01-go-core/testing-tools/"
relatedEntries:
  - "/1-go-core/1.5-testing/10-golangci-lint"
  - "/1-go-core/1.5-testing/11-tools"
prerequisites:
  - "/1-go-core/1.1-go-basics/06-functions"
estimatedTime: "20min"
---

# go vet 静态分析

## 概念说明

`go vet` 是 Go 内置的静态分析工具，用于检查代码中的常见错误。它不检查代码风格，而是发现编译器无法捕获但很可能是 bug 的问题，如 `fmt.Printf` 格式化参数不匹配、无法到达的代码、错误的结构体标签等。

## 核心原理

### 基本用法

```bash
# 检查当前包
go vet ./...

# 检查特定包
go vet ./mypackage/

# 指定检查器
go vet -printf ./...
```

### 常见检查项

| 检查器 | 检查内容 | 示例 |
|--------|---------|------|
| `printf` | Printf 格式化参数不匹配 | `fmt.Printf("%d", "hello")` |
| `shadow` | 变量遮蔽 | 内层作用域重新声明同名变量 |
| `structtag` | 结构体标签格式错误 | `` `json:"name" xml:name` `` |
| `unreachable` | 不可达代码 | return 后的语句 |
| `unusedresult` | 未使用的函数返回值 | `fmt.Errorf(...)` 结果未使用 |
| `copylocks` | 复制了包含锁的值 | 值传递 `sync.Mutex` |
| `loopclosure` | 循环变量闭包捕获 | Go 1.22 前的 for-range 陷阱 |
| `nilfunc` | nil 函数比较 | `if fn == nil` 但 fn 永远非 nil |
| `assign` | 自赋值检测 | `x = x` |

### 典型错误示例

```go
// ❌ Printf 格式化参数不匹配
fmt.Printf("%d", "hello") // go vet: Printf format %d has arg "hello" of wrong type string

// ❌ 复制了包含锁的值
var mu sync.Mutex
mu2 := mu // go vet: assignment copies lock value

// ❌ 结构体标签格式错误
type User struct {
    Name string `json:"name" xml:name` // go vet: struct field tag `xml:name` not compatible
}
```

## 标准库方案

`go vet` 是 Go 工具链的一部分，无需安装。它是 `go test` 的默认前置步骤（Go 1.10+）。

```bash
# go test 默认会先运行 go vet
go test ./...

# 跳过 vet（不推荐）
go test -vet=off ./...
```

## 第三方库方案

`go vet` 的检查能力有限，更全面的静态分析推荐使用 [golangci-lint](/1-go-core/1.5-testing/10-golangci-lint)，它集成了 `go vet` 和数十个其他 linter。

## 代码示例

> 💻 `go vet` 是命令行工具，直接在项目中运行即可
> 🏷️ Demo 模式：Part A（直接运行）

```bash
cd code-examples
go vet ./...
```

## 常见面试题

### Q1: go vet 和 golangci-lint 的区别？

**难度**：⭐ | **频率**：🔥

**答题思路**：

1. go vet 的定位和检查范围
2. golangci-lint 的定位和优势
3. 两者的关系

**标准答案**：

`go vet` 是 Go 内置的轻量级静态分析工具，检查编译器无法发现的常见 bug（格式化参数、锁复制、结构体标签等）。`golangci-lint` 是第三方多 linter 聚合工具，集成了 `go vet` 和数十个其他 linter（staticcheck、errcheck、gosec 等），检查范围更广。`go vet` 是基础，`golangci-lint` 是增强。

## 常见陷阱

1. **忽略 go vet 警告**：go vet 报告的问题几乎都是真实 bug，不应忽略
2. **只依赖 go vet**：go vet 检查范围有限，应配合 golangci-lint 使用

## 参考资料

- [Go 官方 vet 文档](https://pkg.go.dev/cmd/vet)
- [Go Blog: Introducing Go Vet](https://go.dev/blog/vet)
