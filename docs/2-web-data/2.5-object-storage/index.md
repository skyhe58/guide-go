---
title: "对象存储与文档数据库"
module: "object-storage"
difficulty: "intermediate"
tags:
  - MinIO
  - MongoDB
  - 对象存储
  - 文档数据库
  - S3
  - BSON
---

# 对象存储与文档数据库

> **前置依赖：** [Go 基础语法](/1-go-core/1.1-go-basics/)

## 模块概述

在现代后端开发中，**对象存储**和**文档数据库**是两种与传统关系型数据库互补的数据存储方案。对象存储用于管理图片、视频、文档等非结构化文件；文档数据库用于存储结构灵活的 JSON/BSON 文档数据。

本模块深入讲解两种核心技术在 Go 项目中的使用：

- **MinIO**：兼容 S3 API 的高性能对象存储，适合私有化部署的文件存储场景。minio-go SDK 提供上传、下载、预签名 URL、分片上传、Bucket 策略等完整功能
- **MongoDB**：最流行的文档数据库，以 BSON 格式存储灵活的文档数据。mongo-go-driver 是官方 Go 驱动，支持 CRUD、聚合管道、索引管理、事务等操作

### 对象存储 vs 文件系统 vs 关系型数据库

| 维度 | 对象存储（MinIO/S3） | 文件系统 | 关系型数据库（BLOB） |
|------|---------------------|---------|-------------------|
| 适用数据 | 图片/视频/文档等大文件 | 任意文件 | 小型二进制数据 |
| 扩展性 | 水平扩展，PB 级 | 受限于单机磁盘 | 不适合大文件 |
| 访问方式 | HTTP REST API（S3 兼容） | 文件路径 | SQL 查询 |
| 元数据 | 自定义元数据标签 | 文件属性 | 表字段 |
| CDN 集成 | 天然支持 | 需额外配置 | 不支持 |

### 文档数据库 vs 关系型数据库

| 维度 | MongoDB（文档数据库） | MySQL/PostgreSQL（关系型） |
|------|---------------------|--------------------------|
| 数据模型 | 灵活的 JSON/BSON 文档 | 固定 Schema 的表结构 |
| 查询语言 | MongoDB 查询语法 | SQL |
| 事务支持 | 4.0+ 支持多文档事务 | 完整 ACID 事务 |
| 扩展方式 | 原生分片（Sharding） | 主从复制 + 分库分表 |
| 适用场景 | 内容管理/日志/IoT 数据 | 金融/订单/强一致性场景 |

## 知识点索引

### 对象存储与文档数据库

| 序号 | 知识点 | 难度 | 面试频率 | 预计时间 |
|------|--------|------|---------|---------|
| 01 | [MinIO 对象存储与 minio-go SDK](./01-minio.md) | ⭐⭐⭐ | 🔥🔥 | 50min |
| 02 | [MongoDB 文档数据库与 mongo-go-driver](./02-mongodb.md) | ⭐⭐⭐ | 🔥🔥🔥 | 60min |

### 面试指南

| 📝 | [面试指南](./interview.md) | - | 🔥🔥🔥 | 45min |
|------|--------|------|---------|---------|

## 代码示例

> 💻 完整可运行代码：[code-examples/02-web-data/object-storage/](https://github.com/skyhe58/guide-go/tree/main/code-examples/02-web-data/object-storage/)

| 示例目录 | 对应知识点 | 运行方式 | Demo 模式 |
|---------|-----------|---------|----------|
| `minio/` | minio-go 完整示例 | `go run ./minio/` / `go run ./minio/ real` | 混合 |
| `mongodb/` | mongo-go-driver 完整示例 | `go run ./mongodb/` / `go run ./mongodb/ real` | 混合 |

## 前置条件

- 已完成 [Go 基础语法](/1-go-core/1.1-go-basics/) 模块
- Part B 需要 Docker：
  - 全部启动：`docker compose -f docker/docker-compose.yml up -d minio mongodb`
  - MinIO：`docker compose -f docker/docker-compose.yml up -d minio`
  - MongoDB：`docker compose -f docker/docker-compose.yml up -d mongodb`
