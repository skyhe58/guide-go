---
title: Go 初学者路径
description: 零基础到能独立开发 Go 项目的系统学习路径
---

# Go 初学者路径

## 适合人群

- 编程零基础或有其他语言（Java/Python/JavaScript）基础，想转学 Go 的开发者
- 计算机专业在校生，准备用 Go 做毕业设计或实习项目
- 对 Go 感兴趣，想系统入门的自学者

## 预计时间

**4～6 周**（每天 2～3 小时）

## 学习步骤

### 第一阶段：环境搭建与基础语法（第 1～2 周）

> 目标：掌握 Go 基础语法，能写出简单的命令行程序

| 步骤 | 知识点 | 文档链接 | 代码示例 | 建议时间 |
|------|--------|---------|---------|---------|
| 1 | 环境搭建 | [环境搭建](/1-go-core/1.1-go-basics/01-environment) | — | 1h |
| 2 | 数据类型 | [数据类型](/1-go-core/1.1-go-basics/02-data-types) | `01-go-core/go-basics/datatypes/` | 2h |
| 3 | 变量与常量 | [变量、常量与作用域](/1-go-core/1.1-go-basics/03-variables) | `01-go-core/go-basics/variables/` | 2h |
| 4 | 运算符 | [运算符](/1-go-core/1.1-go-basics/04-operators) | `01-go-core/go-basics/operators/` | 1h |
| 5 | 控制流 | [控制流](/1-go-core/1.1-go-basics/05-control-flow) | `01-go-core/go-basics/controlflow/` | 3h |
| 6 | 函数 | [函数](/1-go-core/1.1-go-basics/06-functions) | `01-go-core/go-basics/functions/` | 3h |
| 7 | 错误处理 | [错误处理](/1-go-core/1.1-go-basics/07-error-handling) | `01-go-core/go-basics/errorhandling/` | 2h |
| 8 | 结构体与方法 | [结构体与方法](/1-go-core/1.1-go-basics/08-struct-method) | `01-go-core/go-basics/structs/` | 3h |
| 9 | 数组与切片 | [数组与切片](/1-go-core/1.1-go-basics/09-slice) | `01-go-core/go-basics/slice/` | 3h |
| 10 | Map | [Map](/1-go-core/1.1-go-basics/10-map) | `01-go-core/go-basics/maps/` | 2h |
| 11 | 指针 | [指针](/1-go-core/1.1-go-basics/11-pointer) | `01-go-core/go-basics/pointers/` | 2h |
| 12 | 包管理 | [包管理与 Go Module](/1-go-core/1.1-go-basics/12-module) | `01-go-core/go-basics/modules/` | 2h |
| 13 | 字符串处理 | [字符串处理](/1-go-core/1.1-go-basics/13-string) | `01-go-core/go-basics/strings/` | 2h |

**🏁 里程碑检查点 1：**
- [ ] 能独立编写包含控制流、函数、结构体的 Go 程序
- [ ] 理解 slice 和 map 的基本用法
- [ ] 理解 Go 的错误处理模式（`if err != nil`）
- [ ] 能使用 `go mod init` 创建项目并管理依赖

### 第二阶段：综合练习与测试入门（第 3 周）

> 目标：通过实战项目巩固基础，学会写测试

| 步骤 | 知识点 | 文档链接 | 代码示例 | 建议时间 |
|------|--------|---------|---------|---------|
| 14 | 综合练习项目 | [Go 基础语法模块](/1-go-core/1.1-go-basics/) | `01-go-core/go-basics/project-todo-cli/` | 4h |
| 15 | testing 包 | [testing 包](/1-go-core/1.5-testing/01-testing) | `01-go-core/testing-tools/tabledriven/` | 3h |
| 16 | HTTP 测试 | [HTTP 测试](/1-go-core/1.5-testing/07-httptest) | `01-go-core/testing-tools/httptest/` | 2h |
| 17 | 基础面试题复习 | [Go 基础面试指南](/1-go-core/1.1-go-basics/interview) | — | 3h |

**🏁 里程碑检查点 2：**
- [ ] 完成命令行 TODO 工具项目
- [ ] 能编写表驱动测试（Table-Driven Tests）
- [ ] 能回答 slice 扩容、defer 执行顺序等基础面试题

### 第三阶段：并发编程入门（第 4 周）

> 目标：理解 goroutine 和 channel，掌握 Go 的核心竞争力

| 步骤 | 知识点 | 文档链接 | 代码示例 | 建议时间 |
|------|--------|---------|---------|---------|
| 18 | goroutine | [goroutine](/1-go-core/1.3-concurrent/01-goroutine) | `01-go-core/concurrent/goroutine/` | 3h |
| 19 | channel | [channel](/1-go-core/1.3-concurrent/02-channel) | `01-go-core/concurrent/channel/` | 3h |
| 20 | sync 包 | [sync 包](/1-go-core/1.3-concurrent/03-sync) | `01-go-core/concurrent/sync/` | 3h |
| 21 | context 包 | [context 包](/1-go-core/1.3-concurrent/04-context) | `01-go-core/concurrent/context/` | 2h |
| 22 | errgroup | [errgroup](/1-go-core/1.3-concurrent/08-errgroup) | `01-go-core/concurrent/errgroup/` | 2h |

**🏁 里程碑检查点 3：**
- [ ] 理解 goroutine 的创建和调度
- [ ] 能使用 channel 实现 goroutine 间通信
- [ ] 理解 WaitGroup、Mutex 的使用场景
- [ ] 能使用 context 控制 goroutine 生命周期

### 第四阶段：Web 开发入门（第 5～6 周）

> 目标：能用 Gin 框架开发简单的 REST API

| 步骤 | 知识点 | 文档链接 | 代码示例 | 建议时间 |
|------|--------|---------|---------|---------|
| 23 | net/http 标准库 | [net/http 标准库](/2-web-data/2.1-web-framework/01-net-http) | `02-web-data/web-framework/net-http-server/` | 3h |
| 24 | Gin 框架 | [Gin 框架](/2-web-data/2.1-web-framework/03-gin) | `02-web-data/web-framework/gin-rest-api/` | 4h |
| 25 | database/sql | [database/sql](/2-web-data/2.2-database/01-database-sql) | `02-web-data/database/database-sql/` | 3h |
| 26 | GORM | [GORM](/2-web-data/2.2-database/02-gorm) | `02-web-data/database/gorm-examples/` | 4h |

**🏁 里程碑检查点 4（初学者路径完成）：**
- [ ] 能用 Gin 框架开发包含 CRUD 的 REST API
- [ ] 能使用 GORM 操作数据库
- [ ] 理解 HTTP 中间件的概念
- [ ] 具备独立开发简单 Go Web 项目的能力

## 下一步

完成初学者路径后，建议进入 [Go 中级进阶路径](/learning-paths/intermediate)，深入学习 Go 进阶特性、设计模式和更多中间件。
