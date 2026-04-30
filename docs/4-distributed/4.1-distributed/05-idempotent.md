---
title: "幂等性设计"
module: "distributed"
difficulty: "advanced"
interviewFrequency: "high"
tags:
  - 幂等性
  - 接口设计
  - Token 机制
  - 唯一索引
  - 分布式
codeExample: "04-distributed/distributed/"
relatedEntries:
  - "/4-distributed/4.1-distributed/03-distributed-lock"
  - "/4-distributed/4.1-distributed/04-distributed-transaction"
prerequisites:
  - "/1-go-core/1.1-go-basics/"
estimatedTime: "40min"
---

# 幂等性设计

## 概念说明

幂等性（Idempotency）是指同一个操作执行一次和执行多次的效果相同。在分布式系统中，由于网络超时、重试机制、消息重复消费等原因，同一个请求可能被执行多次，幂等性设计是保证系统正确性的基本要求。

**为什么需要幂等性？**
- 网络超时后客户端重试
- 消息队列重复投递
- 用户重复点击提交按钮
- 微服务间调用失败重试

**天然幂等的操作**：查询（SELECT）、删除（DELETE WHERE id=x）、更新为固定值（UPDATE SET status='done'）
**非幂等的操作**：插入（INSERT）、累加（UPDATE SET count=count+1）、扣减（UPDATE SET stock=stock-1）

## 核心原理

### 常见幂等性方案

| 方案 | 原理 | 适用场景 | 优缺点 |
|------|------|----------|--------|
| **Token 机制** | 请求前获取 Token，提交时校验并删除 | 表单提交、订单创建 | 简单，需额外请求获取 Token |
| **唯一索引** | 数据库唯一索引防止重复插入 | 订单号、流水号 | 可靠，依赖数据库 |
| **状态机** | 状态只能单向流转，重复请求被拒绝 | 订单状态变更 | 业务语义清晰 |
| **乐观锁** | 版本号控制，CAS 更新 | 库存扣减、余额变更 | 高并发下重试多 |
| **分布式锁** | 加锁保证同一时刻只有一个请求执行 | 通用场景 | 性能开销 |
| **去重表** | 将请求唯一标识存入去重表 | 消息消费、支付回调 | 通用，需维护去重表 |

### Token 机制流程

1. 客户端请求服务端获取 Token（存入 Redis，设置过期时间）
2. 客户端提交请求时携带 Token
3. 服务端用 Lua 脚本原子性地校验并删除 Token
4. Token 存在则执行业务，不存在则拒绝（重复请求）

### 去重表方案

```go
// 去重表结构
// CREATE TABLE idempotent_record (
//     request_id VARCHAR(64) PRIMARY KEY,
//     result     TEXT,
//     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
//     expired_at TIMESTAMP
// )

func ProcessWithIdempotent(requestID string, fn func() (string, error)) (string, error) {
    // 1. 查询去重表
    record, err := getRecord(requestID)
    if err == nil {
        return record.Result, nil // 已处理过，直接返回结果
    }
    
    // 2. 执行业务逻辑
    result, err := fn()
    if err != nil {
        return "", err
    }
    
    // 3. 写入去重表
    saveRecord(requestID, result)
    return result, nil
}
```

## 标准库方案

Go 标准库没有直接的幂等性支持，但提供了构建幂等性方案的基础工具：

```go
// 使用 sync.Once 实现单机幂等（只执行一次）
var once sync.Once
once.Do(func() {
    // 只会执行一次的初始化操作
})

// 使用 sync.Map 实现简单的内存去重
var processed sync.Map
if _, loaded := processed.LoadOrStore(requestID, true); loaded {
    return // 已处理过
}
```

## 第三方库方案

幂等性通常是业务层面的设计，没有通用的第三方库。常见的实现方式：

- **Redis + Lua**：Token 机制、去重表
- **数据库唯一索引**：订单号、流水号
- **消息队列 offset**：Kafka 消费者 offset 管理

## 代码示例

> 💻 完整可运行代码：[code-examples/04-distributed/distributed/](https://github.com/your-repo/code-examples/04-distributed/distributed/)
> 🏷️ 幂等性是设计层面的知识，核心代码示例体现在分布式锁和架构设计模块中

## 常见面试题

### Q1: 什么是幂等性？如何设计幂等接口？

**难度**：⭐⭐⭐ | **频率**：🔥🔥🔥

**答题思路**：

1. 定义幂等性
2. 说明为什么需要幂等性
3. 列举常见方案
4. 结合具体场景说明

**标准答案**：

幂等性是指同一操作执行一次和多次效果相同。在分布式系统中，网络超时重试、消息重复消费等场景要求接口必须幂等。常见方案包括：Token 机制（请求前获取 Token，提交时校验删除）、数据库唯一索引（防止重复插入）、状态机（状态单向流转）、乐观锁（版本号 CAS 更新）、去重表（记录已处理的请求 ID）。选择哪种方案取决于业务场景和性能要求。

**深入追问**：

- 如何保证 Token 机制的原子性？
- 去重表会不会无限增长？怎么清理？
- HTTP 的哪些方法是幂等的？

### Q2: 如何保证消息消费的幂等性？

**难度**：⭐⭐⭐ | **频率**：🔥🔥🔥

**答题思路**：

1. 消息重复消费的原因
2. 去重表方案
3. 业务层面的幂等设计

**标准答案**：

消息重复消费的原因包括：生产者重试发送、消费者处理成功但 ACK 失败、消息队列 rebalance 等。保证幂等的方案：一是使用去重表，将消息唯一 ID 存入数据库或 Redis，消费前先查询是否已处理；二是业务层面设计幂等操作，如使用唯一索引防止重复插入、使用乐观锁防止重复扣减。推荐两种方案结合使用。

**深入追问**：

- Kafka 的 Exactly-Once 语义是怎么实现的？
- 去重表用 Redis 还是数据库？各有什么优缺点？

## 常见陷阱

1. **只考虑正常流程的幂等**：异常流程（如半成功状态）也需要幂等处理
2. **去重表不设过期时间**：去重记录会无限增长，需要定期清理
3. **Token 校验不是原子操作**：必须用 Lua 脚本保证"检查+删除"的原子性
4. **忽视并发场景**：两个相同请求同时到达，去重表查询都返回"未处理"

## 参考资料

- [幂等性设计模式](https://microservices.io/patterns/communication-style/idempotent-consumer.html)
- [Go 官方文档](https://go.dev/doc/)
