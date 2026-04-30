---
title: "ES 映射与分析器"
module: "cache-search"
difficulty: "intermediate"
interviewFrequency: "medium"
tags:
  - Elasticsearch
  - Mapping
  - 分析器
  - 字段类型
codeExample: "02-web-data/cache-search/elasticsearch/"
relatedEntries:
  - "/2-web-data/2.3-cache-search/08-es-inverted-index"
  - "/2-web-data/2.3-cache-search/10-es-crud"
prerequisites:
  - "/2-web-data/2.3-cache-search/08-es-inverted-index"
estimatedTime: "35min"
---

# ES 映射与分析器

## 概念说明

映射（Mapping）是 ES 中定义文档结构的方式，类似于关系型数据库的表结构定义（Schema）。它定义了每个字段的数据类型、是否索引、使用什么分析器等。分析器（Analyzer）决定了文本如何被分词和处理，直接影响搜索效果。

## 核心原理

### 常用字段类型

| 类型 | 说明 | 是否分词 | 典型场景 |
|------|------|---------|---------|
| `text` | 全文搜索字段 | 是 | 文章标题、内容 |
| `keyword` | 精确匹配字段 | 否 | 状态、标签、ID |
| `integer/long` | 整数 | 否 | 年龄、数量 |
| `float/double` | 浮点数 | 否 | 价格、评分 |
| `date` | 日期 | 否 | 创建时间 |
| `boolean` | 布尔值 | 否 | 是否发布 |
| `object` | JSON 对象 | - | 嵌套结构 |
| `nested` | 嵌套对象（独立索引） | - | 数组中的对象 |

### 分析器组成

```mermaid
graph LR
    A[原始文本] --> B[Character Filter<br/>字符过滤器]
    B --> C[Tokenizer<br/>分词器]
    C --> D[Token Filter<br/>词项过滤器]
    D --> E[最终词项]
```

**内置分析器**：

| 分析器 | 说明 | 示例输入 → 输出 |
|--------|------|----------------|
| `standard` | 默认，按词分割，小写化 | "Quick Brown Fox" → [quick, brown, fox] |
| `simple` | 按非字母字符分割 | "user-name_123" → [user, name] |
| `whitespace` | 按空格分割 | "Quick Brown" → [Quick, Brown] |
| `keyword` | 不分词，整体作为一个词项 | "Quick Brown" → [Quick Brown] |

**中文分析器**：
- **IK 分词器**：最常用的中文分词插件
  - `ik_smart`：粗粒度分词（"中华人民共和国" → "中华人民共和国"）
  - `ik_max_word`：细粒度分词（"中华人民共和国" → "中华人民共和国, 中华人民, 中华, 华人, 人民共和国, 人民, 共和国, 共和, 国"）

### 映射定义示例

```json
{
  "mappings": {
    "properties": {
      "title": {
        "type": "text",
        "analyzer": "ik_max_word",
        "search_analyzer": "ik_smart",
        "fields": {
          "keyword": { "type": "keyword" }
        }
      },
      "content": {
        "type": "text",
        "analyzer": "ik_max_word"
      },
      "status": { "type": "keyword" },
      "tags": { "type": "keyword" },
      "created_at": { "type": "date" },
      "view_count": { "type": "integer" }
    }
  }
}
```

### 动态映射 vs 显式映射

- **动态映射**：ES 自动推断字段类型，方便但不精确
- **显式映射**：手动定义字段类型，生产环境推荐

注意：映射一旦创建，字段类型不可修改（只能新增字段或重建索引）。

## 代码示例

> 💻 完整可运行代码：[code-examples/02-web-data/cache-search/elasticsearch/](https://github.com/your-repo/code-examples/02-web-data/cache-search/elasticsearch/)
> 🏷️ Demo 模式：Part A（模拟映射概念）/ Part B（go-elasticsearch 创建映射）

## 常见面试题

### Q1: text 和 keyword 类型的区别？

**难度**：⭐⭐ | **频率**：🔥🔥🔥

**答题思路**：

1. text 会分词，keyword 不分词
2. 各自适用场景
3. multi-field 同时使用两种类型

**标准答案**：

`text` 类型会经过分析器分词处理，适用于全文搜索（如文章标题、内容）。`keyword` 类型不分词，整体作为一个词项存储，适用于精确匹配（如状态、标签、ID）。实际中常用 multi-field 同时定义两种类型：`title` 用于全文搜索，`title.keyword` 用于精确匹配和聚合。

**深入追问**：
- 为什么 keyword 类型不能用于全文搜索？
- 如何选择合适的中文分析器？

## 常见陷阱

1. **映射不可修改**：字段类型一旦创建不可更改，需要重建索引（reindex）
2. **动态映射类型不准确**：字符串可能被映射为 text + keyword，数字字符串可能被映射为 text
3. **分析器不匹配**：索引时和搜索时使用不同的分析器可能导致搜索不到结果

## 参考资料

- [Elasticsearch 官方文档 - Mapping](https://www.elastic.co/guide/en/elasticsearch/reference/current/mapping.html)
- [Elasticsearch 官方文档 - Analyzers](https://www.elastic.co/guide/en/elasticsearch/reference/current/analysis.html)
