---
title: "缓存与搜索"
module: "cache-search"
difficulty: "intermediate"
tags:
  - Redis
  - Elasticsearch
  - 缓存
  - 搜索
  - go-redis
  - go-elasticsearch
---

# 缓存与搜索

> **前置依赖：** [Go 基础语法](/1-go-core/1.1-go-basics/)

## 模块概述

缓存和搜索是后端系统中提升性能和用户体验的两大核心组件。Redis 作为最流行的内存数据库，广泛用于缓存、分布式锁、消息队列等场景；Elasticsearch 作为分布式搜索引擎，是全文搜索和日志分析的事实标准。

本模块深入讲解 Redis 和 Elasticsearch 的核心原理，并通过 `go-redis` 和 `go-elasticsearch` 客户端展示在 Go 项目中的实际使用。

## 知识点索引

### Redis 部分

| 序号 | 知识点 | 难度 | 面试频率 | 预计时间 |
|------|--------|------|---------|---------|
| 01 | [Redis 数据结构与底层实现](./01-redis-data-structures.md) | ⭐⭐⭐ | 🔥🔥🔥 | 50min |
| 02 | [Redis 持久化（RDB/AOF）](./02-redis-persistence.md) | ⭐⭐⭐ | 🔥🔥🔥 | 40min |
| 03 | [Redis 主从与哨兵](./03-redis-replication.md) | ⭐⭐⭐ | 🔥🔥🔥 | 40min |
| 04 | [Redis Cluster 集群](./04-redis-cluster.md) | ⭐⭐⭐ | 🔥🔥🔥 | 45min |
| 05 | [缓存穿透/击穿/雪崩方案](./05-redis-cache-problems.md) | ⭐⭐⭐ | 🔥🔥🔥 | 45min |
| 06 | [分布式锁（Redlock/单节点锁）](./06-redis-distributed-lock.md) | ⭐⭐⭐ | 🔥🔥🔥 | 40min |
| 07 | [go-redis 客户端](./07-redis-go-client.md) | ⭐⭐ | 🔥🔥🔥 | 50min |

### Elasticsearch 部分

| 序号 | 知识点 | 难度 | 面试频率 | 预计时间 |
|------|--------|------|---------|---------|
| 08 | [Elasticsearch 倒排索引原理](./08-es-inverted-index.md) | ⭐⭐⭐ | 🔥🔥🔥 | 40min |
| 09 | [ES 映射与分析器](./09-es-mapping.md) | ⭐⭐ | 🔥🔥 | 35min |
| 10 | [ES CRUD 操作](./10-es-crud.md) | ⭐⭐ | 🔥🔥 | 30min |
| 11 | [ES DSL 查询](./11-es-dsl.md) | ⭐⭐⭐ | 🔥🔥🔥 | 45min |
| 12 | [ES 聚合分析](./12-es-aggregation.md) | ⭐⭐⭐ | 🔥🔥 | 40min |
| 13 | [go-elasticsearch 客户端使用](./13-es-go-client.md) | ⭐⭐ | 🔥🔥 | 40min |

### 面试指南

| 📝 | [面试指南](./interview.md) | - | 🔥🔥🔥 | 60min |
|------|--------|------|---------|---------|

## 代码示例

> 💻 完整可运行代码：[code-examples/02-web-data/cache-search/](https://github.com/skyhe58/guide-go/tree/main/code-examples/02-web-data/cache-search/)

| 示例目录 | 对应知识点 | 运行方式 | Demo 模式 |
|---------|-----------|---------|----------|
| `redis/` | go-redis 完整示例（数据结构/Pipeline/分布式锁） | `go run main.go` / `go run main.go real` | 混合 |
| `elasticsearch/` | go-elasticsearch 完整示例 | `go run main.go` / `go run main.go real` | 混合 |

## 前置条件

- 已完成 [Go 基础语法](/1-go-core/1.1-go-basics/) 模块
- Part B 需要 Docker：
  - Redis：`docker compose -f docker/docker-compose.yml up -d redis`
  - Elasticsearch：`docker compose -f docker/docker-compose.es.yml up -d`
