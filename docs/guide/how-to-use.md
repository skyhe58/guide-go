# 使用指南

## 知识库组织方式

本知识库采用 **五层递进学习架构**，每一层建立在前一层的基础之上：

```
第一层：语言核心（必修基础）
    ↓
第二层：Web 开发与数据（完成第一层后学习）
    ↓
第三层：微服务与云原生（完成第二层后学习）
    ↓
第四层：分布式与架构（完成第三层后学习）
    ↓
第五层：运维与部署（贯穿全程）
```

## 知识条目结构

每个知识条目遵循统一的文档模板，包含以下章节：

| 章节 | 说明 |
|------|------|
| **概念说明** | 用通俗易懂的语言解释知识点 |
| **核心原理** | 深入分析底层原理，复杂流程配 Mermaid 图 |
| **标准库方案** | 优先展示 Go 标准库实现，体现 Go 哲学 |
| **第三方库方案** | 介绍更成熟的第三方方案及选型建议 |
| **代码示例** | 链接到可运行的 Go 代码 |
| **常见面试题** | 高频面试题及答题思路 |
| **常见陷阱** | Go 开发者容易犯的错误 |
| **参考资料** | 官方文档和优质学习资源 |

## 难度标识

知识条目通过 frontmatter 标注难度和面试频率：

| 难度 | 含义 | 面试场景 |
|------|------|---------|
| `beginner` | 初级 — 基础概念和用法 | 校招 / 初级岗位 |
| `intermediate` | 中级 — 原理和源码分析 | 中级岗位 / 大厂初面 |
| `advanced` | 高级 — 底层实现和架构设计 | 高级岗位 / 大厂深面 |

| 面试频率 | 含义 | 建议 |
|---------|------|------|
| 🔥🔥🔥 `high` | 几乎每次面试都会问 | 必须掌握，优先复习 |
| 🔥🔥 `medium` | 经常出现 | 应该掌握 |
| 🔥 `low` | 偶尔出现 | 了解即可 |

## 代码示例说明

### Go Workspace 多模块管理

所有代码示例位于 `code-examples/` 目录，使用 Go Workspace（`go.work`）统一管理 20 个独立 Go Module。各模块互不依赖，可独立编译运行。

### Demo 模式

| 模式 | 说明 | 适用场景 |
|------|------|---------|
| **Part A** | 纯内存模拟，直接 `go run` | 原理类：GMP 调度、GC、B+树、Raft |
| **Part B** | 连接真实中间件，需 Docker | API 类：Redis 操作、Kafka 收发、gRPC |
| **混合** | Part A + Part B | 既有原理又有操作 |

### Docker 中间件

依赖外部服务的代码示例需要通过 Docker Compose 启动中间件：

```bash
# 基础中间件（MySQL/PostgreSQL/Redis/MongoDB/MinIO）
docker compose -f docker/docker-compose.yml up -d

# 消息队列（Kafka/NATS/RabbitMQ/EMQX）
docker compose -f docker/docker-compose.mq.yml up -d

# 按需启动单个服务
docker compose -f docker/docker-compose.yml up -d redis
```

## 面试准备建议

1. **按模块复习**：每个模块末尾都有 `interview.md` 面试指南
2. **查看知识图谱**：[面试知识图谱](/interview/knowledge-map) 展示知识点关联和追问路径
3. **按公司类型准备**：[按公司分类](/interview/by-company) 了解不同公司的面试重点
4. **动手实践**：运行代码示例加深理解，尤其是并发、切片、接口等高频考点

## 贡献指南

欢迎贡献内容！请参阅项目根目录的 [CONTRIBUTING.md](https://github.com/your-username/guide-go/blob/main/CONTRIBUTING.md) 了解如何添加新模块或知识点。
