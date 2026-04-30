---
title: "数据库与 ORM"
module: "database"
difficulty: "intermediate"
tags:
  - 数据库
  - ORM
  - GORM
  - sqlx
  - MySQL
  - PostgreSQL
  - database/sql
---

# 数据库与 ORM

> **前置依赖：** [Go 基础语法](/1-go-core/1.1-go-basics/)

## 模块概述

数据库操作是后端开发的核心技能。Go 标准库 `database/sql` 提供了统一的数据库访问接口，配合 GORM（最流行的 ORM）和 sqlx（轻量级扩展），可以满足从简单 CRUD 到复杂查询的各种需求。

本模块同时覆盖 MySQL 和 PostgreSQL 两大主流关系型数据库，深入讲解索引原理、事务隔离、锁机制等面试高频知识点。

## 知识点索引

| 序号 | 知识点 | 难度 | 面试频率 | 预计时间 |
|------|--------|------|---------|---------|
| 01 | [database/sql 标准库](./01-database-sql.md) | ⭐⭐ | 🔥🔥🔥 | 45min |
| 02 | [GORM](./02-gorm.md) | ⭐⭐ | 🔥🔥🔥 | 60min |
| 03 | [sqlx](./03-sqlx.md) | ⭐⭐ | 🔥🔥 | 35min |
| 04 | [数据库迁移](./04-migration.md) | ⭐⭐ | 🔥🔥 | 30min |
| 05 | [MySQL 索引原理](./05-mysql-index.md) | ⭐⭐⭐ | 🔥🔥🔥 | 50min |
| 06 | [MySQL 事务与隔离级别](./06-mysql-transaction.md) | ⭐⭐⭐ | 🔥🔥🔥 | 45min |
| 07 | [MySQL 锁机制](./07-mysql-lock.md) | ⭐⭐⭐ | 🔥🔥🔥 | 40min |
| 08 | [SQL 优化与 EXPLAIN](./08-mysql-optimization.md) | ⭐⭐⭐ | 🔥🔥🔥 | 40min |
| 09 | [PostgreSQL](./09-postgresql.md) | ⭐⭐⭐ | 🔥🔥 | 50min |
| 10 | [ORM 选型对比](./10-comparison.md) | ⭐⭐ | 🔥🔥🔥 | 25min |
| 📝 | [面试指南](./interview.md) | - | 🔥🔥🔥 | 60min |

## 代码示例

> 💻 完整可运行代码：[code-examples/02-web-data/database/](https://github.com/your-repo/code-examples/02-web-data/database/)

| 示例目录 | 对应知识点 | 运行方式 | Demo 模式 |
|---------|-----------|---------|----------|
| `database-sql/` | database/sql CRUD | `go run main.go` / `go run main.go real` | 混合 |
| `gorm-examples/` | GORM 完整示例 | `go run main.go` / `go run main.go real` | 混合 |
| `sqlx-examples/` | sqlx 示例 | `go run main.go` / `go run main.go real` | 混合 |
| `migration/` | 数据库迁移示例 | `go run main.go` | Part A |

## 前置条件

- 已完成 [Go 基础语法](/1-go-core/1.1-go-basics/) 模块
- Part B 需要 Docker：`docker compose -f docker/docker-compose.yml up -d mysql postgresql`
