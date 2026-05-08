---
title: "运算符"
module: "go-basics"
difficulty: "beginner"
interviewFrequency: "low"
tags:
  - 运算符
  - 位运算
  - 取地址
  - 解引用
codeExample: "01-go-core/go-basics/operators/"
relatedEntries:
  - "/1-go-core/1.1-go-basics/02-data-types"
  - "/1-go-core/1.1-go-basics/11-pointer"
prerequisites:
  - "/1-go-core/1.1-go-basics/02-data-types"
estimatedTime: "20min"
---

# 运算符

## 概念说明

Go 的运算符设计简洁，没有运算符重载，没有三元运算符（`?:`），这些都是 Go "少即是多"哲学的体现。Go 的位运算在权限控制、标志位、底层编程中非常常用。

## 核心原理

### 运算符分类

| 类别 | 运算符 | 说明 |
|------|--------|------|
| 算术 | `+` `-` `*` `/` `%` | 加减乘除取模 |
| 关系 | `==` `!=` `<` `>` `<=` `>=` | 比较运算 |
| 逻辑 | `&&` `\|\|` `!` | 与或非（短路求值） |
| 位运算 | `&` `\|` `^` `<<` `>>` `&^` | 与或异或左移右移位清除 |
| 赋值 | `=` `:=` `+=` `-=` `*=` `/=` `%=` `&=` `\|=` `^=` `<<=` `>>=` | 赋值及复合赋值 |
| 取地址/解引用 | `&` `*` | 取变量地址 / 解引用指针 |
| 通道 | `<-` | 发送/接收数据 |

### 位运算详解

```go
// 位运算在权限控制中的应用
const (
    Read    = 1 << iota // 001 = 1
    Write               // 010 = 2
    Execute             // 100 = 4
)

perm := Read | Write    // 011 = 3（读+写权限）
hasRead := perm & Read != 0  // true（检查是否有读权限）

// &^ 位清除运算符（Go 特有）
perm = perm &^ Write    // 001 = 1（移除写权限）
```

### 取地址与解引用

```go
x := 42
p := &x    // p 是 *int 类型，值为 x 的内存地址
fmt.Println(*p) // 42（解引用，获取指针指向的值）
*p = 100
fmt.Println(x)  // 100（通过指针修改了原变量）
```

## 标准库方案

```go
package main

import "fmt"

func main() {
    // Go 没有三元运算符，用 if-else 替代
    x := 10
    var result string
    if x > 5 {
        result = "大于5"
    } else {
        result = "不大于5"
    }
    fmt.Println(result)
}
```

## 代码示例

> 💻 完整可运行代码：[code-examples/01-go-core/go-basics/operators/](https://github.com/skyhe58/guide-go/tree/main/code-examples/01-go-core/go-basics/operators/)

## 常见面试题

### Q1: Go 的 `&^` 运算符是什么？

**难度**：⭐⭐ | **频率**：🔥

**标准答案**：

`&^` 是位清除（bit clear）运算符，`a &^ b` 的含义是：如果 b 的某位为 1，则结果对应位为 0；否则保持 a 的对应位不变。等价于 `a & (^b)`。常用于清除标志位。

## 常见陷阱

1. **整数除法截断**：`7 / 2 = 3`（不是 3.5），需要浮点除法时先转换类型
2. **没有三元运算符**：Go 故意不提供 `?:`，必须用 `if-else`
3. **不支持运算符重载**：不能为自定义类型定义 `+` 等运算符

## 参考资料

- [Go 语言规范 - 运算符](https://go.dev/ref/spec#Operators)
- [Go 语言规范 - 运算符优先级](https://go.dev/ref/spec#Operator_precedence)
