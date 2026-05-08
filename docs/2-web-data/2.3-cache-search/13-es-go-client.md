---
title: "go-elasticsearch 客户端使用"
module: "cache-search"
difficulty: "intermediate"
interviewFrequency: "medium"
tags:
  - Elasticsearch
  - go-elasticsearch
  - Go客户端
codeExample: "02-web-data/cache-search/elasticsearch/"
relatedEntries:
  - "/2-web-data/2.3-cache-search/10-es-crud"
  - "/2-web-data/2.3-cache-search/11-es-dsl"
  - "/2-web-data/2.3-cache-search/12-es-aggregation"
prerequisites:
  - "/2-web-data/2.3-cache-search/10-es-crud"
estimatedTime: "40min"
---

# go-elasticsearch 客户端使用

## 概念说明

`github.com/elastic/go-elasticsearch/v8` 是 Elasticsearch 官方维护的 Go 客户端库。它提供了低级别的 API（直接发送 HTTP 请求）和类型化的 API（TypedClient），支持所有 ES 操作。

Go 生态中操作 ES 的主要客户端：

| 库 | 维护方 | 特点 |
|----|--------|------|
| `go-elasticsearch` | Elastic 官方 | 官方维护、版本同步、低级 + 类型化 API |
| `olivere/elastic` | 社区 | API 更友好，但版本更新可能滞后 |

推荐使用官方的 `go-elasticsearch`，与 ES 版本保持同步。

## 核心原理

### 客户端初始化

```go
import "github.com/elastic/go-elasticsearch/v8"

// 基础配置
es, err := elasticsearch.NewClient(elasticsearch.Config{
    Addresses: []string{"http://localhost:9200"},
    // 如果开启了安全认证
    // Username: "elastic",
    // Password: "changeme",
})
```

### 低级别 API

直接构造 JSON 请求体，灵活但需要手动处理序列化：

```go
// 索引文档
body := strings.NewReader(`{"title":"Go 入门","author":"张三"}`)
res, err := es.Index("articles", body, es.Index.WithDocumentID("1"))

// 搜索
query := strings.NewReader(`{"query":{"match":{"title":"Go"}}}`)
res, err := es.Search(es.Search.WithIndex("articles"), es.Search.WithBody(query))
```

### TypedClient API

类型安全的 API，编译时检查参数：

```go
import "github.com/elastic/go-elasticsearch/v8/typedapi/types"

tc := es.TypedClient

// 搜索
res, err := tc.Search().
    Index("articles").
    Query(&types.Query{
        Match: map[string]types.MatchQuery{
            "title": {Query: "Go"},
        },
    }).
    Do(ctx)
```

### Bulk 批量操作

```go
// 使用 esutil.BulkIndexer 高效批量写入
bi, _ := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
    Client:     es,
    Index:      "articles",
    NumWorkers: 4,
    FlushBytes: 5e+6, // 5MB
})

bi.Add(ctx, esutil.BulkIndexerItem{
    Action:     "index",
    DocumentID: "1",
    Body:       strings.NewReader(`{"title":"Go 入门"}`),
})

bi.Close(ctx)
```

## 代码示例

> 💻 完整可运行代码：[code-examples/02-web-data/cache-search/elasticsearch/](https://github.com/skyhe58/guide-go/tree/main/code-examples/02-web-data/cache-search/elasticsearch/)
> 🏷️ Demo 模式：Part A（模拟 ES 客户端概念）/ Part B（go-elasticsearch 完整操作）

## 常见面试题

### Q1: Go 中操作 Elasticsearch 有哪些客户端？如何选择？

**难度**：⭐⭐ | **频率**：🔥🔥

**答题思路**：

1. 官方 go-elasticsearch vs 社区 olivere/elastic
2. 各自优缺点
3. 推荐选择

**标准答案**：

官方 `go-elasticsearch` 由 Elastic 维护，版本与 ES 同步更新，提供低级别 API 和 TypedClient 两种使用方式。社区 `olivere/elastic` API 更友好（链式调用），但版本更新可能滞后。新项目推荐使用官方客户端，保证兼容性和长期维护。

**深入追问**：
- 如何提升 ES 写入性能？（Bulk API、调整 refresh_interval、增加 worker 数）
- go-elasticsearch 的连接池如何配置？（底层使用 net/http.Transport）

## 常见陷阱

1. **响应体未关闭**：`res.Body` 必须 `defer res.Body.Close()`，否则连接泄漏
2. **版本不匹配**：go-elasticsearch 的大版本号必须与 ES 服务端版本一致（v8 对应 ES 8.x）
3. **错误处理不完整**：需要同时检查 `err` 和 `res.IsError()`

## 参考资料

- [go-elasticsearch GitHub](https://github.com/elastic/go-elasticsearch)
- [go-elasticsearch 官方文档](https://www.elastic.co/guide/en/elasticsearch/client/go-api/current/index.html)
