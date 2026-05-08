---
title: "网络编程与 Web 框架"
module: "web-framework"
difficulty: "intermediate"
tags:
  - 网络编程
  - Web 框架
  - net/http
  - Gin
  - gRPC
  - WebSocket
---

# 网络编程与 Web 框架

> **前置依赖：** [Go 基础语法](/1-go-core/1.1-go-basics/)、[并发编程](/1-go-core/1.3-concurrent/)

## 模块概述

网络编程是 Go 语言最核心的应用场景之一。Go 标准库 `net/http` 本身就是一个生产级的 HTTP 框架，许多公司直接使用标准库构建高性能 Web 服务。在此基础上，Gin 以其高性能和简洁 API 成为 Go 社区使用率最高的 Web 框架（2025 年使用率达 48%）。gRPC 则是微服务间通信的事实标准。

本模块从标准库出发，逐步深入到主流框架和协议，体现 Go "标准库优先"的哲学。

## 知识点索引

| 序号 | 知识点 | 难度 | 面试频率 | 预计时间 |
|------|--------|------|---------|---------|
| 01 | [net/http 标准库](./01-net-http.md) | ⭐⭐⭐ | 🔥🔥🔥 | 60min |
| 02 | [TCP 编程](./02-tcp.md) | ⭐⭐⭐ | 🔥🔥 | 45min |
| 03 | [Gin 框架](./03-gin.md) | ⭐⭐ | 🔥🔥🔥 | 60min |
| 04 | [gRPC](./04-grpc.md) | ⭐⭐⭐ | 🔥🔥🔥 | 60min |
| 05 | [WebSocket](./05-websocket.md) | ⭐⭐ | 🔥🔥 | 40min |
| 06 | [框架选型对比](./06-comparison.md) | ⭐⭐ | 🔥🔥🔥 | 30min |
| 📝 | [面试指南](./interview.md) | - | 🔥🔥🔥 | 60min |

## 代码示例

> 💻 完整可运行代码：[code-examples/02-web-data/web-framework/](https://github.com/skyhe58/guide-go/tree/main/code-examples/02-web-data/web-framework/)

| 示例目录 | 对应知识点 | 运行方式 | Demo 模式 |
|---------|-----------|---------|----------|
| `net-http-server/` | net/http 标准库 HTTP 服务器 | `go run main.go` | Part A |
| `gin-rest-api/` | Gin REST API 完整示例 | `go run main.go` | Part A |
| `grpc-examples/` | gRPC 四种通信模式 | `go run main.go` | Part A |
| `websocket/` | WebSocket 示例 | `go run main.go` | Part A |

## 学习建议

1. **先学标准库**：`net/http` 是 Go 网络编程的基石，理解 Handler 接口和 ServeMux 后再学框架
2. **Gin 是主线**：Gin 是国内 Go 开发最常用的框架，面试必考
3. **gRPC 是进阶**：微服务架构中 gRPC 是服务间通信的首选，理解四种通信模式
4. **关注面试**：标注 🔥🔥🔥 的知识点是面试高频考点，需重点掌握

## 前置条件

- 已完成 [Go 基础语法](/1-go-core/1.1-go-basics/) 模块
- 已完成 [并发编程](/1-go-core/1.3-concurrent/) 模块
- 了解 HTTP 协议基础（请求方法、状态码、Header）
