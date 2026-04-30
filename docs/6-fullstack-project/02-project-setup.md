---
title: "项目初始化与环境搭建"
module: "fullstack-project"
difficulty: "beginner"
tags:
  - 项目初始化
  - Go Module
  - Docker
codeExample: "06-fullstack-project/goblog/"
estimatedTime: "1h"
---

# 项目初始化与环境搭建

## 概念说明

本节介绍 GoBlog 项目的初始化流程，包括环境准备、项目结构创建、依赖安装和基础配置。

## 环境要求

| 工具 | 版本 | 用途 |
|------|------|------|
| Go | 1.22+ | 编译和运行 |
| Docker | 24+ | 运行 PostgreSQL 和 Redis |
| Docker Compose | v2+ | 编排容器服务 |
| Make | — | 构建自动化 |

## 核心步骤

### 1. 创建项目目录

```bash
mkdir -p goblog/{cmd/goblog,configs,internal,pkg,migrations,scripts}
cd goblog
```

### 2. 初始化 Go Module

```bash
go mod init guide-go/goblog
```

### 3. 安装核心依赖

```bash
# Web 框架
go get github.com/gin-gonic/gin

# ORM
go get gorm.io/gorm
go get gorm.io/driver/postgres

# Redis
go get github.com/redis/go-redis/v9

# JWT
go get github.com/golang-jwt/jwt/v5

# 配置管理
go get github.com/spf13/viper

# 日志
go get github.com/rs/zerolog

# 密码加密
go get golang.org/x/crypto

# 限流
go get golang.org/x/time

# Prometheus
go get github.com/prometheus/client_golang

# UUID
go get github.com/google/uuid
```

### 4. 创建配置文件

`configs/config.yaml` 是 Viper 加载的默认配置文件，支持通过环境变量覆盖：

```yaml
server:
  port: 8080
  mode: debug
  read_timeout: 10s
  write_timeout: 10s
  shutdown_timeout: 30s

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres123
  dbname: goblog
  sslmode: disable

redis:
  addr: localhost:6379
  password: ""
  db: 0

jwt:
  secret: "your-secret-key-change-in-production"
  access_token_ttl: 15m
  refresh_token_ttl: 168h
  issuer: goblog
```

### 5. 启动依赖中间件

```bash
# 使用 Docker Compose 启动 PostgreSQL 和 Redis
docker compose up -d postgres redis
```

### 6. 创建数据库

PostgreSQL 容器启动后会自动创建 `goblog` 数据库（通过 `POSTGRES_DB` 环境变量）。

### 7. 编译验证

```bash
go build ./cmd/goblog/
```

## 项目布局说明

GoBlog 采用 Go 社区推荐的标准项目布局：

| 目录 | 说明 | 对应知识点 |
|------|------|-----------|
| `cmd/` | 程序入口 | main 函数、优雅启停 |
| `internal/` | 私有代码，不可被外部导入 | Go 包可见性规则 |
| `pkg/` | 可被外部导入的公共包 | 通用工具函数 |
| `configs/` | 配置文件 | Viper 配置管理 |
| `migrations/` | 数据库迁移文件 | golang-migrate |
| `scripts/` | 脚本工具 | 种子数据初始化 |

## 代码示例

> 💻 完整可运行代码：[code-examples/06-fullstack-project/goblog/](https://github.com/)

## 参考资料

- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Go Modules 参考](https://go.dev/ref/mod)
