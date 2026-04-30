# 代码质量待优化清单

> 已完成模块中存在过多 fmt.Println 说明型代码的问题，Part A 内存模拟部分应改为真实的数据结构和算法实现。

## 待优化模块

### 01-go-core
- [ ] `design-patterns/project-layout/main.go` — 26 处打印，应改为真实的项目结构生成器
- [ ] `runtime/gmp/main.go` — GMP 模拟应有真实的调度器数据结构
- [ ] `runtime/gc/main.go` — GC 模拟应有真实的三色标记算法实现

### 02-web-data
- [ ] `cache-search/redis/main.go` — 74 处 Println，Part A 应实现真实的内存缓存（LRU/分布式锁状态机）
- [ ] `cache-search/elasticsearch/main.go` — Part A 倒排索引模拟可以，但减少说明性打印
- [ ] `database/database-sql/main.go` — 81 处 Println，Part A 应实现真实的连接池模拟/事务状态机
- [ ] `database/gorm-examples/main.go` — Part A 应实现真实的 ORM 概念演示
- [ ] `database/sqlx-examples/main.go` — Part A 应实现真实的结构体映射演示
- [ ] `database/migration/main.go` — 迁移管理器模拟质量尚可，但减少说明性打印
- [ ] `message-queue/kafka/main.go` — Part A 分区模拟质量尚可
- [ ] `message-queue/rabbitmq/main.go` — Part A Exchange 路由模拟质量尚可
- [ ] `message-queue/nats/main.go` — Part A 模拟质量尚可
- [ ] `message-queue/mqtt/main.go` — Part A 过多说明性打印

## 优化原则

1. Part A 必须有真实的数据结构和算法实现，不是打印教程
2. fmt.Println 只用于输出运行结果，不用于讲解概念
3. 概念说明放在代码注释中，不放在 fmt.Println 里
4. 每个示例应该是可以直接参考的生产级代码片段
