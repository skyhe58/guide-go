// go-elasticsearch 完整示例 — 倒排索引模拟 / CRUD / DSL 查询 / 聚合
// 演示：倒排索引原理、ES 文档 CRUD、DSL 查询、聚合分析
// Go 1.22+ | 验证日期 2025-01-01
//
// Part A：纯内存模拟，直接运行理解 Elasticsearch 核心概念
// Part B：连接真实 Elasticsearch，需传入参数 'real'
//
// 运行方式：
//   go run ./elasticsearch/              # Part A：内存模拟
//   go run ./elasticsearch/ real         # Part B：连接真实 ES
//
// Part B 前置条件：
//   docker compose -f docker/docker-compose.es.yml up -d
//   连接地址：localhost:9200，无需认证

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

// ============================================================
// Part A：纯内存模拟 Elasticsearch 核心概念
// ============================================================

func partA() {
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println("Part A：Elasticsearch 核心概念（内存模拟）")
	fmt.Println("=" + strings.Repeat("=", 59))

	// 1. 倒排索引原理
	demoInvertedIndex()

	// 2. 文档 CRUD 概念
	demoCRUDConcept()

	// 3. DSL 查询概念
	demoDSLConcept()

	// 4. 聚合分析概念
	demoAggregationConcept()
}

// ============================================================
// 1. 倒排索引原理（内存模拟）
// ============================================================

// InvertedIndex 倒排索引：词项 → 文档 ID 列表
type InvertedIndex map[string][]int

// Document 文档模型
type Document struct {
	ID      int
	Title   string
	Content string
	Author  string
	Tags    []string
	Views   int
	Created time.Time
}

// simpleTokenize 简单分词器（按空格和标点分割，转小写）
func simpleTokenize(text string) []string {
	// 替换常见标点为空格
	replacer := strings.NewReplacer(
		"，", " ", "。", " ", "！", " ", "？", " ",
		",", " ", ".", " ", "!", " ", "?", " ",
		"（", " ", "）", " ", "(", " ", ")", " ",
	)
	text = replacer.Replace(text)
	words := strings.Fields(text)
	tokens := make([]string, 0, len(words))
	for _, w := range words {
		w = strings.ToLower(strings.TrimSpace(w))
		if w != "" {
			tokens = append(tokens, w)
		}
	}
	return tokens
}

func demoInvertedIndex() {
	fmt.Println("\n--- 1. 倒排索引原理 ---")

	// 模拟文档集合
	docs := map[int]string{
		1: "Go 语言入门教程",
		2: "Go 并发编程指南",
		3: "Redis 缓存入门",
		4: "Go Redis 实战",
	}

	// 构建倒排索引
	index := make(InvertedIndex)
	for docID, content := range docs {
		tokens := simpleTokenize(content)
		for _, token := range tokens {
			// 去重：检查 docID 是否已存在
			found := false
			for _, id := range index[token] {
				if id == docID {
					found = true
					break
				}
			}
			if !found {
				index[token] = append(index[token], docID)
			}
		}
	}

	// 展示倒排索引
	fmt.Println("\n  文档集合（正排索引）：")
	for id := 1; id <= 4; id++ {
		fmt.Printf("    Doc %d: %s\n", id, docs[id])
	}

	fmt.Println("\n  倒排索引（Inverted Index）：")
	// 按词项排序展示
	terms := make([]string, 0, len(index))
	for term := range index {
		terms = append(terms, term)
	}
	sort.Strings(terms)
	for _, term := range terms {
		docIDs := index[term]
		sort.Ints(docIDs)
		fmt.Printf("    %-10s → Doc %v\n", term, docIDs)
	}

	// 模拟搜索
	fmt.Println("\n  搜索 'go'：")
	if docIDs, ok := index["go"]; ok {
		for _, id := range docIDs {
			fmt.Printf("    命中 Doc %d: %s\n", id, docs[id])
		}
	}

	fmt.Println("\n  搜索 'go' AND '入门'（交集）：")
	goIDs := index["go"]
	entryIDs := index["入门"]
	for _, gid := range goIDs {
		for _, eid := range entryIDs {
			if gid == eid {
				fmt.Printf("    命中 Doc %d: %s\n", gid, docs[gid])
			}
		}
	}
}

// ============================================================
// 2. 文档 CRUD 概念
// ============================================================

func demoCRUDConcept() {
	fmt.Println("\n--- 2. 文档 CRUD 概念 ---")

	// 模拟 ES 索引
	esIndex := make(map[string]Document)

	// Create
	doc1 := Document{
		ID: 1, Title: "Go 入门", Content: "Go 语言基础教程",
		Author: "张三", Tags: []string{"Go", "入门"}, Views: 100,
	}
	esIndex["1"] = doc1
	fmt.Println("\n  [Create] PUT /articles/_doc/1")
	fmt.Printf("    → 索引文档: %+v\n", doc1)

	// Read
	fmt.Println("\n  [Read] GET /articles/_doc/1")
	if doc, ok := esIndex["1"]; ok {
		fmt.Printf("    → 获取文档: Title=%s, Author=%s\n", doc.Title, doc.Author)
	}

	// Update
	doc1.Title = "Go 入门（修订版）"
	doc1.Views = 150
	esIndex["1"] = doc1
	fmt.Println("\n  [Update] POST /articles/_update/1")
	fmt.Printf("    → 更新后: Title=%s, Views=%d\n", doc1.Title, doc1.Views)

	// Delete
	delete(esIndex, "1")
	fmt.Println("\n  [Delete] DELETE /articles/_doc/1")
	fmt.Printf("    → 文档已删除，索引中剩余 %d 条\n", len(esIndex))

	// Bulk
	fmt.Println("\n  [Bulk] POST /_bulk")
	bulkDocs := []Document{
		{ID: 1, Title: "Go 入门", Author: "张三", Views: 100, Tags: []string{"Go", "入门"}},
		{ID: 2, Title: "Go 并发", Author: "李四", Views: 200, Tags: []string{"Go", "并发"}},
		{ID: 3, Title: "Redis 入门", Author: "张三", Views: 150, Tags: []string{"Redis", "入门"}},
	}
	for _, doc := range bulkDocs {
		esIndex[fmt.Sprintf("%d", doc.ID)] = doc
	}
	fmt.Printf("    → 批量写入 %d 条文档\n", len(bulkDocs))
}

// ============================================================
// 3. DSL 查询概念
// ============================================================

func demoDSLConcept() {
	fmt.Println("\n--- 3. DSL 查询概念 ---")

	// 准备测试数据
	docs := []Document{
		{ID: 1, Title: "Go 语言入门教程", Author: "张三", Tags: []string{"Go", "入门"}, Views: 100},
		{ID: 2, Title: "Go 并发编程指南", Author: "李四", Tags: []string{"Go", "并发"}, Views: 250},
		{ID: 3, Title: "Redis 缓存入门", Author: "张三", Tags: []string{"Redis", "入门"}, Views: 180},
		{ID: 4, Title: "Go Redis 实战", Author: "王五", Tags: []string{"Go", "Redis"}, Views: 320},
		{ID: 5, Title: "Elasticsearch 搜索引擎", Author: "李四", Tags: []string{"ES", "搜索"}, Views: 90},
	}

	// match 查询（分词后匹配）
	fmt.Println("\n  [match] 搜索 title 包含 'Go'：")
	for _, doc := range docs {
		if strings.Contains(strings.ToLower(doc.Title), "go") {
			fmt.Printf("    命中 Doc %d: %s (author: %s)\n", doc.ID, doc.Title, doc.Author)
		}
	}

	// term 查询（精确匹配）
	fmt.Println("\n  [term] 精确匹配 author='张三'：")
	for _, doc := range docs {
		if doc.Author == "张三" {
			fmt.Printf("    命中 Doc %d: %s\n", doc.ID, doc.Title)
		}
	}

	// bool 查询（组合条件）
	fmt.Println("\n  [bool] must: title 含 'Go', filter: author='张三'：")
	for _, doc := range docs {
		titleMatch := strings.Contains(strings.ToLower(doc.Title), "go")
		authorMatch := doc.Author == "张三"
		if titleMatch && authorMatch {
			fmt.Printf("    命中 Doc %d: %s\n", doc.ID, doc.Title)
		}
	}

	// range 查询
	fmt.Println("\n  [range] views >= 150：")
	for _, doc := range docs {
		if doc.Views >= 150 {
			fmt.Printf("    命中 Doc %d: %s (views: %d)\n", doc.ID, doc.Title, doc.Views)
		}
	}

	// 排序 + 分页
	fmt.Println("\n  [sort + pagination] 按 views 降序，取前 3 条：")
	sorted := make([]Document, len(docs))
	copy(sorted, docs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Views > sorted[j].Views
	})
	for i := 0; i < 3 && i < len(sorted); i++ {
		fmt.Printf("    #%d Doc %d: %s (views: %d)\n", i+1, sorted[i].ID, sorted[i].Title, sorted[i].Views)
	}
}

// ============================================================
// 4. 聚合分析概念
// ============================================================

func demoAggregationConcept() {
	fmt.Println("\n--- 4. 聚合分析概念 ---")

	docs := []Document{
		{ID: 1, Title: "Go 入门", Author: "张三", Tags: []string{"Go", "入门"}, Views: 100},
		{ID: 2, Title: "Go 并发", Author: "李四", Tags: []string{"Go", "并发"}, Views: 250},
		{ID: 3, Title: "Redis 入门", Author: "张三", Tags: []string{"Redis", "入门"}, Views: 180},
		{ID: 4, Title: "Go Redis", Author: "王五", Tags: []string{"Go", "Redis"}, Views: 320},
		{ID: 5, Title: "ES 搜索", Author: "李四", Tags: []string{"ES", "搜索"}, Views: 90},
	}

	// terms 聚合：按作者分组
	fmt.Println("\n  [terms agg] 按作者分组统计文章数：")
	authorCount := make(map[string]int)
	for _, doc := range docs {
		authorCount[doc.Author]++
	}
	for author, count := range authorCount {
		fmt.Printf("    %s: %d 篇\n", author, count)
	}

	// stats 聚合：阅读量统计
	fmt.Println("\n  [stats agg] 阅读量统计：")
	var totalViews, minViews, maxViews int
	minViews = docs[0].Views
	for _, doc := range docs {
		totalViews += doc.Views
		if doc.Views < minViews {
			minViews = doc.Views
		}
		if doc.Views > maxViews {
			maxViews = doc.Views
		}
	}
	avgViews := float64(totalViews) / float64(len(docs))
	fmt.Printf("    count: %d, min: %d, max: %d, avg: %.1f, sum: %d\n",
		len(docs), minViews, maxViews, avgViews, totalViews)

	// 嵌套聚合：按作者分组，每组计算平均阅读量
	fmt.Println("\n  [nested agg] 按作者分组，计算平均阅读量：")
	authorViews := make(map[string][]int)
	for _, doc := range docs {
		authorViews[doc.Author] = append(authorViews[doc.Author], doc.Views)
	}
	for author, views := range authorViews {
		sum := 0
		for _, v := range views {
			sum += v
		}
		avg := float64(sum) / float64(len(views))
		fmt.Printf("    %s: %d 篇, 平均阅读量 %.1f\n", author, len(views), avg)
	}

	// tags 聚合
	fmt.Println("\n  [terms agg] 热门标签 Top 5：")
	tagCount := make(map[string]int)
	for _, doc := range docs {
		for _, tag := range doc.Tags {
			tagCount[tag]++
		}
	}
	type tagStat struct {
		Tag   string
		Count int
	}
	tagStats := make([]tagStat, 0, len(tagCount))
	for tag, count := range tagCount {
		tagStats = append(tagStats, tagStat{tag, count})
	}
	sort.Slice(tagStats, func(i, j int) bool {
		return tagStats[i].Count > tagStats[j].Count
	})
	for i, ts := range tagStats {
		if i >= 5 {
			break
		}
		fmt.Printf("    %s: %d 篇\n", ts.Tag, ts.Count)
	}
}

// ============================================================
// Part B：连接真实 Elasticsearch（需要 Docker）
// ============================================================

func partB() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Part B：连接真实 Elasticsearch（go-elasticsearch）")
	fmt.Println(strings.Repeat("=", 60))

	ctx := context.Background()

	// 创建 ES 客户端
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
	})
	if err != nil {
		fmt.Printf("❌ 创建 ES 客户端失败: %v\n", err)
		return
	}

	// 测试连接
	res, err := es.Info()
	if err != nil {
		fmt.Printf("❌ 无法连接 Elasticsearch: %v\n", err)
		fmt.Println("请先启动 ES: docker compose -f docker/docker-compose.es.yml up -d")
		return
	}
	defer res.Body.Close()
	if res.IsError() {
		fmt.Printf("❌ ES 返回错误: %s\n", res.String())
		return
	}
	fmt.Println("✅ Elasticsearch 连接成功")

	// 1. 创建索引
	demoRealCreateIndex(ctx, es)

	// 2. 文档 CRUD
	demoRealCRUD(ctx, es)

	// 3. DSL 查询
	demoRealDSLQuery(ctx, es)

	// 4. 聚合分析
	demoRealAggregation(ctx, es)

	// 清理测试索引
	cleanupTestIndex(es)
}

// demoRealCreateIndex 创建索引（带映射）
func demoRealCreateIndex(ctx context.Context, es *elasticsearch.Client) {
	fmt.Println("\n--- 1. 创建索引 ---")

	indexName := "demo-articles"

	// 先删除已有索引
	es.Indices.Delete([]string{indexName})

	// 创建索引（带映射）
	mapping := `{
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0
		},
		"mappings": {
			"properties": {
				"title":      { "type": "text" },
				"content":    { "type": "text" },
				"author":     { "type": "keyword" },
				"tags":       { "type": "keyword" },
				"views":      { "type": "integer" },
				"created_at": { "type": "date" }
			}
		}
	}`

	res, err := es.Indices.Create(indexName, es.Indices.Create.WithBody(strings.NewReader(mapping)))
	if err != nil {
		fmt.Printf("  创建索引失败: %v\n", err)
		return
	}
	defer res.Body.Close()
	if res.IsError() {
		fmt.Printf("  创建索引失败: %s\n", res.String())
		return
	}
	fmt.Printf("  ✅ 索引 '%s' 创建成功\n", indexName)
}

// demoRealCRUD 文档 CRUD 操作
func demoRealCRUD(ctx context.Context, es *elasticsearch.Client) {
	fmt.Println("\n--- 2. 文档 CRUD ---")

	indexName := "demo-articles"

	// 批量写入文档
	articles := []map[string]interface{}{
		{"title": "Go 语言入门教程", "content": "Go 是一门简洁高效的编程语言", "author": "张三", "tags": []string{"Go", "入门"}, "views": 100, "created_at": "2024-01-15"},
		{"title": "Go 并发编程指南", "content": "goroutine 和 channel 是 Go 并发的核心", "author": "李四", "tags": []string{"Go", "并发"}, "views": 250, "created_at": "2024-03-20"},
		{"title": "Redis 缓存入门", "content": "Redis 是高性能的内存数据库", "author": "张三", "tags": []string{"Redis", "入门"}, "views": 180, "created_at": "2024-05-10"},
		{"title": "Go Redis 实战", "content": "使用 go-redis 操作 Redis 数据结构", "author": "王五", "tags": []string{"Go", "Redis"}, "views": 320, "created_at": "2024-07-01"},
		{"title": "Elasticsearch 搜索引擎", "content": "ES 基于倒排索引实现全文搜索", "author": "李四", "tags": []string{"ES", "搜索"}, "views": 90, "created_at": "2024-09-15"},
	}

	for i, article := range articles {
		body, _ := json.Marshal(article)
		docID := fmt.Sprintf("%d", i+1)
		res, err := es.Index(indexName, bytes.NewReader(body),
			es.Index.WithDocumentID(docID),
			es.Index.WithRefresh("true"), // 立即可搜索
		)
		if err != nil {
			fmt.Printf("  写入文档 %s 失败: %v\n", docID, err)
			continue
		}
		res.Body.Close()
	}
	fmt.Printf("  ✅ 批量写入 %d 条文档\n", len(articles))

	// 读取文档
	res, err := es.Get(indexName, "1")
	if err != nil {
		fmt.Printf("  读取文档失败: %v\n", err)
		return
	}
	defer res.Body.Close()
	var getResult map[string]interface{}
	json.NewDecoder(res.Body).Decode(&getResult)
	if source, ok := getResult["_source"].(map[string]interface{}); ok {
		fmt.Printf("  GET /demo-articles/_doc/1 → title: %s, author: %s\n",
			source["title"], source["author"])
	}

	// 更新文档
	updateBody := strings.NewReader(`{"doc": {"views": 150, "title": "Go 语言入门教程（修订版）"}}`)
	res, err = es.Update(indexName, "1", updateBody, es.Update.WithRefresh("true"))
	if err != nil {
		fmt.Printf("  更新文档失败: %v\n", err)
		return
	}
	res.Body.Close()
	fmt.Println("  ✅ 更新文档 1 成功")
}

// demoRealDSLQuery DSL 查询
func demoRealDSLQuery(ctx context.Context, es *elasticsearch.Client) {
	fmt.Println("\n--- 3. DSL 查询 ---")

	indexName := "demo-articles"

	// match 查询
	fmt.Println("\n  [match] 搜索 title 包含 'Go'：")
	query := `{"query": {"match": {"title": "Go"}}}`
	searchAndPrint(es, indexName, query)

	// term 查询
	fmt.Println("\n  [term] 精确匹配 author='张三'：")
	query = `{"query": {"term": {"author": "张三"}}}`
	searchAndPrint(es, indexName, query)

	// bool 查询
	fmt.Println("\n  [bool] must: title 含 'Go', filter: views >= 200：")
	query = `{
		"query": {
			"bool": {
				"must": [{"match": {"title": "Go"}}],
				"filter": [{"range": {"views": {"gte": 200}}}]
			}
		}
	}`
	searchAndPrint(es, indexName, query)

	// 排序 + 分页
	fmt.Println("\n  [sort] 按 views 降序，取前 3 条：")
	query = `{
		"query": {"match_all": {}},
		"sort": [{"views": "desc"}],
		"size": 3
	}`
	searchAndPrint(es, indexName, query)
}

// demoRealAggregation 聚合分析
func demoRealAggregation(ctx context.Context, es *elasticsearch.Client) {
	fmt.Println("\n--- 4. 聚合分析 ---")

	indexName := "demo-articles"

	// terms 聚合：按作者分组
	fmt.Println("\n  [terms agg] 按作者分组：")
	query := `{
		"size": 0,
		"aggs": {
			"by_author": {
				"terms": {"field": "author"}
			}
		}
	}`
	aggResult := searchRaw(es, indexName, query)
	if aggs, ok := aggResult["aggregations"].(map[string]interface{}); ok {
		if byAuthor, ok := aggs["by_author"].(map[string]interface{}); ok {
			if buckets, ok := byAuthor["buckets"].([]interface{}); ok {
				for _, b := range buckets {
					bucket := b.(map[string]interface{})
					fmt.Printf("    %s: %.0f 篇\n", bucket["key"], bucket["doc_count"])
				}
			}
		}
	}

	// stats 聚合：阅读量统计
	fmt.Println("\n  [stats agg] 阅读量统计：")
	query = `{
		"size": 0,
		"aggs": {
			"view_stats": {
				"stats": {"field": "views"}
			}
		}
	}`
	aggResult = searchRaw(es, indexName, query)
	if aggs, ok := aggResult["aggregations"].(map[string]interface{}); ok {
		if stats, ok := aggs["view_stats"].(map[string]interface{}); ok {
			fmt.Printf("    count: %.0f, min: %.0f, max: %.0f, avg: %.1f, sum: %.0f\n",
				stats["count"], stats["min"], stats["max"], stats["avg"], stats["sum"])
		}
	}

	// 嵌套聚合：按作者分组 + 平均阅读量
	fmt.Println("\n  [nested agg] 按作者分组，计算平均阅读量：")
	query = `{
		"size": 0,
		"aggs": {
			"by_author": {
				"terms": {"field": "author"},
				"aggs": {
					"avg_views": {"avg": {"field": "views"}}
				}
			}
		}
	}`
	aggResult = searchRaw(es, indexName, query)
	if aggs, ok := aggResult["aggregations"].(map[string]interface{}); ok {
		if byAuthor, ok := aggs["by_author"].(map[string]interface{}); ok {
			if buckets, ok := byAuthor["buckets"].([]interface{}); ok {
				for _, b := range buckets {
					bucket := b.(map[string]interface{})
					avgViews := bucket["avg_views"].(map[string]interface{})
					fmt.Printf("    %s: %.0f 篇, 平均阅读量 %.1f\n",
						bucket["key"], bucket["doc_count"], avgViews["value"])
				}
			}
		}
	}
}

// searchAndPrint 执行搜索并打印结果
func searchAndPrint(es *elasticsearch.Client, index, query string) {
	res, err := es.Search(
		es.Search.WithIndex(index),
		es.Search.WithBody(strings.NewReader(query)),
	)
	if err != nil {
		fmt.Printf("    搜索失败: %v\n", err)
		return
	}
	defer res.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(res.Body).Decode(&result)

	hits := result["hits"].(map[string]interface{})
	total := hits["total"].(map[string]interface{})
	fmt.Printf("    共命中 %.0f 条：\n", total["value"])

	for _, hit := range hits["hits"].([]interface{}) {
		h := hit.(map[string]interface{})
		source := h["_source"].(map[string]interface{})
		views := ""
		if v, ok := source["views"]; ok {
			views = fmt.Sprintf(", views: %.0f", v)
		}
		fmt.Printf("      [%s] %s (author: %s%s)\n",
			h["_id"], source["title"], source["author"], views)
	}
}

// searchRaw 执行搜索并返回原始结果
func searchRaw(es *elasticsearch.Client, index, query string) map[string]interface{} {
	res, err := es.Search(
		es.Search.WithIndex(index),
		es.Search.WithBody(strings.NewReader(query)),
	)
	if err != nil {
		fmt.Printf("    搜索失败: %v\n", err)
		return nil
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return result
}

// cleanupTestIndex 清理测试索引
func cleanupTestIndex(es *elasticsearch.Client) {
	es.Indices.Delete([]string{"demo-articles"})
}

// ============================================================
// main 入口
// ============================================================

func main() {
	// Part A：纯内存模拟，直接运行理解原理
	partA()

	// Part B：连接真实 Elasticsearch，需传入参数 'real'
	if len(os.Args) > 1 && os.Args[1] == "real" {
		partB()
	}
}
