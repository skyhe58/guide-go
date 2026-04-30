---
title: 面试突击路径
description: 2～3 周快速复习 Go 核心知识点，高效准备面试
---

# 面试突击路径

## 适合人群

- 有 Go 开发经验，近期准备面试的开发者
- 已完成知识库部分模块学习，需要系统复习的学习者
- 想快速了解 Go 面试重点的求职者

## 预计时间

**2～3 周**（每天 3～4 小时）

## 前置条件

- 有至少半年的 Go 开发经验（或完成初学者路径）
- 了解基本的数据结构和算法
- 有 Web 后端开发的基本概念

## 学习策略

面试突击路径按照**面试频率**排序，优先复习最高频的知识点。每个知识点聚焦于：
1. 核心概念速记
2. 面试题 + 标准答案
3. 常见追问和陷阱

> 💡 建议配合 [面试知识图谱](/interview/knowledge-map) 和 [按公司类型面试重点](/interview/by-company) 使用

## 学习步骤

### 第一阶段：Go 语言核心高频题（第 1 周）

> 目标：覆盖 Go 语言层面的高频面试题，这是所有 Go 面试的必考内容

#### Day 1-2：基础语法高频题

| 知识点 | 面试频率 | 文档链接 | 重点掌握 |
|--------|---------|---------|---------|
| Slice 扩容机制 | 🔥🔥🔥 | [数组与切片](/1-go-core/1.1-go-basics/09-slice) | 扩容策略、底层数组共享陷阱 |
| defer 执行顺序 | 🔥🔥🔥 | [函数](/1-go-core/1.1-go-basics/06-functions) | LIFO 顺序、与 return 的关系 |
| 值传递 vs 引用传递 | 🔥🔥🔥 | [指针](/1-go-core/1.1-go-basics/11-pointer) | Go 只有值传递、slice/map 的传递本质 |
| Map 底层原理 | 🔥🔥 | [Map](/1-go-core/1.1-go-basics/10-map) | 哈希表实现、并发不安全 |
| 错误处理模式 | 🔥🔥 | [错误处理](/1-go-core/1.1-go-basics/07-error-handling) | errors.Is/As、%w 包装 |
| **面试指南** | — | [Go 基础面试指南](/1-go-core/1.1-go-basics/interview) | 系统复习 |

#### Day 3：进阶特性高频题

| 知识点 | 面试频率 | 文档链接 | 重点掌握 |
|--------|---------|---------|---------|
| 接口 nil 判断陷阱 | 🔥🔥🔥 | [接口](/1-go-core/1.2-go-advanced/01-interfaces) | 接口值的内部结构（type, value） |
| 反射性能 | 🔥🔥 | [反射](/1-go-core/1.2-go-advanced/03-reflection) | 反射的使用场景和性能开销 |
| 泛型使用场景 | 🔥🔥 | [泛型](/1-go-core/1.2-go-advanced/04-generics) | 何时该用泛型、何时不该用 |
| **面试指南** | — | [Go 进阶面试指南](/1-go-core/1.2-go-advanced/interview) | 系统复习 |

#### Day 4-5：并发编程高频题

| 知识点 | 面试频率 | 文档链接 | 重点掌握 |
|--------|---------|---------|---------|
| Goroutine 泄漏 | 🔥🔥🔥 | [goroutine](/1-go-core/1.3-concurrent/01-goroutine) | 泄漏原因、检测方法、预防措施 |
| Channel 死锁 | 🔥🔥🔥 | [channel](/1-go-core/1.3-concurrent/02-channel) | 常见死锁场景、select 多路复用 |
| sync.Pool 原理 | 🔥🔥 | [sync 包](/1-go-core/1.3-concurrent/03-sync) | 对象复用、GC 时清空 |
| Context 使用 | 🔥🔥🔥 | [context 包](/1-go-core/1.3-concurrent/04-context) | 传播机制、最佳实践 |
| **面试指南** | — | [并发编程面试指南](/1-go-core/1.3-concurrent/interview) | 系统复习 |

#### Day 6-7：运行时高频题

| 知识点 | 面试频率 | 文档链接 | 重点掌握 |
|--------|---------|---------|---------|
| GMP 调度模型 | 🔥🔥🔥 | [GMP 调度模型](/1-go-core/1.4-runtime/01-gmp) | G/M/P 关系、调度策略、抢占式调度 |
| GC 三色标记法 | 🔥🔥🔥 | [垃圾回收](/1-go-core/1.4-runtime/02-gc) | 三色标记、写屏障、GC 触发条件 |
| 内存逃逸 | 🔥🔥 | [逃逸分析](/1-go-core/1.4-runtime/09-escape) | 逃逸场景、分析方法 |
| **面试指南** | — | [运行时面试指南](/1-go-core/1.4-runtime/interview) | 系统复习 |

**🏁 里程碑检查点 1：**
- [ ] 能流畅回答 slice 扩容、defer 顺序、值传递等基础题
- [ ] 能画出 GMP 调度模型并解释调度流程
- [ ] 能分析 goroutine 泄漏的常见原因和解决方案
- [ ] 能解释 GC 三色标记法和写屏障的作用

### 第二阶段：中间件与数据库高频题（第 2 周）

#### Day 8-9：MySQL 高频题

| 知识点 | 面试频率 | 文档链接 | 重点掌握 |
|--------|---------|---------|---------|
| B+树索引原理 | 🔥🔥🔥 | [MySQL 索引原理](/2-web-data/2.2-database/05-mysql-index) | B+树结构、聚簇索引、覆盖索引 |
| 事务隔离级别 | 🔥🔥🔥 | [MySQL 事务](/2-web-data/2.2-database/06-mysql-transaction) | MVCC 实现、幻读解决 |
| 锁机制 | 🔥🔥 | [MySQL 锁机制](/2-web-data/2.2-database/07-mysql-lock) | 行锁、间隙锁、临键锁 |
| SQL 优化 | 🔥🔥 | [SQL 优化](/2-web-data/2.2-database/08-mysql-optimization) | EXPLAIN 分析、索引失效场景 |
| **面试指南** | — | [数据库面试指南](/2-web-data/2.2-database/interview) | 系统复习 |

#### Day 10-11：Redis 高频题

| 知识点 | 面试频率 | 文档链接 | 重点掌握 |
|--------|---------|---------|---------|
| 缓存穿透/击穿/雪崩 | 🔥🔥🔥 | [缓存问题](/2-web-data/2.3-cache-search/05-redis-cache-problems) | 三种问题的区别和解决方案 |
| 分布式锁 | 🔥🔥🔥 | [分布式锁](/2-web-data/2.3-cache-search/06-redis-distributed-lock) | Redlock 算法、etcd 锁对比 |
| Redis 数据结构 | 🔥🔥 | [Redis 数据结构](/2-web-data/2.3-cache-search/01-redis-data-structures) | 底层实现（SDS、跳表、压缩列表） |
| 持久化 | 🔥🔥 | [Redis 持久化](/2-web-data/2.3-cache-search/02-redis-persistence) | RDB vs AOF、混合持久化 |
| **面试指南** | — | [缓存与搜索面试指南](/2-web-data/2.3-cache-search/interview) | 系统复习 |

#### Day 12-13：消息队列与微服务

| 知识点 | 面试频率 | 文档链接 | 重点掌握 |
|--------|---------|---------|---------|
| Kafka 架构 | 🔥🔥 | [Kafka](/2-web-data/2.4-message-queue/01-kafka) | 分区、消费组、消息可靠性 |
| 消息队列选型 | 🔥🔥 | [选型对比](/2-web-data/2.4-message-queue/05-comparison) | Kafka vs RabbitMQ vs NATS |
| 微服务框架选型 | 🔥🔥 | [框架选型](/3-microservice/3.1-microservice/04-comparison) | Kratos vs Go-Zero |
| etcd 与 Raft | 🔥🔥 | [etcd](/3-microservice/3.2-service-governance/01-etcd) | Raft 一致性、Watch 机制 |

#### Day 14：Docker 与 K8s

| 知识点 | 面试频率 | 文档链接 | 重点掌握 |
|--------|---------|---------|---------|
| Docker 多阶段构建 | 🔥🔥 | [Go 应用 Dockerfile](/3-microservice/3.3-docker-k8s/02-dockerfile) | scratch 基础镜像、镜像瘦身 |
| K8s 核心概念 | 🔥🔥 | [K8s 架构](/3-microservice/3.3-docker-k8s/05-k8s-architecture) | Pod/Deployment/Service |

**🏁 里程碑检查点 2：**
- [ ] 能画出 B+树结构并解释聚簇索引和非聚簇索引的区别
- [ ] 能清晰区分缓存穿透、击穿、雪崩并给出解决方案
- [ ] 能对比分析 Kafka 和 RabbitMQ 的适用场景
- [ ] 能解释 Raft 算法的 Leader 选举流程

### 第三阶段：系统设计与架构（第 3 周）

#### Day 15-17：分布式系统

| 知识点 | 面试频率 | 文档链接 | 重点掌握 |
|--------|---------|---------|---------|
| 分布式锁方案 | 🔥🔥🔥 | [分布式锁](/4-distributed/4.1-distributed/03-distributed-lock) | Redis 锁 vs etcd 锁 |
| 分布式事务 | 🔥🔥 | [分布式事务](/4-distributed/4.1-distributed/04-distributed-transaction) | 2PC/TCC/Saga/消息最终一致性 |
| 限流算法 | 🔥🔥 | [限流算法](/4-distributed/4.1-distributed/06-rate-limiting) | 令牌桶、滑动窗口 |
| 熔断降级 | 🔥🔥 | [熔断与降级](/4-distributed/4.1-distributed/07-circuit-breaker) | 熔断器状态机 |

#### Day 18-21：架构设计题

| 知识点 | 面试频率 | 文档链接 | 重点掌握 |
|--------|---------|---------|---------|
| 秒杀系统 | 🔥🔥🔥 | [秒杀系统](/4-distributed/4.2-architecture/01-seckill) | 限流 → 库存扣减 → 异步下单 |
| 缓存一致性 | 🔥🔥🔥 | [缓存一致性](/4-distributed/4.2-architecture/04-cache-consistency) | 双写一致性方案对比 |
| 短链接系统 | 🔥🔥 | [短链接系统](/4-distributed/4.2-architecture/02-short-url) | ID 生成策略、302 重定向 |
| 接口幂等性 | 🔥🔥 | [接口幂等性](/4-distributed/4.2-architecture/05-idempotent-design) | Token 机制、唯一索引 |

**🏁 里程碑检查点 3（面试突击完成）：**
- [ ] 能在 30 分钟内完整设计一个秒杀系统
- [ ] 能对比分析至少 3 种分布式事务方案
- [ ] 能针对不同公司类型调整面试准备重点
- [ ] 对 Go 面试知识图谱中的高频节点全部掌握

## 面试前最后检查

- [ ] 复习 [面试知识图谱](/interview/knowledge-map)，确认所有高频节点已掌握
- [ ] 阅读 [按公司类型面试重点](/interview/by-company)，针对目标公司调整准备
- [ ] 准备 2～3 个项目经验的 STAR 描述
- [ ] 准备 1～2 个系统设计题的完整方案
