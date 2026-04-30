---
title: "GoBlog 全栈实战项目"
module: "fullstack-project"
difficulty: "advanced"
tags:
  - 全栈实战
  - Gin
  - GORM
  - PostgreSQL
  - Redis
  - JWT
  - Docker
codeExample: "06-fullstack-project/goblog/"
estimatedTime: "40h"
---

# GoBlog 全栈实战项目

## 项目概述

GoBlog 是一个**多租户博客平台 REST API 后端服务**，作为知识库的"毕业设计"级别实战项目。项目串联知识库第一层（语言核心）和第二层（Web 开发与数据）的核心知识点，让学习者在真实业务场景中综合运用所学技能。

**项目特色：**

- 🏗️ **标准项目布局**：cmd/internal/pkg/configs/migrations 分层清晰
- 🔐 **完整认证鉴权**：JWT 双令牌 + RBAC 三角色权限控制
- 📦 **多层缓存策略**：Redis 文章缓存 + 热门排行榜 + Token 黑名单
- 📝 **结构化日志**：zerolog 请求日志 + 错误日志 + SQL 慢查询日志
- 📊 **Prometheus 监控**：HTTP 请求计数、延迟直方图、活跃连接数
- 🐳 **Docker 一键部署**：多阶段构建，scratch 镜像 ≤ 20MB
- 🔌 **Wire 依赖注入**：编译时依赖注入，类型安全

## 整体架构

```mermaid
graph TB
    subgraph "客户端"
        Client[HTTP Client / Swagger UI]
    end

    subgraph "GoBlog 服务"
        subgraph "中间件链"
            MW1[CORS] --> MW2[Request ID]
            MW2 --> MW3[Logger]
            MW3 --> MW4[Recovery]
            MW4 --> MW5[Rate Limiter]
            MW5 --> MW6[JWT Auth]
            MW6 --> MW7[RBAC]
        end

        subgraph "Handler 层"
            H1[UserHandler]
            H2[ArticleHandler]
            H3[TagHandler]
            H4[CommentHandler]
            H5[AdminHandler]
        end

        subgraph "Service 层"
            S1[UserService]
            S2[ArticleService]
            S3[TagService]
            S4[CommentService]
            S5[AdminService]
        end

        subgraph "Repository 层"
            R1[UserRepo]
            R2[ArticleRepo]
            R3[TagRepo]
            R4[CommentRepo]
        end
    end

    subgraph "外部依赖"
        PG[(PostgreSQL)]
        Redis[(Redis)]
    end

    Client --> MW1
    MW7 --> H1 & H2 & H3 & H4 & H5
    H1 --> S1
    H2 --> S2
    H3 --> S3
    H4 --> S4
    H5 --> S5
    S1 & S2 & S3 & S4 --> R1 & R2 & R3 & R4
    S2 --> Redis
    R1 & R2 & R3 & R4 --> PG
```

## 知识点串联地图

GoBlog 项目串联了知识库中以下模块的核心知识点：

```mermaid
graph LR
    subgraph "第一层：语言核心"
        A1[1.1 Go 基础语法<br/>错误处理/结构体/接口]
        A2[1.3 并发编程<br/>context/优雅启停]
        A3[1.6 设计模式<br/>Wire DI/中间件模式]
    end

    subgraph "第二层：Web 开发与数据"
        B1[2.1 Web 框架<br/>Gin 路由/中间件/验证]
        B2[2.2 数据库<br/>GORM/PostgreSQL/迁移]
        B3[2.3 缓存<br/>Redis/缓存策略]
        B4[2.6 认证鉴权<br/>JWT/RBAC/bcrypt]
        B5[2.7 可观测性<br/>zerolog/Prometheus]
    end

    subgraph "第三层：微服务"
        C1[3.2 服务治理<br/>Viper 配置管理]
        C2[3.3 容器化<br/>Docker 多阶段构建]
    end

    subgraph "GoBlog 项目"
        GB[GoBlog<br/>多租户博客平台]
    end

    A1 --> GB
    A2 --> GB
    A3 --> GB
    B1 --> GB
    B2 --> GB
    B3 --> GB
    B4 --> GB
    B5 --> GB
    C1 --> GB
    C2 --> GB
```

| 技术栈 | 用途 | 对应知识库模块 |
|--------|------|---------------|
| Gin | HTTP 路由与中间件 | [2.1 网络编程与 Web 框架](/2-web-data/2.1-web-framework/) |
| GORM + PostgreSQL | ORM 数据持久化 | [2.2 数据库与 ORM](/2-web-data/2.2-database/) |
| go-redis | 缓存与会话管理 | [2.3 缓存与搜索](/2-web-data/2.3-cache-search/) |
| golang-jwt | JWT 双令牌认证 | [2.6 认证鉴权](/2-web-data/2.6-auth/) |
| RBAC | 角色权限控制 | [2.6 认证鉴权](/2-web-data/2.6-auth/) |
| zerolog | 结构化日志 | [2.7 日志与可观测性](/2-web-data/2.7-observability/) |
| Prometheus | 指标监控 | [2.7 日志与可观测性](/2-web-data/2.7-observability/) |
| Wire | 编译时依赖注入 | [1.6 设计模式与工程化](/1-go-core/1.6-patterns/) |
| golang-migrate | 数据库迁移 | [2.2 数据库与 ORM](/2-web-data/2.2-database/) |
| Viper | 配置管理 | [3.2 服务治理](/3-microservice/3.2-service-governance/) |
| Docker | 容器化部署 | [3.3 容器化与 K8s](/3-microservice/3.3-docker-k8s/) |
| Makefile | 构建自动化 | [1.6 设计模式与工程化](/1-go-core/1.6-patterns/) |

## 项目目录结构

```
goblog/
├── cmd/goblog/main.go          # 程序入口（优雅启停）
├── configs/config.yaml         # Viper 配置文件
├── Dockerfile                  # 多阶段构建
├── docker-compose.yml          # 一键启动
├── Makefile                    # 构建自动化
├── internal/                   # 私有应用代码
│   ├── auth/                   # JWT + bcrypt + RBAC
│   ├── cache/                  # Redis 缓存层
│   ├── config/                 # Viper 配置加载
│   ├── database/               # GORM 数据库连接
│   ├── errcode/                # 业务错误码
│   ├── handler/                # HTTP Handler 层
│   ├── middleware/             # Gin 中间件链
│   ├── model/                  # GORM 数据模型
│   ├── repository/             # 数据访问层
│   ├── router/                 # 路由注册
│   ├── service/                # 业务逻辑层
│   └── wire/                   # 依赖注入
├── pkg/                        # 公共包
│   ├── pagination/             # 分页工具
│   └── response/               # 统一响应格式
├── migrations/                 # 数据库迁移文件
└── scripts/seed.go             # 种子数据脚本
```

## 快速开始

```bash
# 1. 启动依赖中间件
cd code-examples/06-fullstack-project/goblog
docker compose up -d postgres redis

# 2. 执行数据库迁移
make migrate-up

# 3. 初始化种子数据
go run scripts/seed.go

# 4. 启动服务
make run

# 5. 测试 API
curl http://localhost:8080/healthz
```

## 分步实现指南

建议按以下顺序学习和实现：

1. [技术选型说明](./01-tech-stack.md)
2. [项目初始化与环境搭建](./02-project-setup.md)
3. [用户模块实现](./03-user-module.md)
4. [文章模块实现](./04-article-module.md)
5. [标签与评论模块](./05-tag-comment-module.md)
6. [中间件链实现](./06-middleware.md)
7. [缓存策略实现](./07-cache-strategy.md)
8. [测试策略与实现](./08-testing.md)
9. [Docker 部署](./09-deployment.md)

## 参考资料

- [Gin 官方文档](https://gin-gonic.com/docs/)
- [GORM 官方文档](https://gorm.io/docs/)
- [go-redis 文档](https://redis.uptrace.dev/)
- [golang-jwt 文档](https://golang-jwt.github.io/jwt/)
