// mongo-go-driver 完整示例 — 文档模型 / CRUD / 聚合管道 / 索引管理 / 事务
// 演示：MongoDB 文档数据库的核心操作，包含内存模拟和真实 mongo-go-driver 调用
// Go 1.22+ | 验证日期 2025-01-01
//
// Part A：纯内存模拟，直接运行理解 MongoDB 核心概念
// Part B：连接真实 MongoDB，需传入参数 'real'
//
// 运行方式：
//   go run ./mongodb/              # Part A：内存模拟
//   go run ./mongodb/ real         # Part B：连接真实 MongoDB
//
// Part B 前置条件：
//   docker compose -f docker/docker-compose.yml up -d mongodb
//   连接地址：localhost:27017，用户名：root，密码：root123

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ============================================================
// Part A：纯内存模拟 MongoDB 核心概念
// ============================================================

func partA() {
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println("Part A：MongoDB 核心概念（内存模拟）")
	fmt.Println("=" + strings.Repeat("=", 59))

	db := NewMemoryDocStore()

	// 1. 文档模型与 BSON
	demoDocumentModel(db)

	// 2. CRUD 操作
	demoCRUDOperations(db)

	// 3. 聚合管道模拟
	demoAggregationPipeline(db)

	// 4. 索引模拟
	demoIndexManagement(db)

	// 5. 事务模拟
	demoTransactionConcept(db)
}

// ============================================================
// 内存文档存储实现
// ============================================================

// Document 文档（模拟 BSON 文档）
type Document map[string]interface{}

// IndexDef 索引定义
type IndexDef struct {
	Name   string
	Fields []string
	Unique bool
}

// Collection 集合
type Collection struct {
	Name    string
	Docs    []Document
	Indexes []IndexDef
}

// MemoryDocStore 内存文档数据库
type MemoryDocStore struct {
	mu          sync.RWMutex
	collections map[string]*Collection
	idCounter   int64
}

// NewMemoryDocStore 创建内存文档数据库
func NewMemoryDocStore() *MemoryDocStore {
	return &MemoryDocStore{
		collections: make(map[string]*Collection),
	}
}

// getOrCreateCollection 获取或创建集合
func (db *MemoryDocStore) getOrCreateCollection(name string) *Collection {
	if coll, exists := db.collections[name]; exists {
		return coll
	}
	coll := &Collection{
		Name: name,
		Docs: make([]Document, 0),
		Indexes: []IndexDef{
			{Name: "_id_", Fields: []string{"_id"}, Unique: true},
		},
	}
	db.collections[name] = coll
	return coll
}

// generateID 生成自增 ID（模拟 ObjectID）
func (db *MemoryDocStore) generateID() string {
	db.idCounter++
	return fmt.Sprintf("ObjectID(%06d)", db.idCounter)
}

// InsertOne 插入单个文档
func (db *MemoryDocStore) InsertOne(collName string, doc Document) (string, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	coll := db.getOrCreateCollection(collName)
	id := db.generateID()
	doc["_id"] = id
	coll.Docs = append(coll.Docs, doc)
	return id, nil
}

// InsertMany 批量插入文档
func (db *MemoryDocStore) InsertMany(collName string, docs []Document) ([]string, error) {
	ids := make([]string, 0, len(docs))
	for _, doc := range docs {
		id, err := db.InsertOne(collName, doc)
		if err != nil {
			return ids, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// matchDocument 检查文档是否匹配过滤条件
func matchDocument(doc Document, filter Document) bool {
	for key, filterVal := range filter {
		docVal, exists := doc[key]
		if !exists {
			return false
		}
		// 处理嵌套文档查询（如 {"address.city": "北京"}）
		if strings.Contains(key, ".") {
			parts := strings.SplitN(key, ".", 2)
			if nested, ok := docVal.(Document); ok {
				return matchDocument(nested, Document{parts[1]: filterVal})
			}
			return false
		}
		// 处理比较操作符（如 {"$gte": 25}）
		if filterMap, ok := filterVal.(Document); ok {
			for op, opVal := range filterMap {
				switch op {
				case "$gte":
					if !compareGTE(docVal, opVal) {
						return false
					}
				case "$lte":
					if !compareLTE(docVal, opVal) {
						return false
					}
				case "$gt":
					if !compareGT(docVal, opVal) {
						return false
					}
				case "$in":
					if arr, ok := opVal.([]interface{}); ok {
						found := false
						for _, v := range arr {
							if fmt.Sprintf("%v", docVal) == fmt.Sprintf("%v", v) {
								found = true
								break
							}
						}
						if !found {
							return false
						}
					}
				}
			}
			continue
		}
		if fmt.Sprintf("%v", docVal) != fmt.Sprintf("%v", filterVal) {
			return false
		}
	}
	return true
}

func compareGTE(a, b interface{}) bool {
	return toFloat(a) >= toFloat(b)
}

func compareLTE(a, b interface{}) bool {
	return toFloat(a) <= toFloat(b)
}

func compareGT(a, b interface{}) bool {
	return toFloat(a) > toFloat(b)
}

func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case float64:
		return val
	default:
		return 0
	}
}

// Find 查询文档
func (db *MemoryDocStore) Find(collName string, filter Document) []Document {
	db.mu.RLock()
	defer db.mu.RUnlock()
	coll, exists := db.collections[collName]
	if !exists {
		return nil
	}
	result := make([]Document, 0)
	for _, doc := range coll.Docs {
		if matchDocument(doc, filter) {
			// 深拷贝文档
			copied := make(Document)
			for k, v := range doc {
				copied[k] = v
			}
			result = append(result, copied)
		}
	}
	return result
}

// FindOne 查询单个文档
func (db *MemoryDocStore) FindOne(collName string, filter Document) (Document, bool) {
	docs := db.Find(collName, filter)
	if len(docs) == 0 {
		return nil, false
	}
	return docs[0], true
}

// UpdateOne 更新单个文档
func (db *MemoryDocStore) UpdateOne(collName string, filter Document, update Document) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	coll, exists := db.collections[collName]
	if !exists {
		return 0, nil
	}
	for i, doc := range coll.Docs {
		if matchDocument(doc, filter) {
			// 处理 $set 操作符
			if setFields, ok := update["$set"]; ok {
				if fields, ok := setFields.(Document); ok {
					for k, v := range fields {
						coll.Docs[i][k] = v
					}
				}
			}
			// 处理 $inc 操作符
			if incFields, ok := update["$inc"]; ok {
				if fields, ok := incFields.(Document); ok {
					for k, v := range fields {
						current := toFloat(coll.Docs[i][k])
						increment := toFloat(v)
						coll.Docs[i][k] = current + increment
					}
				}
			}
			return 1, nil
		}
	}
	return 0, nil
}

// DeleteOne 删除单个文档
func (db *MemoryDocStore) DeleteOne(collName string, filter Document) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	coll, exists := db.collections[collName]
	if !exists {
		return 0, nil
	}
	for i, doc := range coll.Docs {
		if matchDocument(doc, filter) {
			coll.Docs = append(coll.Docs[:i], coll.Docs[i+1:]...)
			return 1, nil
		}
	}
	return 0, nil
}

// DeleteMany 批量删除文档
func (db *MemoryDocStore) DeleteMany(collName string, filter Document) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	coll, exists := db.collections[collName]
	if !exists {
		return 0, nil
	}
	remaining := make([]Document, 0)
	deleted := 0
	for _, doc := range coll.Docs {
		if matchDocument(doc, filter) {
			deleted++
		} else {
			remaining = append(remaining, doc)
		}
	}
	coll.Docs = remaining
	return deleted, nil
}

// Count 统计文档数量
func (db *MemoryDocStore) Count(collName string, filter Document) int {
	return len(db.Find(collName, filter))
}

// CreateIndex 创建索引
func (db *MemoryDocStore) CreateIndex(collName string, fields []string, unique bool) string {
	db.mu.Lock()
	defer db.mu.Unlock()
	coll := db.getOrCreateCollection(collName)
	name := strings.Join(fields, "_") + "_1"
	coll.Indexes = append(coll.Indexes, IndexDef{
		Name:   name,
		Fields: fields,
		Unique: unique,
	})
	return name
}

// ListIndexes 列出索引
func (db *MemoryDocStore) ListIndexes(collName string) []IndexDef {
	db.mu.RLock()
	defer db.mu.RUnlock()
	coll, exists := db.collections[collName]
	if !exists {
		return nil
	}
	return coll.Indexes
}

// Aggregate 简化版聚合管道
func (db *MemoryDocStore) Aggregate(collName string, pipeline []Document) []Document {
	db.mu.RLock()
	defer db.mu.RUnlock()
	coll, exists := db.collections[collName]
	if !exists {
		return nil
	}
	// 复制文档集合作为管道输入
	docs := make([]Document, len(coll.Docs))
	for i, d := range coll.Docs {
		copied := make(Document)
		for k, v := range d {
			copied[k] = v
		}
		docs[i] = copied
	}
	// 依次执行管道阶段
	for _, stage := range pipeline {
		for op, val := range stage {
			switch op {
			case "$match":
				if filter, ok := val.(Document); ok {
					filtered := make([]Document, 0)
					for _, doc := range docs {
						if matchDocument(doc, filter) {
							filtered = append(filtered, doc)
						}
					}
					docs = filtered
				}
			case "$group":
				docs = executeGroup(docs, val.(Document))
			case "$sort":
				if sortSpec, ok := val.(Document); ok {
					executeSort(docs, sortSpec)
				}
			case "$limit":
				if limit, ok := val.(int); ok && limit < len(docs) {
					docs = docs[:limit]
				}
			}
		}
	}
	return docs
}

// executeGroup 执行 $group 阶段
func executeGroup(docs []Document, groupSpec Document) []Document {
	groupKey, _ := groupSpec["_id"].(string)
	groups := make(map[string][]Document)
	for _, doc := range docs {
		var key string
		if strings.HasPrefix(groupKey, "$") {
			fieldName := groupKey[1:]
			key = fmt.Sprintf("%v", doc[fieldName])
		} else {
			key = groupKey
		}
		groups[key] = append(groups[key], doc)
	}
	result := make([]Document, 0, len(groups))
	for key, groupDocs := range groups {
		grouped := Document{"_id": key}
		for field, aggSpec := range groupSpec {
			if field == "_id" {
				continue
			}
			if spec, ok := aggSpec.(Document); ok {
				for aggOp, aggField := range spec {
					fieldName := ""
					if s, ok := aggField.(string); ok && strings.HasPrefix(s, "$") {
						fieldName = s[1:]
					}
					switch aggOp {
					case "$sum":
						if aggField == 1 {
							grouped[field] = len(groupDocs)
						} else {
							var sum float64
							for _, d := range groupDocs {
								sum += toFloat(d[fieldName])
							}
							grouped[field] = sum
						}
					case "$avg":
						var sum float64
						for _, d := range groupDocs {
							sum += toFloat(d[fieldName])
						}
						grouped[field] = sum / float64(len(groupDocs))
					case "$max":
						maxVal := toFloat(groupDocs[0][fieldName])
						for _, d := range groupDocs[1:] {
							if v := toFloat(d[fieldName]); v > maxVal {
								maxVal = v
							}
						}
						grouped[field] = maxVal
					}
				}
			}
		}
		result = append(result, grouped)
	}
	return result
}

// executeSort 执行 $sort 阶段
func executeSort(docs []Document, sortSpec Document) {
	for field, order := range sortSpec {
		f := field
		asc := toFloat(order) > 0
		sort.SliceStable(docs, func(i, j int) bool {
			vi := toFloat(docs[i][f])
			vj := toFloat(docs[j][f])
			if asc {
				return vi < vj
			}
			return vi > vj
		})
	}
}

// ============================================================
// Part A 演示函数
// ============================================================

// demoDocumentModel 演示文档模型与 BSON 概念
func demoDocumentModel(db *MemoryDocStore) {
	fmt.Println("\n--- 1. 文档模型与 BSON ---")

	// 创建一个包含嵌套文档和数组的复杂文档
	userDoc := Document{
		"name":  "张三",
		"email": "zhangsan@example.com",
		"age":   28,
		"tags":  []string{"Go", "后端", "云原生"},
		"address": Document{
			"city":     "北京",
			"district": "朝阳区",
			"zipcode":  "100020",
		},
		"skills": []Document{
			{"name": "Go", "level": "高级", "years": 5},
			{"name": "Python", "level": "中级", "years": 3},
		},
		"created_at": time.Now().Format(time.RFC3339),
	}

	// 序列化为 JSON 展示（模拟 BSON 的可读形式）
	jsonBytes, _ := json.MarshalIndent(userDoc, "  ", "  ")
	fmt.Printf("  文档结构（JSON 表示）:\n  %s\n", string(jsonBytes))

	fmt.Println("\n  BSON vs JSON 对比:")
	fmt.Println("    BSON 额外支持的类型: ObjectID, Date, Binary, Decimal128, Regex")
	fmt.Println("    BSON 优势: 二进制编码更紧凑，支持按字段定位，解析更快")
	fmt.Println("    Go 中的 BSON 类型:")
	fmt.Println("      bson.D — 有序文档（用于命令和索引，保持字段顺序）")
	fmt.Println("      bson.M — 无序文档（用于普通查询，map 类型）")
	fmt.Println("      bson.A — 数组")
	fmt.Println("      bson.E — 单个键值对元素")
}

// demoCRUDOperations 演示 CRUD 操作
func demoCRUDOperations(db *MemoryDocStore) {
	fmt.Println("\n--- 2. CRUD 操作 ---")

	// --- InsertOne ---
	fmt.Println("\n  [InsertOne] 插入单个文档:")
	id, _ := db.InsertOne("users", Document{
		"name": "张三", "email": "zhangsan@example.com",
		"age": 28, "city": "北京", "role": "developer",
	})
	fmt.Printf("    插入成功, _id=%s\n", id)

	// --- InsertMany ---
	fmt.Println("\n  [InsertMany] 批量插入文档:")
	users := []Document{
		{"name": "李四", "email": "lisi@example.com", "age": 32, "city": "上海", "role": "developer"},
		{"name": "王五", "email": "wangwu@example.com", "age": 25, "city": "北京", "role": "designer"},
		{"name": "赵六", "email": "zhaoliu@example.com", "age": 35, "city": "深圳", "role": "developer"},
		{"name": "孙七", "email": "sunqi@example.com", "age": 29, "city": "上海", "role": "manager"},
		{"name": "周八", "email": "zhouba@example.com", "age": 27, "city": "北京", "role": "developer"},
	}
	ids, _ := db.InsertMany("users", users)
	fmt.Printf("    批量插入 %d 条文档\n", len(ids))

	// --- FindOne ---
	fmt.Println("\n  [FindOne] 查询单个文档:")
	if doc, found := db.FindOne("users", Document{"name": "张三"}); found {
		fmt.Printf("    找到: name=%s, age=%v, city=%s\n", doc["name"], doc["age"], doc["city"])
	}

	// --- Find（条件查询） ---
	fmt.Println("\n  [Find] 条件查询 — 北京的用户:")
	beijingUsers := db.Find("users", Document{"city": "北京"})
	for _, u := range beijingUsers {
		fmt.Printf("    %s (age=%v, role=%s)\n", u["name"], u["age"], u["role"])
	}

	// --- Find（比较操作符） ---
	fmt.Println("\n  [Find] 比较查询 — 年龄 >= 30 的用户:")
	seniorUsers := db.Find("users", Document{
		"age": Document{"$gte": 30},
	})
	for _, u := range seniorUsers {
		fmt.Printf("    %s (age=%v)\n", u["name"], u["age"])
	}

	// --- UpdateOne ---
	fmt.Println("\n  [UpdateOne] 更新文档:")
	modified, _ := db.UpdateOne("users",
		Document{"name": "张三"},
		Document{"$set": Document{"age": 29, "city": "深圳"}},
	)
	fmt.Printf("    更新 %d 条文档\n", modified)
	if doc, found := db.FindOne("users", Document{"name": "张三"}); found {
		fmt.Printf("    更新后: name=%s, age=%v, city=%s\n", doc["name"], doc["age"], doc["city"])
	}

	// --- UpdateOne（$inc 操作符） ---
	fmt.Println("\n  [UpdateOne] $inc 操作符 — 年龄 +1:")
	db.UpdateOne("users",
		Document{"name": "李四"},
		Document{"$inc": Document{"age": 1}},
	)
	if doc, found := db.FindOne("users", Document{"name": "李四"}); found {
		fmt.Printf("    李四 age=%v\n", doc["age"])
	}

	// --- DeleteOne ---
	fmt.Println("\n  [DeleteOne] 删除文档:")
	deleted, _ := db.DeleteOne("users", Document{"name": "孙七"})
	fmt.Printf("    删除 %d 条文档\n", deleted)
	fmt.Printf("    剩余文档数: %d\n", db.Count("users", Document{}))
}

// demoAggregationPipeline 演示聚合管道
func demoAggregationPipeline(db *MemoryDocStore) {
	fmt.Println("\n--- 3. 聚合管道 ---")

	// 按城市分组统计
	fmt.Println("\n  [聚合] 按城市分组统计用户数和平均年龄:")
	pipeline := []Document{
		{"$group": Document{
			"_id":    "$city",
			"count":  Document{"$sum": 1},
			"avgAge": Document{"$avg": "$age"},
			"maxAge": Document{"$max": "$age"},
		}},
		{"$sort": Document{"count": -1}},
	}
	results := db.Aggregate("users", pipeline)
	for _, r := range results {
		fmt.Printf("    城市: %s, 人数: %v, 平均年龄: %.1f, 最大年龄: %.0f\n",
			r["_id"], r["count"], r["avgAge"], r["maxAge"])
	}

	// 按角色分组统计
	fmt.Println("\n  [聚合] 按角色分组统计:")
	pipeline2 := []Document{
		{"$group": Document{
			"_id":   "$role",
			"count": Document{"$sum": 1},
		}},
		{"$sort": Document{"count": -1}},
	}
	results2 := db.Aggregate("users", pipeline2)
	for _, r := range results2 {
		fmt.Printf("    角色: %s, 人数: %v\n", r["_id"], r["count"])
	}

	// 带过滤的聚合
	fmt.Println("\n  [聚合] 先过滤再分组 — 只统计 developer:")
	pipeline3 := []Document{
		{"$match": Document{"role": "developer"}},
		{"$group": Document{
			"_id":    "$city",
			"count":  Document{"$sum": 1},
			"avgAge": Document{"$avg": "$age"},
		}},
	}
	results3 := db.Aggregate("users", pipeline3)
	for _, r := range results3 {
		fmt.Printf("    城市: %s, developer 人数: %v, 平均年龄: %.1f\n",
			r["_id"], r["count"], r["avgAge"])
	}

	fmt.Println("\n  聚合管道 vs SQL 对照:")
	fmt.Println("    $match  → WHERE")
	fmt.Println("    $group  → GROUP BY")
	fmt.Println("    $sort   → ORDER BY")
	fmt.Println("    $limit  → LIMIT")
	fmt.Println("    $lookup → LEFT JOIN")
	fmt.Println("    $unwind → 展开数组（SQL 无对应）")
}

// demoIndexManagement 演示索引管理
func demoIndexManagement(db *MemoryDocStore) {
	fmt.Println("\n--- 4. 索引管理 ---")

	// 创建单字段索引
	name1 := db.CreateIndex("users", []string{"email"}, true)
	fmt.Printf("  创建唯一索引: %s (unique=true) ✅\n", name1)

	// 创建复合索引
	name2 := db.CreateIndex("users", []string{"city", "role"}, false)
	fmt.Printf("  创建复合索引: %s ✅\n", name2)

	// 创建 TTL 索引（概念演示）
	name3 := db.CreateIndex("sessions", []string{"expire_at"}, false)
	fmt.Printf("  创建 TTL 索引: %s (expireAfterSeconds=0) ✅\n", name3)

	// 列出所有索引
	fmt.Println("\n  users 集合的索引:")
	for _, idx := range db.ListIndexes("users") {
		uniqueStr := ""
		if idx.Unique {
			uniqueStr = " [unique]"
		}
		fmt.Printf("    %s: fields=%v%s\n", idx.Name, idx.Fields, uniqueStr)
	}

	fmt.Println("\n  索引最佳实践:")
	fmt.Println("    1. ESR 原则: 复合索引字段顺序 Equality → Sort → Range")
	fmt.Println("    2. 覆盖索引: 查询字段全在索引中，避免回表")
	fmt.Println("    3. 使用 explain() 分析查询计划")
	fmt.Println("    4. 避免低选择性字段单独建索引（如布尔字段）")
	fmt.Println("    5. TTL 索引自动清理过期数据（会话、日志等）")
}

// demoTransactionConcept 演示事务概念
func demoTransactionConcept(db *MemoryDocStore) {
	fmt.Println("\n--- 5. 事务 ---")

	// 初始化账户数据
	db.InsertOne("accounts", Document{"_id": "account_A", "name": "账户A", "balance": float64(1000)})
	db.InsertOne("accounts", Document{"_id": "account_B", "name": "账户B", "balance": float64(500)})

	fmt.Println("  转账前:")
	if a, found := db.FindOne("accounts", Document{"_id": "account_A"}); found {
		fmt.Printf("    账户A 余额: %.0f\n", a["balance"])
	}
	if b, found := db.FindOne("accounts", Document{"_id": "account_B"}); found {
		fmt.Printf("    账户B 余额: %.0f\n", b["balance"])
	}

	// 模拟事务：A 转账 200 给 B
	fmt.Println("\n  执行事务: A 转账 200 给 B")
	transferAmount := float64(200)

	// 检查余额
	accountA, _ := db.FindOne("accounts", Document{"_id": "account_A"})
	if toFloat(accountA["balance"]) < transferAmount {
		fmt.Println("    事务失败: 余额不足")
		return
	}

	// 原子操作（模拟事务中的两个操作）
	db.UpdateOne("accounts",
		Document{"_id": "account_A"},
		Document{"$inc": Document{"balance": -transferAmount}},
	)
	db.UpdateOne("accounts",
		Document{"_id": "account_B"},
		Document{"$inc": Document{"balance": transferAmount}},
	)

	fmt.Println("  转账后:")
	if a, found := db.FindOne("accounts", Document{"_id": "account_A"}); found {
		fmt.Printf("    账户A 余额: %.0f\n", a["balance"])
	}
	if b, found := db.FindOne("accounts", Document{"_id": "account_B"}); found {
		fmt.Printf("    账户B 余额: %.0f\n", b["balance"])
	}

	fmt.Println("\n  MongoDB 事务要点:")
	fmt.Println("    1. MongoDB 4.0+ 支持副本集多文档事务")
	fmt.Println("    2. MongoDB 4.2+ 支持分片集群多文档事务")
	fmt.Println("    3. 事务有 60 秒超时限制")
	fmt.Println("    4. 优先通过文档模型设计避免事务（嵌套文档替代关联表）")
	fmt.Println("    5. mongo-go-driver 使用 session.WithTransaction() 自动处理重试")
}

// ============================================================
// Part B：连接真实 MongoDB（需要 Docker）
// ============================================================

func partB() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Part B：连接真实 MongoDB（mongo-go-driver）")
	fmt.Println(strings.Repeat("=", 60))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 创建 MongoDB 客户端
	clientOpts := options.Client().ApplyURI("mongodb://root:root123@localhost:27017")
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		fmt.Printf("❌ 创建 MongoDB 客户端失败: %v\n", err)
		return
	}
	defer client.Disconnect(ctx)

	// 测试连接
	if err := client.Ping(ctx, nil); err != nil {
		fmt.Printf("❌ 无法连接 MongoDB: %v\n", err)
		fmt.Println("请先启动 MongoDB: docker compose -f docker/docker-compose.yml up -d mongodb")
		return
	}
	fmt.Println("✅ MongoDB 连接成功")

	db := client.Database("guide_go_demo")

	// 1. CRUD 操作
	demoRealCRUD(ctx, db)

	// 2. 聚合管道
	demoRealAggregation(ctx, db)

	// 3. 索引管理
	demoRealIndexes(ctx, db)

	// 4. 事务
	demoRealTransaction(ctx, client, db)

	// 清理测试数据
	db.Drop(ctx)
	fmt.Println("\n  测试数据库已清理 ✅")
}

// demoRealCRUD 演示真实 CRUD 操作
func demoRealCRUD(ctx context.Context, db *mongo.Database) {
	fmt.Println("\n--- 1. CRUD 操作（mongo-go-driver） ---")

	coll := db.Collection("users")
	coll.Drop(ctx) // 清理旧数据

	// InsertOne — 插入单个文档
	fmt.Println("\n  [InsertOne]")
	result, err := coll.InsertOne(ctx, bson.D{
		{Key: "name", Value: "张三"},
		{Key: "email", Value: "zhangsan@example.com"},
		{Key: "age", Value: 28},
		{Key: "city", Value: "北京"},
		{Key: "role", Value: "developer"},
		{Key: "tags", Value: bson.A{"Go", "后端", "云原生"}},
		{Key: "created_at", Value: time.Now()},
	})
	if err != nil {
		fmt.Printf("  插入失败: %v\n", err)
		return
	}
	fmt.Printf("  插入成功, _id=%v\n", result.InsertedID)

	// InsertMany — 批量插入
	fmt.Println("\n  [InsertMany]")
	docs := []interface{}{
		bson.D{
			{Key: "name", Value: "李四"}, {Key: "email", Value: "lisi@example.com"},
			{Key: "age", Value: 32}, {Key: "city", Value: "上海"}, {Key: "role", Value: "developer"},
			{Key: "tags", Value: bson.A{"Go", "微服务"}}, {Key: "created_at", Value: time.Now()},
		},
		bson.D{
			{Key: "name", Value: "王五"}, {Key: "email", Value: "wangwu@example.com"},
			{Key: "age", Value: 25}, {Key: "city", Value: "北京"}, {Key: "role", Value: "designer"},
			{Key: "tags", Value: bson.A{"UI", "Figma"}}, {Key: "created_at", Value: time.Now()},
		},
		bson.D{
			{Key: "name", Value: "赵六"}, {Key: "email", Value: "zhaoliu@example.com"},
			{Key: "age", Value: 35}, {Key: "city", Value: "深圳"}, {Key: "role", Value: "developer"},
			{Key: "tags", Value: bson.A{"Go", "K8s", "分布式"}}, {Key: "created_at", Value: time.Now()},
		},
		bson.D{
			{Key: "name", Value: "周八"}, {Key: "email", Value: "zhouba@example.com"},
			{Key: "age", Value: 27}, {Key: "city", Value: "北京"}, {Key: "role", Value: "developer"},
			{Key: "tags", Value: bson.A{"Go", "gRPC"}}, {Key: "created_at", Value: time.Now()},
		},
	}
	manyResult, err := coll.InsertMany(ctx, docs)
	if err != nil {
		fmt.Printf("  批量插入失败: %v\n", err)
		return
	}
	fmt.Printf("  批量插入 %d 条文档\n", len(manyResult.InsertedIDs))

	// FindOne — 查询单个文档
	fmt.Println("\n  [FindOne]")
	var user bson.M
	err = coll.FindOne(ctx, bson.D{{Key: "name", Value: "张三"}}).Decode(&user)
	if err != nil {
		fmt.Printf("  查询失败: %v\n", err)
		return
	}
	fmt.Printf("  找到: name=%s, age=%v, city=%s, tags=%v\n",
		user["name"], user["age"], user["city"], user["tags"])

	// Find — 条件查询
	fmt.Println("\n  [Find] 北京的用户:")
	cursor, err := coll.Find(ctx, bson.D{{Key: "city", Value: "北京"}})
	if err != nil {
		fmt.Printf("  查询失败: %v\n", err)
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var u bson.M
		cursor.Decode(&u)
		fmt.Printf("    %s (age=%v, role=%s)\n", u["name"], u["age"], u["role"])
	}

	// Find — 比较操作符
	fmt.Println("\n  [Find] 年龄 >= 30 的用户:")
	cursor2, _ := coll.Find(ctx, bson.D{
		{Key: "age", Value: bson.D{{Key: "$gte", Value: 30}}},
	})
	defer cursor2.Close(ctx)
	for cursor2.Next(ctx) {
		var u bson.M
		cursor2.Decode(&u)
		fmt.Printf("    %s (age=%v)\n", u["name"], u["age"])
	}

	// UpdateOne — 更新文档
	fmt.Println("\n  [UpdateOne]")
	updateResult, err := coll.UpdateOne(ctx,
		bson.D{{Key: "name", Value: "张三"}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "age", Value: 29},
			{Key: "city", Value: "深圳"},
		}}},
	)
	if err != nil {
		fmt.Printf("  更新失败: %v\n", err)
		return
	}
	fmt.Printf("  匹配 %d 条, 修改 %d 条\n", updateResult.MatchedCount, updateResult.ModifiedCount)

	// DeleteOne — 删除文档
	fmt.Println("\n  [DeleteOne]")
	delResult, err := coll.DeleteOne(ctx, bson.D{{Key: "name", Value: "王五"}})
	if err != nil {
		fmt.Printf("  删除失败: %v\n", err)
		return
	}
	fmt.Printf("  删除 %d 条文档\n", delResult.DeletedCount)

	// Count — 统计
	count, _ := coll.CountDocuments(ctx, bson.D{})
	fmt.Printf("\n  当前文档总数: %d\n", count)
}

// demoRealAggregation 演示真实聚合管道
func demoRealAggregation(ctx context.Context, db *mongo.Database) {
	fmt.Println("\n--- 2. 聚合管道（mongo-go-driver） ---")

	coll := db.Collection("users")

	// 按城市分组统计
	fmt.Println("\n  [聚合] 按城市分组统计:")
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$city"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "avgAge", Value: bson.D{{Key: "$avg", Value: "$age"}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
	}
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		fmt.Printf("  聚合失败: %v\n", err)
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var result bson.M
		cursor.Decode(&result)
		fmt.Printf("    城市: %s, 人数: %v, 平均年龄: %.1f\n",
			result["_id"], result["count"], result["avgAge"])
	}

	// 带过滤的聚合 — 只统计 developer
	fmt.Println("\n  [聚合] developer 按城市分布:")
	pipeline2 := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "role", Value: "developer"}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$city"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
	}
	cursor2, _ := coll.Aggregate(ctx, pipeline2)
	defer cursor2.Close(ctx)
	for cursor2.Next(ctx) {
		var result bson.M
		cursor2.Decode(&result)
		fmt.Printf("    城市: %s, developer 人数: %v\n", result["_id"], result["count"])
	}
}

// demoRealIndexes 演示真实索引管理
func demoRealIndexes(ctx context.Context, db *mongo.Database) {
	fmt.Println("\n--- 3. 索引管理（mongo-go-driver） ---")

	coll := db.Collection("users")

	// 创建唯一索引
	indexName, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		fmt.Printf("  创建唯一索引失败: %v\n", err)
	} else {
		fmt.Printf("  创建唯一索引: %s ✅\n", indexName)
	}

	// 创建复合索引
	indexName2, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "city", Value: 1},
			{Key: "role", Value: 1},
		},
	})
	if err != nil {
		fmt.Printf("  创建复合索引失败: %v\n", err)
	} else {
		fmt.Printf("  创建复合索引: %s ✅\n", indexName2)
	}

	// 列出所有索引
	fmt.Println("\n  users 集合的索引:")
	indexCursor, _ := coll.Indexes().List(ctx)
	defer indexCursor.Close(ctx)
	for indexCursor.Next(ctx) {
		var idx bson.M
		indexCursor.Decode(&idx)
		fmt.Printf("    %s: keys=%v\n", idx["name"], idx["key"])
	}

	// 演示 explain 查询计划
	fmt.Println("\n  查询优化提示:")
	fmt.Println("    使用 explain() 分析查询是否命中索引:")
	fmt.Println("    db.users.find({city:'北京'}).explain('executionStats')")
	fmt.Println("    关注 winningPlan.stage: IXSCAN(索引扫描) vs COLLSCAN(全集合扫描)")
}

// demoRealTransaction 演示真实事务
func demoRealTransaction(ctx context.Context, client *mongo.Client, db *mongo.Database) {
	fmt.Println("\n--- 4. 事务（mongo-go-driver） ---")

	// 注意：MongoDB 事务需要副本集，单节点 Docker 默认不支持
	// 这里演示事务 API 的使用方式
	fmt.Println("  ⚠️  MongoDB 事务需要副本集环境")
	fmt.Println("  单节点 Docker 不支持事务，以下展示 API 用法:")

	fmt.Println(`
  // 事务示例代码（需要副本集环境）:
  session, err := client.StartSession()
  defer session.EndSession(ctx)

  _, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
      accountsColl := db.Collection("accounts")
      
      // 扣减账户 A
      _, err := accountsColl.UpdateOne(sc,
          bson.D{{Key: "_id", Value: "A"}},
          bson.D{{Key: "$inc", Value: bson.D{{Key: "balance", Value: -100}}}},
      )
      if err != nil {
          return nil, err // 自动回滚
      }
      
      // 增加账户 B
      _, err = accountsColl.UpdateOne(sc,
          bson.D{{Key: "_id", Value: "B"}},
          bson.D{{Key: "$inc", Value: bson.D{{Key: "balance", Value: 100}}}},
      )
      return nil, err // 成功则自动提交
  })`)

	// 演示非事务的单文档原子操作（FindOneAndUpdate）
	fmt.Println("\n  [FindOneAndUpdate] 单文档原子操作（不需要事务）:")
	coll := db.Collection("users")
	var updated bson.M
	err := coll.FindOneAndUpdate(ctx,
		bson.D{{Key: "name", Value: "张三"}},
		bson.D{{Key: "$inc", Value: bson.D{{Key: "age", Value: 1}}}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)
	if err != nil {
		fmt.Printf("  原子更新失败: %v\n", err)
		return
	}
	fmt.Printf("  原子更新后: name=%s, age=%v\n", updated["name"], updated["age"])
}

// ============================================================
// main 入口
// ============================================================

func main() {
	// Part A：纯内存模拟，直接运行理解原理
	partA()

	// Part B：连接真实 MongoDB，需传入参数 'real'
	if len(os.Args) > 1 && os.Args[1] == "real" {
		partB()
	}
}
