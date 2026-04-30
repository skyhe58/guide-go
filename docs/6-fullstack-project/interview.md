---
title: "项目面试指南"
module: "fullstack-project"
difficulty: "advanced"
interviewFrequency: "high"
tags:
  - 面试
  - 架构设计
  - 项目经验
codeExample: "06-fullstack-project/goblog/"
estimatedTime: "2h"
---

# GoBlog 项目面试指南

## 概念说明

本指南帮助你在面试中清晰地介绍 GoBlog 项目，涵盖架构设计思路、技术选型理由、性能优化方案和技术难点。

## 项目介绍模板

> "GoBlog 是一个多租户博客平台的 REST API 后端服务，使用 Go 语言开发。技术栈包括 Gin + GORM + PostgreSQL + Redis + JWT，采用分层架构（Handler → Service → Repository），实现了完整的用户认证、RBAC 权限控制、多层缓存策略和 Docker 容器化部署。"

## 架构设计思路

### Q1: 为什么采用分层架构？

**难度**：⭐⭐ | **频率**：🔥🔥🔥

**答题思路**：

1. **职责分离**：Handler 负责 HTTP 协议处理，Service 负责业务逻辑，Repository 负责数据访问
2. **可测试性**：通过接口抽象各层依赖，Service 层可以 Mock Repository 进行单元测试
3. **可维护性**：修改数据库操作不影响业务逻辑，修改 HTTP 框架不影响核心业务
4. **团队协作**：不同开发者可以并行开发不同层级

**深入追问**：
- 如果项目规模更大，你会如何调整架构？（引入领域驱动设计 DDD）
- Handler 层和 Service 层的边界在哪里？（Handler 只做参数绑定和响应格式化）

### Q2: 为什么选择 PostgreSQL 而不是 MySQL？

**难度**：⭐⭐ | **频率**：🔥🔥

**答题思路**：

1. JSON/JSONB 原生支持，适合灵活的数据结构
2. CTE 和窗口函数更强大
3. MVCC 实现更优雅
4. 云原生生态中更受欢迎
5. 与 GORM 集成良好

### Q3: 依赖注入是如何实现的？

**难度**：⭐⭐⭐ | **频率**：🔥🔥

**答题思路**：

1. 采用 Google Wire 的设计思想，手动实现编译时依赖注入
2. 按层级定义 Provider：Infrastructure → Repository → Service → Handler
3. 所有依赖通过构造函数注入，不使用全局变量
4. 类型安全，编译期发现依赖错误

## 技术选型理由

### Q4: JWT 双令牌机制的设计考虑？

**难度**：⭐⭐⭐ | **频率**：🔥🔥🔥

**答题思路**：

1. **Access Token 短有效期（15 分钟）**：减少 Token 泄露的风险窗口
2. **Refresh Token 长有效期（7 天）**：避免用户频繁登录
3. **Token 黑名单（Redis）**：解决 JWT 无法主动注销的问题
4. **JTI（JWT ID）**：每个 Token 唯一标识，用于黑名单匹配

**深入追问**：
- Token 黑名单会不会导致 Redis 内存膨胀？（TTL 自动过期，不会无限增长）
- 如果 Redis 宕机，Token 黑名单失效怎么办？（降级策略：允许短时间内已注销 Token 通过）

### Q5: 中间件链的设计顺序有什么讲究？

**难度**：⭐⭐ | **频率**：🔥🔥

**答题思路**：

1. CORS 最先执行，处理预检请求
2. Request ID 在 Logger 之前，确保日志包含请求 ID
3. Recovery 在业务逻辑之前，捕获所有 panic
4. Rate Limiter 在认证之前，防止未认证请求消耗资源
5. JWT Auth 和 RBAC 按需挂载到特定路由组

## 性能优化

### Q6: 缓存策略是如何设计的？

**难度**：⭐⭐⭐ | **频率**：🔥🔥🔥

**答题思路**：

1. **Cache-Aside 模式**：读时查缓存，未命中查数据库并回填缓存
2. **写时删除**：更新文章时删除缓存，而不是更新缓存（避免并发写入不一致）
3. **空值缓存**：防止缓存穿透，TTL 5 分钟
4. **singleflight**：防止缓存击穿，同一 Key 只允许一个请求查数据库
5. **热门排行榜**：使用 Redis Sorted Set，按浏览量排序

### Q7: 如何保证接口的高可用？

**难度**：⭐⭐⭐ | **频率**：🔥🔥

**答题思路**：

1. **优雅启停**：signal.NotifyContext 监听系统信号，等待进行中的请求完成
2. **令牌桶限流**：防止突发流量打垮服务
3. **Panic Recovery**：捕获异常，防止单个请求导致服务崩溃
4. **健康检查**：/healthz 端点供负载均衡器探测
5. **连接池配置**：数据库和 Redis 连接池合理配置

## 技术难点

### Q8: 遇到过什么技术难点？如何解决的？

**难度**：⭐⭐⭐ | **频率**：🔥🔥🔥

**参考回答**：

**难点 1：缓存与数据库双写一致性**
- 问题：更新文章后，缓存和数据库数据不一致
- 方案：采用"先更新数据库，再删除缓存"策略，配合 TTL 兜底

**难点 2：JWT Token 注销**
- 问题：JWT 是无状态的，签发后无法主动失效
- 方案：使用 Redis 维护 Token 黑名单，每次请求校验

**难点 3：并发请求下的缓存击穿**
- 问题：热点文章缓存过期瞬间，大量请求同时查数据库
- 方案：使用 singleflight 确保同一 Key 只有一个请求查数据库

### Q9: 如果让你重新设计，你会做哪些改进？

**难度**：⭐⭐⭐ | **频率**：🔥🔥

**参考回答**：

1. 引入消息队列（如 NATS）处理异步任务（邮件通知、日志采集）
2. 使用 Elasticsearch 替代数据库模糊搜索
3. 引入 OpenTelemetry 实现分布式链路追踪
4. 使用 Casbin 替代自实现 RBAC，支持更灵活的权限模型
5. 添加 API 版本管理和向后兼容策略

## 项目亮点总结

面试时可以重点强调以下亮点：

1. **完整的分层架构**：不是简单的 CRUD demo，有清晰的架构设计
2. **生产级别的安全设计**：JWT 双令牌 + Token 黑名单 + RBAC
3. **多层缓存策略**：Cache-Aside + 空值缓存 + singleflight
4. **可观测性**：结构化日志 + Prometheus 指标
5. **容器化部署**：多阶段构建，scratch 镜像 ≤ 20MB
6. **代码质量**：golangci-lint + 单元测试覆盖率 ≥ 60%

## 参考资料

- [系统设计面试指南](https://github.com/donnemartin/system-design-primer)
- [Go 项目最佳实践](https://go.dev/doc/effective_go)
