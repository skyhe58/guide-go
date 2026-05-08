---
title: "数据库迁移"
module: "database"
difficulty: "intermediate"
interviewFrequency: "medium"
tags:
  - 数据库迁移
  - golang-migrate
  - goose
  - Schema 管理
codeExample: "02-web-data/database/migration/"
relatedEntries:
  - "/2-web-data/2.2-database/01-database-sql"
  - "/2-web-data/2.2-database/02-gorm"
prerequisites:
  - "/2-web-data/2.2-database/01-database-sql"
estimatedTime: "30min"
---

# 数据库迁移

## 概念说明

数据库迁移（Database Migration）是管理数据库 Schema 变更的版本化方案。类似于代码的版本控制（Git），数据库迁移工具将每次 Schema 变更记录为一个迁移文件，支持按顺序执行（up）和回滚（down），确保数据库结构在不同环境间保持一致。

为什么需要数据库迁移：
- **版本化管理**：每次 Schema 变更都有记录，可追溯
- **团队协作**：多人开发时避免 Schema 冲突
- **环境一致性**：开发、测试、生产环境的数据库结构保持同步
- **可回滚**：出问题时可以回退到上一个版本

Go 生态中主流的迁移工具：
- **golang-migrate**：最流行，支持 CLI 和 Go 库两种使用方式
- **goose**：轻量级，支持 SQL 和 Go 两种迁移文件格式
- **GORM AutoMigrate**：简单场景可用，但不支持回滚和复杂变更

## 核心原理

### 迁移工作流

```mermaid
sequenceDiagram
    participant Dev as 开发者
    participant Tool as 迁移工具
    participant DB as 数据库
    participant Table as schema_migrations 表

    Dev->>Tool: 创建迁移文件 (up + down)
    Dev->>Tool: 执行 migrate up
    Tool->>Table: 检查当前版本
    Table-->>Tool: 版本 3
    Tool->>DB: 执行 004_add_column.up.sql
    Tool->>Table: 记录版本 4
    Tool-->>Dev: 迁移完成

    Note over Dev,Table: 需要回滚时
    Dev->>Tool: 执行 migrate down 1
    Tool->>Table: 检查当前版本
    Table-->>Tool: 版本 4
    Tool->>DB: 执行 004_add_column.down.sql
    Tool->>Table: 记录版本 3
    Tool-->>Dev: 回滚完成
```

### 迁移文件结构

```
migrations/
├── 000001_create_users.up.sql      # 创建 users 表
├── 000001_create_users.down.sql    # 回滚：删除 users 表
├── 000002_add_email_index.up.sql   # 添加 email 索引
├── 000002_add_email_index.down.sql # 回滚：删除 email 索引
├── 000003_create_articles.up.sql   # 创建 articles 表
└── 000003_create_articles.down.sql # 回滚：删除 articles 表
```

### golang-migrate vs goose 对比

| 特性 | golang-migrate | goose |
|------|---------------|-------|
| 迁移文件格式 | SQL | SQL + Go |
| CLI 工具 | ✅ | ✅ |
| Go 库 | ✅ | ✅ |
| 版本号格式 | 时间戳或序号 | 时间戳或序号 |
| 数据库支持 | 20+ 种 | 10+ 种 |
| 社区活跃度 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| 特色 | 支持远程数据源 | 支持 Go 代码迁移 |

## 第三方库方案

### golang-migrate

```bash
# 安装 CLI
go install -tags 'mysql postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 创建迁移文件
migrate create -ext sql -dir migrations -seq create_users

# 执行迁移
migrate -path migrations -database "mysql://root:root123@tcp(localhost:3306)/testdb" up

# 回滚一步
migrate -path migrations -database "mysql://root:root123@tcp(localhost:3306)/testdb" down 1

# 查看当前版本
migrate -path migrations -database "mysql://root:root123@tcp(localhost:3306)/testdb" version
```

Go 库方式：

```go
import (
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/mysql"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

m, err := migrate.New("file://migrations", "mysql://root:root123@tcp(localhost:3306)/testdb")
if err != nil {
    log.Fatal(err)
}

// 执行所有待执行的迁移
if err := m.Up(); err != nil && err != migrate.ErrNoChange {
    log.Fatal(err)
}
```

### goose

```bash
# 安装
go install github.com/pressly/goose/v3/cmd/goose@latest

# 创建 SQL 迁移
goose -dir migrations create create_users sql

# 创建 Go 迁移
goose -dir migrations create seed_data go

# 执行迁移
goose -dir migrations mysql "root:root123@tcp(localhost:3306)/testdb" up

# 回滚
goose -dir migrations mysql "root:root123@tcp(localhost:3306)/testdb" down
```

### GORM AutoMigrate（简单场景）

```go
// 自动迁移：只能添加字段/索引，不能删除或修改
db.AutoMigrate(&User{}, &Article{}, &Tag{})

// 局限性：
// ❌ 不能删除字段
// ❌ 不能修改字段类型
// ❌ 不支持回滚
// ❌ 不适合生产环境
```

## 代码示例

> 💻 完整可运行代码：[code-examples/02-web-data/database/migration/](https://github.com/skyhe58/guide-go/tree/main/code-examples/02-web-data/database/migration/)
> 🏷️ Demo 模式：Part A（内存模拟迁移概念）

## 常见面试题

### Q1: 为什么不推荐在生产环境使用 GORM AutoMigrate？

**难度**：⭐⭐⭐ | **频率**：🔥🔥

**标准答案**：

AutoMigrate 只能添加字段和索引，不能删除字段、修改字段类型、不支持回滚。生产环境需要精确控制 Schema 变更，应使用 golang-migrate 或 goose 等专业迁移工具，配合 CI/CD 流水线自动执行。

### Q2: 数据库迁移的最佳实践有哪些？

**难度**：⭐⭐⭐ | **频率**：🔥🔥

**标准答案**：

1. 每个迁移文件必须有对应的 down 文件（可回滚）
2. 迁移文件一旦提交不可修改，只能创建新的迁移
3. 大表变更使用 `pt-online-schema-change` 或 `gh-ost` 避免锁表
4. 迁移文件纳入版本控制，与代码一起 review
5. CI/CD 中自动执行迁移，避免手动操作

## 常见陷阱

1. **修改已执行的迁移文件**：会导致校验和不匹配，应创建新的迁移文件
2. **大表 DDL 锁表**：ALTER TABLE 在大表上可能锁表数分钟，需使用在线 DDL 工具
3. **忘记写 down 文件**：回滚时无法执行，应养成同时编写 up/down 的习惯

## 参考资料

- [golang-migrate GitHub](https://github.com/golang-migrate/migrate)
- [goose GitHub](https://github.com/pressly/goose)
- [GORM 迁移文档](https://gorm.io/zh_CN/docs/migration.html)
