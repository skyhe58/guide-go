---
title: "MongoDB 文档数据库与 mongo-go-driver"
module: "object-storage"
difficulty: "intermediate"
interviewFrequency: "high"
tags:
  - MongoDB
  - 文档数据库
  - BSON
  - 聚合管道
  - mongo-go-driver
  - 索引
  - 事务
codeExample: "02-web-data/object-storage/mongodb/"
relatedEntries:
  - "/2-web-data/2.5-object-storage/01-minio"
  - "/2-web-data/2.2-database/01-database-sql"
prerequisites:
  - "/1-go-core/1.1-go-basics/"
estimatedTime: "60min"
---

# MongoDB 文档数据库与 mongo-go-driver

## 概念说明

**MongoDB** 是最流行的文档数据库（Document Database），使用 **BSON**（Binary JSON）格式存储数据。与关系型数据库的固定表结构不同，MongoDB 的文档模型允许每条记录拥有不同的字段结构，非常适合存储半结构化数据。

### 核心概念对照

| MongoDB 概念 | 关系型数据库对照 | 说明 |
|-------------|----------------|------|
| Database | Database | 数据库 |
| Collection | Table | 集合（表） |
| Document | Row | 文档（行） |
| Field | Column | 字段（列） |
| `_id` | Primary Key | 主键（自动生成 ObjectID） |
| Embedded Document | JOIN | 嵌套文档（替代关联查询） |
| Index | Index | 索引 |

### BSON 数据格式

BSON（Binary JSON）是 MongoDB 的数据存储格式，相比 JSON 有以下优势：

- **更多数据类型**：支持 Date、ObjectID、Binary、Decimal128 等 JSON 不支持的类型
- **二进制编码**：比 JSON 文本格式更紧凑，解析更快
- **可遍历性**：支持按字段名直接定位，无需解析整个文档

```go
// Go 中使用 bson 包构建文档
doc := bson.D{
    {Key: "name", Value: "张三"},
    {Key: "age", Value: 28},
    {Key: "tags", Value: bson.A{"Go", "后端"}},
    {Key: "address", Value: bson.D{
        {Key: "city", Value: "北京"},
        {Key: "district", Value: "朝阳区"},
    }},
}
```

## 核心原理

### MongoDB 架构

```mermaid
graph TB
    App[Go 应用<br/>mongo-go-driver] -->|MongoDB Wire Protocol| Server[MongoDB Server]
    
    Server --> DB1[Database: guide_go]
    Server --> DB2[Database: admin]
    
    DB1 --> Coll1[Collection: users]
    DB1 --> Coll2[Collection: articles]
    DB1 --> Coll3[Collection: logs]
    
    Coll1 --> Doc1["{ _id: ObjectID, name: '张三', age: 28 }"]
    Coll1 --> Doc2["{ _id: ObjectID, name: '李四', skills: ['Go','Java'] }"]
    
    subgraph "文档模型特点"
        F1[灵活 Schema：同一集合的文档可以有不同字段]
        F2[嵌套文档：替代 JOIN，减少查询次数]
        F3[数组字段：天然支持一对多关系]
    end
```

### 聚合管道流程

```mermaid
graph LR
    Input[输入文档集合] --> Match["$match<br/>过滤条件"]
    Match --> Group["$group<br/>分组聚合"]
    Group --> Sort["$sort<br/>排序"]
    Sort --> Project["$project<br/>字段投影"]
    Project --> Limit["$limit<br/>限制数量"]
    Limit --> Output[输出结果]
    
    style Match fill:#e1f5fe
    style Group fill:#f3e5f5
    style Sort fill:#e8f5e9
    style Project fill:#fff3e0
    style Limit fill:#fce4ec
```

### 索引类型

```mermaid
graph TB
    Index[MongoDB 索引] --> Single[单字段索引<br/>db.coll.createIndex]
    Index --> Compound[复合索引<br/>多字段组合]
    Index --> Text[文本索引<br/>全文搜索]
    Index --> Hashed[哈希索引<br/>等值查询]
    Index --> TTL[TTL 索引<br/>自动过期删除]
    Index --> Unique[唯一索引<br/>字段值唯一]
    
    Single --> S1["{ email: 1 }"]
    Compound --> C1["{ status: 1, created_at: -1 }"]
    TTL --> T1["{ expire_at: 1 }, expireAfterSeconds: 0"]
```

## 标准库方案

Go 标准库没有内置 MongoDB 驱动。官方推荐使用 `go.mongodb.org/mongo-driver`，这是 MongoDB 官方维护的 Go 驱动。

## 第三方库方案

### mongo-go-driver

`go.mongodb.org/mongo-driver` 是 MongoDB 官方 Go 驱动，提供完整的 CRUD、聚合、索引、事务支持。

**连接与基本操作：**

::: v-pre
```go
// 创建客户端
client, err := mongo.Connect(options.Client().ApplyURI(
    "mongodb://root:root123@localhost:27017",
))
defer client.Disconnect(ctx)

// 获取集合
coll := client.Database("guide_go").Collection("users")

// 插入文档
result, err := coll.InsertOne(ctx, bson.D{
    {Key: "name", Value: "张三"},
    {Key: "age", Value: 28},
})

// 查询文档
var user bson.M
err = coll.FindOne(ctx, bson.D{{Key: "name", Value: "张三"}}).Decode(&user)

// 更新文档
coll.UpdateOne(ctx, 
    bson.D{{Key: "name", Value: "张三"}},
    bson.D{{Key: "$set", Value: bson.D{{Key: "age", Value: 29}}}},
)

// 删除文档
coll.DeleteOne(ctx, bson.D{{Key: "name", Value: "张三"}})
```
:::

**聚合管道：**

::: v-pre
```go
// 按城市统计用户数量，按数量降序排列
pipeline := mongo.Pipeline{
    {{Key: "$group", Value: bson.D{
        {Key: "_id", Value: "$city"},
        {Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
        {Key: "avgAge", Value: bson.D{{Key: "$avg", Value: "$age"}}},
    }}},
    {{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
    {{Key: "$limit", Value: 10}},
}
cursor, err := coll.Aggregate(ctx, pipeline)
```
:::

**索引管理：**

::: v-pre
```go
// 创建单字段索引
coll.Indexes().CreateOne(ctx, mongo.IndexModel{
    Keys: bson.D{{Key: "email", Value: 1}},
    Options: options.Index().SetUnique(true),
})

// 创建复合索引
coll.Indexes().CreateOne(ctx, mongo.IndexModel{
    Keys: bson.D{
        {Key: "status", Value: 1},
        {Key: "created_at", Value: -1},
    },
})

// 创建 TTL 索引（文档自动过期）
coll.Indexes().CreateOne(ctx, mongo.IndexModel{
    Keys: bson.D{{Key: "expire_at", Value: 1}},
    Options: options.Index().SetExpireAfterSeconds(0),
})
```
:::

**事务：**

::: v-pre
```go
// MongoDB 4.0+ 支持多文档事务
session, err := client.StartSession()
defer session.EndSession(ctx)

_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
    // 扣减账户 A
    _, err := accountsColl.UpdateOne(sc,
        bson.D{{Key: "_id", Value: "A"}},
        bson.D{{Key: "$inc", Value: bson.D{{Key: "balance", Value: -100}}}},
    )
    if err != nil {
        return nil, err
    }
    // 增加账户 B
    _, err = accountsColl.UpdateOne(sc,
        bson.D{{Key: "_id", Value: "B"}},
        bson.D{{Key: "$inc", Value: bson.D{{Key: "balance", Value: 100}}}},
    )
    return nil, err
})
```
:::

## 代码示例

> 💻 完整可运行代码：[code-examples/02-web-data/object-storage/mongodb/](https://github.com/skyhe58/guide-go/tree/main/code-examples/02-web-data/object-storage/mongodb/)
> 🏷️ Demo 模式：Part A（内存模拟文档存储）/ Part B（连接真实 MongoDB）

## 常见面试题

### Q1: MongoDB 和 MySQL 的区别？什么场景选 MongoDB？

**难度**：⭐⭐ | **频率**：🔥🔥🔥

**答题思路**：

1. 数据模型对比（文档 vs 表）
2. 查询能力对比
3. 事务支持对比
4. 适用场景分析

**标准答案**：

MongoDB 使用灵活的文档模型（BSON），同一集合中的文档可以有不同字段；MySQL 使用固定的表结构（Schema）。MongoDB 适合内容管理系统、日志存储、IoT 数据、用户画像等 Schema 频繁变化的场景；MySQL 适合金融交易、订单系统等需要强一致性和复杂关联查询的场景。

MongoDB 4.0+ 支持多文档事务，但性能不如关系型数据库的事务。MongoDB 的优势在于水平扩展（原生 Sharding）和灵活的数据模型。

**深入追问**：

- MongoDB 的 BSON 和 JSON 有什么区别？
- MongoDB 的事务和 MySQL 的事务有什么不同？

### Q2: MongoDB 的聚合管道是什么？常用的聚合阶段有哪些？

**难度**：⭐⭐⭐ | **频率**：🔥🔥

**答题思路**：

1. 解释聚合管道的概念（流水线处理）
2. 列举常用阶段
3. 与 SQL 的 GROUP BY 对比

**标准答案**：

聚合管道（Aggregation Pipeline）是 MongoDB 的数据处理框架，文档依次通过多个阶段（Stage），每个阶段对文档进行转换或计算。类似 Unix 管道，前一个阶段的输出是后一个阶段的输入。

常用阶段：`$match`（过滤，类似 WHERE）、`$group`（分组聚合，类似 GROUP BY）、`$sort`（排序）、`$project`（字段投影，类似 SELECT）、`$limit`/`$skip`（分页）、`$lookup`（关联查询，类似 LEFT JOIN）、`$unwind`（展开数组）。

**深入追问**：

- 聚合管道的性能优化技巧？
- `$lookup` 和关系型数据库的 JOIN 有什么区别？

### Q3: MongoDB 的索引类型有哪些？如何优化查询性能？

**难度**：⭐⭐⭐ | **频率**：🔥🔥🔥

**答题思路**：

1. 列举索引类型
2. 索引选择原则
3. 常见索引优化技巧

**标准答案**：

MongoDB 支持多种索引类型：单字段索引、复合索引、文本索引（全文搜索）、哈希索引（等值查询）、TTL 索引（自动过期）、唯一索引、地理空间索引。

优化技巧：使用 `explain()` 分析查询计划；复合索引遵循 ESR 原则（Equality → Sort → Range）；避免全集合扫描（COLLSCAN）；覆盖索引（Covered Query）避免回表；合理使用 TTL 索引自动清理过期数据。

**深入追问**：

- 什么是覆盖索引？
- MongoDB 的索引和 MySQL 的 B+ 树索引有什么区别？

## 常见陷阱

1. **忘记创建索引**：MongoDB 默认只有 `_id` 索引，查询其他字段会全集合扫描。生产环境必须为常用查询字段创建索引
2. **嵌套文档过深**：虽然 MongoDB 支持嵌套文档，但嵌套过深会导致查询复杂、更新困难。建议嵌套不超过 3 层
3. **文档大小超限**：单个文档最大 16MB。存储大文件应使用 GridFS 或对象存储（MinIO/S3）
4. **事务滥用**：MongoDB 事务性能不如关系型数据库，不应将 MongoDB 当作关系型数据库使用。优先通过文档模型设计避免事务
5. **bson.D 和 bson.M 混用**：`bson.D` 保持字段顺序（用于命令和索引），`bson.M` 不保证顺序（用于普通查询）。创建索引时必须用 `bson.D`

## 参考资料

- [MongoDB 官方文档](https://www.mongodb.com/docs/)
- [mongo-go-driver 文档](https://pkg.go.dev/go.mongodb.org/mongo-driver)
- [MongoDB 聚合管道参考](https://www.mongodb.com/docs/manual/reference/operator/aggregation-pipeline/)
- [MongoDB 索引策略](https://www.mongodb.com/docs/manual/applications/indexes/)
- [MongoDB University 免费课程](https://university.mongodb.com/)
