---
title: "消息队列"
module: "message-queue"
difficulty: "intermediate"
tags:
  - Kafka
  - NATS
  - RabbitMQ
  - MQTT
  - 消息队列
  - 异步处理
  - 系统解耦
---

# 消息队列

> **前置依赖：** [Go 基础语法](/1-go-core/1.1-go-basics/)、[并发编程](/1-go-core/1.3-concurrent/)

## 模块概述

消息队列是分布式系统中实现**异步处理**、**系统解耦**和**流量削峰**的核心组件。在 Go 生态中，消息队列的应用场景极为广泛——从互联网后端的订单处理、日志采集，到 IoT 领域的设备通信、边缘计算，不同场景对消息队列的吞吐量、延迟、可靠性有不同要求。

本模块深入讲解四种主流消息队列的核心原理和 Go 客户端使用：

- **Kafka**：高吞吐量分布式流处理平台，适合日志采集、事件溯源、大数据管道
- **NATS**：Go 原生轻量级消息系统，适合微服务通信、云原生场景
- **RabbitMQ**：功能丰富的传统消息队列，适合企业级异步任务、复杂路由
- **MQTT**：轻量级 IoT 消息协议，适合物联网设备通信、边缘计算

## 知识点索引

### 消息队列核心

| 序号 | 知识点 | 难度 | 面试频率 | 预计时间 |
|------|--------|------|---------|---------|
| 01 | [Kafka 架构原理与 Go 客户端](./01-kafka.md) | ⭐⭐⭐ | 🔥🔥🔥 | 60min |
| 02 | [NATS 轻量级消息系统](./02-nats.md) | ⭐⭐ | 🔥🔥 | 45min |
| 03 | [RabbitMQ 核心概念与 Go 客户端](./03-rabbitmq.md) | ⭐⭐⭐ | 🔥🔥🔥 | 55min |
| 04 | [MQTT 物联网消息协议](./04-mqtt.md) | ⭐⭐⭐ | 🔥🔥 | 50min |
| 05 | [消息队列选型对比](./05-comparison.md) | ⭐⭐⭐ | 🔥🔥🔥 | 40min |

### 面试指南

| 📝 | [面试指南](./interview.md) | - | 🔥🔥🔥 | 60min |
|------|--------|------|---------|---------|

## 代码示例

> 💻 完整可运行代码：[code-examples/02-web-data/message-queue/](https://github.com/your-repo/code-examples/02-web-data/message-queue/)

| 示例目录 | 对应知识点 | 运行方式 | Demo 模式 |
|---------|-----------|---------|----------|
| `kafka/` | sarama 生产者/消费者 | `go run ./kafka/` / `go run ./kafka/ real` | 混合 |
| `nats/` | nats.go Core + JetStream | `go run ./nats/` / `go run ./nats/ real` | 混合 |
| `rabbitmq/` | amqp091-go 生产者/消费者 | `go run ./rabbitmq/` / `go run ./rabbitmq/ real` | 混合 |
| `mqtt/` | paho.mqtt.golang 发布/订阅 | `go run ./mqtt/` / `go run ./mqtt/ real` | 混合 |

## 前置条件

- 已完成 [Go 基础语法](/1-go-core/1.1-go-basics/) 模块
- 已完成 [并发编程](/1-go-core/1.3-concurrent/) 模块
- Part B 需要 Docker：
  - 全部启动：`docker compose -f docker/docker-compose.mq.yml up -d`
  - Kafka：`docker compose -f docker/docker-compose.mq.yml up -d kafka`
  - NATS：`docker compose -f docker/docker-compose.mq.yml up -d nats`
  - RabbitMQ：`docker compose -f docker/docker-compose.mq.yml up -d rabbitmq`
  - EMQX：`docker compose -f docker/docker-compose.mq.yml up -d emqx`
