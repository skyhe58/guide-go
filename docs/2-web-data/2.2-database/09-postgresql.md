---
title: "PostgreSQL"
module: "database"
difficulty: "advanced"
interviewFrequency: "medium"
tags:
  - PostgreSQL
  - pgx
  - pgxpool
  - JSONB
  - Listen-Notify
  - 多租户
codeExample: "02-web-data/database/database-sql/"
relatedEntries:
  - "/2-web-data/2.2-database/01-database-sql"
  - "/2-web-data/2.2-database/02-gorm"
  - "/2-web-data/2.2-database/05-mysql-index"
prerequisites:
  - "/2-web-data/2.2-database/01-database-sql"
estimatedTime: "50min"
---

# PostgreSQL

## 概念说明

PostgreSQL 是世界上最先进的开源关系型数据库，以其丰富的数据类型、强大的扩展性和严格的 SQL 标准遵从性著称。在 Go 生态中，pgx 是最推荐的 PostgreSQL 驱动，提供了比 `database/sql` 更高效的原生接口。

PostgreSQL 与 MySQL 的核心差异使其在以下场景更具优势：
- **复杂查询**：CTE、窗口函数、递归查询
- **JSON 数据**：原生 JSONB 类型，支持索引和查询
- **数据完整性**：更严格的约束检查、更丰富的数据类型
- **扩展性**：支持自定义类型、函数、索引方法

## 核心原理

### PostgreSQL vs MySQL 核心差异

| 维度 | PostgreSQL | MySQL (InnoDB) |
|------|-----------|----------------|
| MVCC 实现 | 多版本存储在表中（需要 VACUUM） | undo log 版本链 |
| 默认隔离级别 | READ COMMITTED | REPEATABLE READ |
| JSON 支持 | JSONB（二进制存储，可索引） | JSON（文本存储，5.7+） |
| 数组类型 | 原生支持 `integer[]` | 不支持 |
| CTE 支持 | 完整支持（含递归 CTE） | 8.0+ 支持 |
| 窗口函数 | 完整支持 | 8.0+ 支持 |
| 全文搜索 | 内置 `tsvector/tsquery` | 内置但功能较弱 |
| 复制方式 | 逻辑复制 + 流复制 | binlog 复制 |
| 扩展机制 | Extension（PostGIS、pg_trgm 等） | 插件（较少） |
| 序列 | `SERIAL` / `GENERATED ALWAYS` | `AUTO_INCREMENT` |

### pgx 驱动架构

```mermaid
graph TB
    APP[Go 应用] --> CHOICE{选择接口}
    CHOICE -->|标准接口| STDLIB["pgx/stdlib<br/>兼容 database/sql"]
    CHOICE -->|原生接口| PGX["pgx 原生 API<br/>性能更好"]

    STDLIB --> POOL1["database/sql 连接池"]
    PGX --> POOL2["pgxpool 连接池"]

    POOL1 --> PG[(PostgreSQL)]
    POOL2 --> PG

    subgraph "pgx 特色功能"
        COPY["COPY 协议<br/>批量导入"]
        LISTEN["Listen/Notify<br/>实时通知"]
        BATCH["Batch 批量查询"]
        CUSTOM["自定义类型映射"]
    end

    PGX --> COPY
    PGX --> LISTEN
    PGX --> BATCH
    PGX --> CUSTOM
```

### COPY 协议批量导入

PostgreSQL 的 COPY 协议是批量数据导入的最高效方式，比逐行 INSERT 快 10-100 倍：

```go
// 使用 pgx 的 CopyFrom 批量导入
rows := [][]interface{}{
    {"张三", "zhangsan@example.com", 25},
    {"李四", "lisi@example.com", 30},
    {"王五", "wangwu@example.com", 28},
}

copyCount, err := conn.CopyFrom(
    ctx,
    pgx.Identifier{"users"},
    []string{"name", "email", "age"},
    pgx.CopyFromRows(rows),
)
```

### Listen/Notify 实时通知

PostgreSQL 内置的发布/订阅机制，适合轻量级实时通知场景：

```mermaid
sequenceDiagram
    participant Pub as 发布者
    participant PG as PostgreSQL
    participant Sub as 订阅者

    Sub->>PG: LISTEN channel_name
    Pub->>PG: NOTIFY channel_name, 'payload'
    PG-->>Sub: 推送通知 (channel, payload)
```

```go
// 监听通知
conn.Exec(ctx, "LISTEN order_events")

for {
    notification, err := conn.WaitForNotification(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Channel: %s, Payload: %s\n", notification.Channel, notification.Payload)
}

// 发送通知
conn.Exec(ctx, "NOTIFY order_events, '{\"order_id\": 123, \"status\": \"paid\"}'")
```

### JSONB 操作

```sql
-- 创建包含 JSONB 字段的表
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    attrs JSONB DEFAULT '{}'
);

-- 插入 JSONB 数据
INSERT INTO products (name, attrs) VALUES
('iPhone', '{"color": "black", "storage": 256, "tags": ["phone", "apple"]}');

-- JSONB 查询
SELECT * FROM products WHERE attrs->>'color' = 'black';        -- 文本提取
SELECT * FROM products WHERE (attrs->>'storage')::int > 128;   -- 类型转换
SELECT * FROM products WHERE attrs @> '{"color": "black"}';    -- 包含查询
SELECT * FROM products WHERE attrs ? 'color';                  -- 键存在

-- JSONB 索引（GIN 索引）
CREATE INDEX idx_products_attrs ON products USING GIN (attrs);
```

### 多租户方案

```mermaid
graph TB
    subgraph "方案一：Schema 级别隔离"
        APP1[应用层] --> ROUTER1[Schema 路由]
        ROUTER1 --> S1["tenant_a Schema"]
        ROUTER1 --> S2["tenant_b Schema"]
        ROUTER1 --> S3["tenant_c Schema"]
        S1 --> DB1[(同一数据库)]
        S2 --> DB1
        S3 --> DB1
    end

    subgraph "方案二：按租户分库"
        APP2[应用层] --> ROUTER2[数据库路由]
        ROUTER2 --> D1[(tenant_a DB)]
        ROUTER2 --> D2[(tenant_b DB)]
        ROUTER2 --> D3[(tenant_c DB)]
    end
```

| 方案 | 隔离性 | 运维复杂度 | 适用场景 |
|------|--------|-----------|---------|
| Schema 隔离 | 中 | 低 | 租户数量适中（<100） |
| 按租户分库 | 高 | 高 | 数据安全要求高、租户数据量大 |
| 行级隔离（tenant_id） | 低 | 最低 | 租户数量多、数据量小 |

Schema 隔离示例：

```go
// 设置当前 Schema（每个请求根据租户切换）
func setTenantSchema(ctx context.Context, pool *pgxpool.Pool, tenantID string) error {
    _, err := pool.Exec(ctx, fmt.Sprintf("SET search_path TO tenant_%s, public", tenantID))
    return err
}
```

## 第三方库方案

### pgx 连接池（pgxpool）

```go
import "github.com/jackc/pgx/v5/pgxpool"

config, err := pgxpool.ParseConfig("postgres://postgres:postgres123@localhost:5432/testdb")
config.MaxConns = 25
config.MinConns = 5
config.MaxConnLifetime = 5 * time.Minute

pool, err := pgxpool.NewWithConfig(ctx, config)
defer pool.Close()

// 查询
var name string
err = pool.QueryRow(ctx, "SELECT name FROM users WHERE id = $1", 1).Scan(&name)

// 批量查询（Batch）
batch := &pgx.Batch{}
batch.Queue("SELECT name FROM users WHERE id = $1", 1)
batch.Queue("SELECT name FROM users WHERE id = $1", 2)
results := pool.SendBatch(ctx, batch)
defer results.Close()
```

### GORM + PostgreSQL

::: v-pre
```go
import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

dsn := "host=localhost user=postgres password=postgres123 dbname=testdb port=5432 sslmode=disable"
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

// PostgreSQL 特有类型
type Product struct {
    gorm.Model
    Name  string
    Tags  pq.StringArray `gorm:"type:text[]"`       // 数组类型
    Attrs datatypes.JSON `gorm:"type:jsonb"`         // JSONB 类型
}

// Upsert（ON CONFLICT）
db.Clauses(clause.OnConflict{
    Columns:   []clause.Column{{Name: "email"}},
    DoUpdates: clause.AssignmentColumns([]string{"name", "age"}),
}).Create(&user)
```
:::

## 代码示例

> 💻 完整可运行代码：[code-examples/02-web-data/database/database-sql/](https://github.com/skyhe58/guide-go/tree/main/code-examples/02-web-data/database/database-sql/)
> 🏷️ Demo 模式：Part A（模拟 PostgreSQL 特性）/ Part B（连接真实 PostgreSQL）

## 常见面试题

### Q1: PostgreSQL 和 MySQL 的 MVCC 实现有什么区别？

**难度**：⭐⭐⭐⭐ | **频率**：🔥🔥🔥

**标准答案**：

- **MySQL InnoDB**：旧版本存储在 undo log 中，通过 roll_pointer 形成版本链。读取时沿版本链查找可见版本。undo log 由 purge 线程异步清理。
- **PostgreSQL**：旧版本直接存储在表的数据页中（堆表），通过 `xmin`/`xmax` 标记版本可见性。需要 VACUUM 进程定期清理死元组（dead tuples），否则表会膨胀。

PostgreSQL 的方式更简单直接，但需要 VACUUM 维护；MySQL 的方式不需要额外清理表空间，但 undo log 管理更复杂。

### Q2: 什么场景下选择 PostgreSQL 而不是 MySQL？

**难度**：⭐⭐⭐ | **频率**：🔥🔥

**标准答案**：

选择 PostgreSQL 的场景：
1. 需要 JSONB 存储和查询（如半结构化数据）
2. 需要复杂查询（CTE、窗口函数、递归查询）
3. 需要地理空间数据（PostGIS 扩展）
4. 需要数组类型
5. 数据完整性要求高（更严格的约束检查）
6. 多租户 Schema 隔离方案

## 常见陷阱

1. **VACUUM 不及时**：导致表膨胀和性能下降，应配置 autovacuum
2. **连接数限制**：PostgreSQL 每个连接一个进程，连接数不宜过多，建议使用 pgxpool 或 PgBouncer
3. **$1 占位符**：PostgreSQL 使用 `$1, $2` 而不是 MySQL 的 `?` 占位符

## 参考资料

- [PostgreSQL 官方文档](https://www.postgresql.org/docs/)
- [pgx GitHub](https://github.com/jackc/pgx)
- [GORM PostgreSQL 驱动](https://gorm.io/docs/connecting_to_the_database.html#PostgreSQL)
