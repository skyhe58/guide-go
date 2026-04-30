---
title: "技术选型说明"
module: "fullstack-project"
difficulty: "intermediate"
tags:
  - 技术选型
  - 架构设计
codeExample: "06-fullstack-project/goblog/"
estimatedTime: "30min"
---

# 技术选型说明

## 概念说明

GoBlog 的技术选型遵循 Go 社区的最佳实践，每个技术组件都对应知识库中的具体模块，方便学习者在实战中回顾和深化理论知识。

## 技术栈总览

| 类别 | 技术 | 版本 | 选型理由 | 知识库模块 |
|------|------|------|---------|-----------|
| Web 框架 | Gin | v1.10 | 使用率 48%（2025），性能优秀，生态成熟 | [2.1 Web 框架](/2-web-data/2.1-web-framework/03-gin) |
| ORM | GORM | v1.25 | Go 生态最流行的 ORM，功能全面 | [2.2 数据库](/2-web-data/2.2-database/02-gorm) |
| 数据库 | PostgreSQL | 16 | 功能强大，JSON 支持好，开源免费 | [2.2 数据库](/2-web-data/2.2-database/09-postgresql) |
| 缓存 | Redis | 7 | 高性能内存数据库，数据结构丰富 | [2.3 缓存](/2-web-data/2.3-cache-search/07-redis-go-client) |
| 认证 | golang-jwt | v5 | JWT 标准实现，社区活跃 | [2.6 认证鉴权](/2-web-data/2.6-auth/01-jwt) |
| 权限 | 自实现 RBAC | — | 简单三角色模型，无需引入 Casbin | [2.6 认证鉴权](/2-web-data/2.6-auth/04-rbac) |
| 日志 | zerolog | v1.33 | 零分配高性能，JSON 输出 | [2.7 可观测性](/2-web-data/2.7-observability/02-zerolog) |
| 监控 | Prometheus | v1.20 | 云原生监控标准 | [2.7 可观测性](/2-web-data/2.7-observability/08-prometheus) |
| 配置 | Viper | v1.19 | 多格式支持，环境变量覆盖 | [3.2 服务治理](/3-microservice/3.2-service-governance/04-viper) |
| DI | Wire（手动模拟） | — | 编译时依赖注入，类型安全 | [1.6 设计模式](/1-go-core/1.6-patterns/08-wire) |
| 迁移 | golang-migrate | — | 版本化数据库迁移 | [2.2 数据库](/2-web-data/2.2-database/04-migration) |
| 验证 | validator | v10 | Gin 内置集成，标签式验证 | [2.1 Web 框架](/2-web-data/2.1-web-framework/03-gin) |
| 限流 | x/time/rate | — | 标准库扩展，令牌桶算法 | [4.1 分布式系统](/4-distributed/4.1-distributed/06-rate-limiting) |
| 容器化 | Docker | — | 多阶段构建，scratch 镜像 | [3.3 容器化](/3-microservice/3.3-docker-k8s/02-dockerfile) |

## 核心设计决策

### 为什么选 Gin 而不是标准库 net/http？

Gin 在 Go 1.22 增强 net/http 路由后仍然是最佳选择：
- 中间件链机制成熟，开箱即用
- 参数绑定和验证集成 validator
- 路由分组和版本管理方便
- Swagger 集成（swaggo）生态完善
- 性能与 net/http 接近，开发效率更高

### 为什么选 PostgreSQL 而不是 MySQL？

- JSON/JSONB 原生支持，适合灵活的数据结构
- 数组类型支持
- CTE（公用表表达式）和窗口函数更强大
- MVCC 实现更优雅（不需要 undo log）
- 云原生生态中更受欢迎

### 为什么选 zerolog 而不是 zap？

- 零内存分配，性能更优
- 链式 API 更符合 Go 风格
- JSON 输出天然适合日志聚合
- 与 Gin 集成简单

### 为什么手动模拟 Wire 而不是直接用 Wire？

- 降低学习门槛，不需要安装 Wire CLI
- 代码更直观，便于理解依赖注入原理
- 项目规模适中，手动注入可维护
- 保留了 Wire 的分层思想（Provider Sets）

## 分层架构说明

```
Handler 层（HTTP 接口）
    ↓ 调用
Service 层（业务逻辑）
    ↓ 调用
Repository 层（数据访问）
    ↓ 操作
Database / Cache（存储层）
```

- **Handler**：参数绑定、验证、调用 Service、返回响应
- **Service**：业务逻辑、缓存策略、事务管理
- **Repository**：SQL 查询封装、GORM 操作
- **Model**：GORM 数据模型定义

## 参考资料

- [Gin 框架 GitHub](https://github.com/gin-gonic/gin)
- [GORM 官方文档](https://gorm.io/)
- [PostgreSQL 官方文档](https://www.postgresql.org/docs/)
