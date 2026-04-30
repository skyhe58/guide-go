---
title: 按公司类型面试重点
description: 按大厂、中厂、创业公司分类的 Go 面试重点汇总
---

# 按公司类型面试重点

不同类型的公司对 Go 开发者的考察侧重点不同。本文按公司类型分类，帮助你有针对性地准备面试。

## 大厂（字节跳动、B 站、腾讯、阿里、美团）

### 面试特点

- **轮次多**：通常 3～4 轮技术面 + 1 轮 HR 面
- **深度优先**：每个知识点会追问到底层原理
- **算法必考**：每轮至少 1～2 道算法题
- **系统设计**：高级岗位必考系统设计题
- **项目深挖**：会深入追问项目中的技术决策和优化

### 面试重点

#### 初级岗位（校招/1～2 年经验）

| 优先级 | 知识点 | 难度 | 关联模块 |
|--------|--------|------|---------|
| P0 | Slice 扩容机制 | ⭐ | [数组与切片](/1-go-core/1.1-go-basics/09-slice) |
| P0 | defer 执行顺序 | ⭐ | [函数](/1-go-core/1.1-go-basics/06-functions) |
| P0 | Goroutine 和线程区别 | ⭐ | [goroutine](/1-go-core/1.3-concurrent/01-goroutine) |
| P0 | GMP 调度模型 | ⭐⭐ | [GMP 调度模型](/1-go-core/1.4-runtime/01-gmp) |
| P0 | GC 三色标记法 | ⭐⭐ | [垃圾回收](/1-go-core/1.4-runtime/02-gc) |
| P0 | Channel 死锁场景 | ⭐⭐ | [channel](/1-go-core/1.3-concurrent/02-channel) |
| P1 | 接口 nil 判断陷阱 | ⭐⭐ | [接口](/1-go-core/1.2-go-advanced/01-interfaces) |
| P1 | Context 使用 | ⭐⭐ | [context 包](/1-go-core/1.3-concurrent/04-context) |
| P1 | MySQL B+树索引 | ⭐⭐ | [MySQL 索引原理](/2-web-data/2.2-database/05-mysql-index) |
| P1 | Redis 缓存问题 | ⭐⭐ | [缓存问题](/2-web-data/2.3-cache-search/05-redis-cache-problems) |
| P2 | 算法：链表、树、动态规划 | — | [数据结构与算法](/1-go-core/1.7-algorithm/) |

#### 中高级岗位（3～5 年经验）

| 优先级 | 知识点 | 难度 | 关联模块 |
|--------|--------|------|---------|
| P0 | 内存管理与逃逸分析 | ⭐⭐⭐ | [内存管理](/1-go-core/1.4-runtime/03-memory) |
| P0 | 分布式锁方案对比 | ⭐⭐⭐ | [分布式锁](/4-distributed/4.1-distributed/03-distributed-lock) |
| P0 | 缓存一致性方案 | ⭐⭐⭐ | [缓存一致性](/4-distributed/4.2-architecture/04-cache-consistency) |
| P0 | 秒杀系统设计 | ⭐⭐⭐ | [秒杀系统](/4-distributed/4.2-architecture/01-seckill) |
| P0 | Raft 一致性算法 | ⭐⭐⭐ | [Raft](/4-distributed/4.1-distributed/02-raft) |
| P1 | 分布式事务方案 | ⭐⭐⭐ | [分布式事务](/4-distributed/4.1-distributed/04-distributed-transaction) |
| P1 | 限流和熔断 | ⭐⭐ | [限流算法](/4-distributed/4.1-distributed/06-rate-limiting) |
| P1 | 微服务框架选型 | ⭐⭐ | [框架选型](/3-microservice/3.1-microservice/04-comparison) |
| P2 | pprof 线上排查 | ⭐⭐ | [pprof](/1-go-core/1.4-runtime/05-pprof) |
| P2 | K8s 核心概念 | ⭐⭐ | [K8s 架构](/3-microservice/3.3-docker-k8s/05-k8s-architecture) |

### 大厂代表性追问链

**字节跳动风格**（深度追问）：
> Slice 扩容 → 底层数组共享 → 逃逸分析 → 内存分配器 → GC 调优

**B 站风格**（Kratos 生态）：
> 微服务框架 → Kratos 架构 → Wire 依赖注入 → gRPC 拦截器 → 服务治理

---

## 中厂（PingCAP、七牛云、声网、得物、快手）

### 面试特点

- **轮次适中**：通常 2～3 轮技术面
- **广度优先**：覆盖面广，但不会追问太深
- **实战导向**：更关注实际项目经验和解决问题的能力
- **中间件使用**：重点考察 Redis、Kafka、MySQL 的实际使用经验
- **算法适中**：通常 1 道中等难度算法题

### 面试重点

| 优先级 | 知识点 | 难度 | 关联模块 |
|--------|--------|------|---------|
| P0 | Goroutine 和 Channel | ⭐⭐ | [并发编程](/1-go-core/1.3-concurrent/) |
| P0 | GMP 调度模型（概念级） | ⭐⭐ | [GMP 调度模型](/1-go-core/1.4-runtime/01-gmp) |
| P0 | MySQL 索引和事务 | ⭐⭐ | [数据库与 ORM](/2-web-data/2.2-database/) |
| P0 | Redis 使用和缓存问题 | ⭐⭐ | [缓存与搜索](/2-web-data/2.3-cache-search/) |
| P0 | Gin 框架使用 | ⭐⭐ | [Gin 框架](/2-web-data/2.1-web-framework/03-gin) |
| P1 | JWT 认证 | ⭐⭐ | [JWT](/2-web-data/2.6-auth/01-jwt) |
| P1 | Docker 部署 | ⭐⭐ | [Docker](/3-microservice/3.3-docker-k8s/02-dockerfile) |
| P1 | 消息队列使用 | ⭐⭐ | [消息队列](/2-web-data/2.4-message-queue/) |
| P1 | 分布式锁 | ⭐⭐ | [分布式锁](/4-distributed/4.1-distributed/03-distributed-lock) |
| P2 | 日志和监控 | ⭐⭐ | [日志与可观测性](/2-web-data/2.7-observability/) |
| P2 | CI/CD 流程 | ⭐ | [CI/CD](/5-devops/5.1-cicd/) |

### 中厂典型面试题组合

**后端开发岗**：
1. Go 基础（slice/map/goroutine）— 15 分钟
2. MySQL 索引和事务 — 10 分钟
3. Redis 缓存方案 — 10 分钟
4. 项目经验深挖 — 15 分钟
5. 算法题 — 20 分钟

---

## 创业公司（50～200 人规模）

### 面试特点

- **轮次少**：通常 1～2 轮技术面
- **实用导向**：更关注能不能快速上手干活
- **全栈能力**：可能需要兼顾前端、运维、数据库管理
- **项目经验**：重点看过往项目的完整度和独立性
- **算法较少**：可能不考或只考简单题

### 面试重点

| 优先级 | 知识点 | 难度 | 关联模块 |
|--------|--------|------|---------|
| P0 | Go 基础语法 | ⭐ | [Go 基础语法](/1-go-core/1.1-go-basics/) |
| P0 | Gin REST API 开发 | ⭐⭐ | [Gin 框架](/2-web-data/2.1-web-framework/03-gin) |
| P0 | GORM 数据库操作 | ⭐⭐ | [GORM](/2-web-data/2.2-database/02-gorm) |
| P0 | Redis 基本使用 | ⭐ | [go-redis 客户端](/2-web-data/2.3-cache-search/07-redis-go-client) |
| P0 | JWT 认证 | ⭐⭐ | [JWT](/2-web-data/2.6-auth/01-jwt) |
| P1 | Docker 部署 | ⭐ | [Docker](/3-microservice/3.3-docker-k8s/01-docker-basics) |
| P1 | 日志管理 | ⭐ | [zerolog](/2-web-data/2.7-observability/02-zerolog) |
| P1 | 并发编程基础 | ⭐⭐ | [并发编程](/1-go-core/1.3-concurrent/) |
| P2 | 消息队列 | ⭐⭐ | [消息队列](/2-web-data/2.4-message-queue/) |
| P2 | Nginx 配置 | ⭐ | [Nginx](/5-devops/5.3-nginx/) |

### 创业公司典型面试流程

1. **技术面**（45～60 分钟）：
   - Go 基础 + 项目经验 — 20 分钟
   - 现场编码（实现一个简单 API）— 20 分钟
   - 技术方案讨论 — 10 分钟

2. **CTO/技术负责人面**（30 分钟）：
   - 技术视野和学习能力
   - 团队协作和沟通能力

---

## 云原生/基础设施公司（阿里云、华为云、PingCAP）

### 面试特点

- **底层原理**：深入考察 Go 运行时、操作系统、网络协议
- **开源经验**：有开源项目贡献经验是加分项
- **系统设计**：重点考察分布式系统设计能力
- **K8s 生态**：云原生岗位必考 Kubernetes 相关知识

### 面试重点

| 优先级 | 知识点 | 难度 | 关联模块 |
|--------|--------|------|---------|
| P0 | GMP 调度模型（深入） | ⭐⭐⭐ | [GMP 调度模型](/1-go-core/1.4-runtime/01-gmp) |
| P0 | 内存管理和 GC | ⭐⭐⭐ | [内存管理](/1-go-core/1.4-runtime/03-memory) |
| P0 | K8s 架构和核心组件 | ⭐⭐⭐ | [K8s 架构](/3-microservice/3.3-docker-k8s/05-k8s-architecture) |
| P0 | etcd 和 Raft | ⭐⭐⭐ | [etcd](/3-microservice/3.2-service-governance/01-etcd) |
| P0 | 分布式系统理论 | ⭐⭐⭐ | [分布式系统](/4-distributed/4.1-distributed/) |
| P1 | Docker 底层原理 | ⭐⭐ | [Docker 核心概念](/3-microservice/3.3-docker-k8s/01-docker-basics) |
| P1 | gRPC 和 Protobuf | ⭐⭐ | [gRPC](/2-web-data/2.1-web-framework/04-grpc) |
| P1 | Prometheus 监控 | ⭐⭐ | [Prometheus](/2-web-data/2.7-observability/08-prometheus) |
| P1 | 性能调优（pprof） | ⭐⭐ | [pprof](/1-go-core/1.4-runtime/05-pprof) |
| P2 | AWS 云服务 | ⭐⭐ | [云服务集成](/3-microservice/3.4-aws/) |

---

## IoT/智能硬件公司

### 面试特点

- **协议知识**：MQTT、CoAP 等 IoT 协议
- **嵌入式思维**：资源受限环境下的编程
- **交叉编译**：Go 的交叉编译能力
- **实时性**：低延迟通信和数据处理

### 面试重点

| 优先级 | 知识点 | 难度 | 关联模块 |
|--------|--------|------|---------|
| P0 | MQTT 协议 | ⭐⭐ | [MQTT](/2-web-data/2.4-message-queue/04-mqtt) |
| P0 | Go 并发编程 | ⭐⭐ | [并发编程](/1-go-core/1.3-concurrent/) |
| P0 | AWS IoT Core | ⭐⭐ | [IoT Core](/3-microservice/3.4-aws/08-iot-core) |
| P1 | Go 基础语法 | ⭐ | [Go 基础语法](/1-go-core/1.1-go-basics/) |
| P1 | 消息队列 | ⭐⭐ | [消息队列](/2-web-data/2.4-message-queue/) |
| P1 | Docker 部署 | ⭐ | [Docker](/3-microservice/3.3-docker-k8s/) |
| P2 | 数据库操作 | ⭐⭐ | [数据库与 ORM](/2-web-data/2.2-database/) |

---

## 面试准备建议总结

| 公司类型 | 准备周期 | 重点方向 | 推荐学习路径 |
|---------|---------|---------|-------------|
| 大厂 | 4～6 周 | 底层原理 + 系统设计 + 算法 | [面试突击路径](/learning-paths/interview-sprint) + [高级深入路径](/learning-paths/advanced) |
| 中厂 | 2～3 周 | 中间件使用 + 项目经验 | [面试突击路径](/learning-paths/interview-sprint) |
| 创业公司 | 1～2 周 | 实战能力 + 全栈技能 | [初学者路径](/learning-paths/beginner) + [中级进阶路径](/learning-paths/intermediate) |
| 云原生公司 | 4～6 周 | K8s + 分布式 + 底层原理 | [云原生工程师路径](/learning-paths/cloud-native) |
| IoT 公司 | 2～3 周 | MQTT + 并发 + 嵌入式 | [中级进阶路径](/learning-paths/intermediate) |
