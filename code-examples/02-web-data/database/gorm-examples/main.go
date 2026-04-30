// GORM 完整示例
// 演示：模型定义、CRUD、关联关系、Hook、自动迁移、Scope、软删除、日志配置
// Go 1.22+ | 验证日期 2025-01-01
//
// Part A：纯内存模拟，直接运行理解 GORM 核心概念
// Part B：连接真实 MySQL 和 PostgreSQL，需传入参数 'real'
//
// 运行方式：
//   go run main.go          # Part A：内存模拟
//   go run main.go real     # Part B：连接真实数据库
//
// Part B 前置条件：
//   docker compose -f docker/docker-compose.yml up -d mysql postgresql

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ============================================================
// 数据模型（GORM 风格）
// ============================================================

// User 用户模型 — 演示 gorm.Model 内嵌、标签定义、关联关系
type User struct {
	gorm.Model             // 内嵌 ID, CreatedAt, UpdatedAt, DeletedAt（软删除）
	Name     string        `gorm:"size:100;not null;index"`
	Email    string        `gorm:"uniqueIndex;size:255"`
	Age      int           `gorm:"default:0"`
	Role     string        `gorm:"type:varchar(20);default:'user'"`
	Articles []Article     // 一对多关联：一个用户有多篇文章
}

// Article 文章模型 — 演示外键关联、多对多
type Article struct {
	gorm.Model
	Title   string `gorm:"size:200;not null"`
	Content string `gorm:"type:text"`
	UserID  uint   // 外键：关联 User
	User    User   // 关联对象
	Tags    []Tag  `gorm:"many2many:article_tags;"` // 多对多关联
}

// Tag 标签模型 — 演示多对多
type Tag struct {
	gorm.Model
	Name     string    `gorm:"size:50;uniqueIndex"`
	Articles []Article `gorm:"many2many:article_tags;"`
}

// BeforeCreate Hook — 创建前自动处理
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Role == "" {
		u.Role = "user"
	}
	fmt.Printf("  [Hook] BeforeCreate: 用户 %s, 角色设为 %s\n", u.Name, u.Role)
	return nil
}

// AfterCreate Hook — 创建后回调
func (u *User) AfterCreate(tx *gorm.DB) error {
	fmt.Printf("  [Hook] AfterCreate: 用户 %s 创建成功, ID=%d\n", u.Name, u.ID)
	return nil
}

// ============================================================
// Part A：纯内存模拟 GORM 核心概念
// ============================================================

func partA() {
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println("Part A：GORM 核心概念（内存模拟）")
	fmt.Println("=" + strings.Repeat("=", 59))

	// 1. 模型定义概念
	demoModelDefinition()

	// 2. CRUD 操作概念
	demoCRUD()

	// 3. 关联关系概念
	demoAssociations()

	// 4. Hook 概念
	demoHooks()

	// 5. Scope 概念
	demoScopes()

	// 6. 软删除概念
	demoSoftDelete()

	// 7. 日志配置概念
	demoLoggerConfig()
}

// demoModelDefinition 演示模型定义
func demoModelDefinition() {
	fmt.Println("\n--- 1. 模型定义 ---")

	fmt.Println("gorm.Model 内嵌字段：")
	fmt.Println("  ID        uint           `gorm:\"primarykey\"`")
	fmt.Println("  CreatedAt time.Time")
	fmt.Println("  UpdatedAt time.Time")
	fmt.Println("  DeletedAt gorm.DeletedAt `gorm:\"index\"`  // 软删除")

	fmt.Println("\n常用标签：")
	fmt.Println("  gorm:\"size:100\"          → VARCHAR(100)")
	fmt.Println("  gorm:\"not null\"          → NOT NULL 约束")
	fmt.Println("  gorm:\"uniqueIndex\"       → 唯一索引")
	fmt.Println("  gorm:\"index\"             → 普通索引")
	fmt.Println("  gorm:\"default:'user'\"    → 默认值")
	fmt.Println("  gorm:\"type:text\"         → 指定列类型")
	fmt.Println("  gorm:\"many2many:xxx\"     → 多对多关联表名")
}

// demoCRUD 演示 CRUD 操作
func demoCRUD() {
	fmt.Println("\n--- 2. CRUD 操作 ---")

	// 模拟用户数据
	users := []User{
		{Model: gorm.Model{ID: 1}, Name: "张三", Email: "zhangsan@example.com", Age: 25, Role: "admin"},
		{Model: gorm.Model{ID: 2}, Name: "李四", Email: "lisi@example.com", Age: 30, Role: "user"},
		{Model: gorm.Model{ID: 3}, Name: "王五", Email: "wangwu@example.com", Age: 28, Role: "user"},
	}

	// Create
	fmt.Println("\nCreate（创建）：")
	fmt.Println("  db.Create(&user)")
	fmt.Println("  db.Create(&users)  // 批量创建")
	fmt.Printf("  模拟创建 %d 个用户\n", len(users))

	// Read
	fmt.Println("\nRead（查询）：")
	fmt.Println("  db.First(&user, 1)                       // 按主键")
	fmt.Println("  db.Where(\"age > ?\", 18).Find(&users)     // 条件查询")
	fmt.Println("  db.Select(\"name\", \"age\").Find(&users)     // 选择字段")
	fmt.Println("  db.Order(\"age desc\").Limit(10).Find(&users) // 排序分页")

	fmt.Println("  模拟查询 age > 25:")
	for _, u := range users {
		if u.Age > 25 {
			fmt.Printf("    → ID=%d, Name=%s, Age=%d\n", u.ID, u.Name, u.Age)
		}
	}

	// Update
	fmt.Println("\nUpdate（更新）：")
	fmt.Println("  db.Model(&user).Update(\"name\", \"新名字\")           // 单字段")
	fmt.Println("  db.Model(&user).Updates(User{Name: \"新\", Age: 30}) // 多字段（零值不更新）")
	fmt.Println("  db.Model(&user).Updates(map[string]interface{}{\"age\": 0}) // Map（零值也更新）")
	fmt.Println("\n  ⚠️ 零值问题：结构体 Updates 时 0/\"\"/false 会被忽略，需用 map 或指针类型")

	// Delete
	fmt.Println("\nDelete（删除）：")
	fmt.Println("  db.Delete(&user, 1)                // 软删除（有 DeletedAt 字段时）")
	fmt.Println("  db.Unscoped().Delete(&user, 1)     // 物理删除")
}

// demoAssociations 演示关联关系
func demoAssociations() {
	fmt.Println("\n--- 3. 关联关系 ---")

	fmt.Println("一对多（User → Articles）：")
	fmt.Println("  type User struct { Articles []Article }")
	fmt.Println("  type Article struct { UserID uint; User User }")
	fmt.Println("  db.Preload(\"Articles\").Find(&users)  // 预加载")

	fmt.Println("\n多对多（Article ↔ Tags）：")
	fmt.Println("  type Article struct { Tags []Tag `gorm:\"many2many:article_tags;\"` }")
	fmt.Println("  type Tag struct { Articles []Article `gorm:\"many2many:article_tags;\"` }")
	fmt.Println("  db.Preload(\"Tags\").Find(&articles)")

	fmt.Println("\nPreload vs Joins：")
	fmt.Println("  Preload: 分两次查询（先主表再关联表），适合一对多")
	fmt.Println("  Joins:   SQL JOIN 一次查询，适合一对一或需要关联条件过滤")
}

// demoHooks 演示 Hook 钩子
func demoHooks() {
	fmt.Println("\n--- 4. Hook 钩子 ---")

	fmt.Println("Hook 执行顺序：")
	fmt.Println("  Create: BeforeSave → BeforeCreate → SQL → AfterCreate → AfterSave")
	fmt.Println("  Update: BeforeSave → BeforeUpdate → SQL → AfterUpdate → AfterSave")
	fmt.Println("  Delete: BeforeDelete → SQL → AfterDelete")
	fmt.Println("  Query:  AfterFind")

	fmt.Println("\n常见使用场景：")
	fmt.Println("  BeforeCreate: 自动生成 UUID、密码加密、数据校验")
	fmt.Println("  AfterCreate:  发送通知、写入审计日志")
	fmt.Println("  BeforeUpdate: 更新时间戳、数据校验")
	fmt.Println("  AfterDelete:  清理关联数据、发送事件")

	// 模拟 Hook 触发
	fmt.Println("\n模拟 Hook 触发：")
	user := &User{Name: "测试用户"}
	user.BeforeCreate(nil)
}

// demoScopes 演示 Scope 作用域
func demoScopes() {
	fmt.Println("\n--- 5. Scope 作用域 ---")

	fmt.Println("Scope 定义（封装常用查询条件）：")
	fmt.Println(`  func ActiveUsers(db *gorm.DB) *gorm.DB {`)
	fmt.Println(`      return db.Where("status = ?", "active")`)
	fmt.Println(`  }`)
	fmt.Println(``)
	fmt.Println(`  func Paginate(page, size int) func(*gorm.DB) *gorm.DB {`)
	fmt.Println(`      return func(db *gorm.DB) *gorm.DB {`)
	fmt.Println(`          return db.Offset((page-1)*size).Limit(size)`)
	fmt.Println(`      }`)
	fmt.Println(`  }`)

	fmt.Println("\nScope 使用：")
	fmt.Println("  db.Scopes(ActiveUsers, Paginate(1, 10)).Find(&users)")
	fmt.Println("  → SELECT * FROM users WHERE status = 'active' LIMIT 10 OFFSET 0")
}

// demoSoftDelete 演示软删除
func demoSoftDelete() {
	fmt.Println("\n--- 6. 软删除 ---")

	fmt.Println("软删除机制（包含 gorm.DeletedAt 字段的模型）：")
	fmt.Println("  db.Delete(&user, 1)")
	fmt.Println("  → UPDATE users SET deleted_at = NOW() WHERE id = 1")
	fmt.Println()
	fmt.Println("  db.Find(&users)")
	fmt.Println("  → SELECT * FROM users WHERE deleted_at IS NULL  // 自动过滤")
	fmt.Println()
	fmt.Println("  db.Unscoped().Find(&users)")
	fmt.Println("  → SELECT * FROM users  // 包含已删除记录")
	fmt.Println()
	fmt.Println("  db.Unscoped().Delete(&user, 1)")
	fmt.Println("  → DELETE FROM users WHERE id = 1  // 物理删除")
}

// demoLoggerConfig 演示日志配置
func demoLoggerConfig() {
	fmt.Println("\n--- 7. 日志配置 ---")

	fmt.Println("日志级别：")
	fmt.Println("  logger.Silent  → 不输出任何日志")
	fmt.Println("  logger.Error   → 只输出错误")
	fmt.Println("  logger.Warn    → 输出错误和慢查询警告")
	fmt.Println("  logger.Info    → 输出所有 SQL（开发环境推荐）")

	fmt.Println("\n配置示例：")
	fmt.Println(`  db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{`)
	fmt.Println(`      Logger: logger.Default.LogMode(logger.Info),`)
	fmt.Println(`  })`)

	fmt.Println("\n慢查询配置：")
	fmt.Println(`  logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags),`)
	fmt.Println(`      logger.Config{`)
	fmt.Println(`          SlowThreshold: 200 * time.Millisecond,`)
	fmt.Println(`          LogLevel:      logger.Warn,`)
	fmt.Println(`      })`)
}

// ============================================================
// Part B：连接真实 MySQL 和 PostgreSQL
// ============================================================

func partB() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Part B：GORM 连接真实数据库")
	fmt.Println(strings.Repeat("=", 60))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// MySQL 示例
	fmt.Println("\n--- GORM + MySQL ---")
	mysqlDSN := "root:root123@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	mysqlDB, err := gorm.Open(mysql.Open(mysqlDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		fmt.Printf("MySQL 连接失败: %v\n", err)
		fmt.Println("请确保已启动: docker compose -f docker/docker-compose.yml up -d mysql")
	} else {
		runGORMDemo(ctx, mysqlDB, "MySQL")
	}

	// PostgreSQL 示例
	fmt.Println("\n--- GORM + PostgreSQL ---")
	pgDSN := "host=localhost user=postgres password=postgres123 dbname=testdb port=5432 sslmode=disable"
	pgDB, err := gorm.Open(postgres.Open(pgDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		fmt.Printf("PostgreSQL 连接失败: %v\n", err)
		fmt.Println("请确保已启动: docker compose -f docker/docker-compose.yml up -d postgresql")
	} else {
		runGORMDemo(ctx, pgDB, "PostgreSQL")
	}
}

// runGORMDemo 执行 GORM CRUD 示例
func runGORMDemo(_ context.Context, db *gorm.DB, dbType string) {
	// 自动迁移
	err := db.AutoMigrate(&User{}, &Article{}, &Tag{})
	if err != nil {
		fmt.Printf("  迁移失败: %v\n", err)
		return
	}
	fmt.Printf("  ✅ %s AutoMigrate 成功\n", dbType)

	// 创建用户（触发 Hook）
	user := User{Name: "GORM测试用户", Email: fmt.Sprintf("gorm_%d@example.com", time.Now().UnixNano()), Age: 25}
	if err := db.Create(&user).Error; err != nil {
		fmt.Printf("  创建用户失败: %v\n", err)
		return
	}
	fmt.Printf("  创建用户成功: ID=%d, Name=%s\n", user.ID, user.Name)

	// 创建文章（关联用户）
	article := Article{Title: "GORM 入门教程", Content: "GORM 是 Go 最流行的 ORM", UserID: user.ID}
	db.Create(&article)
	fmt.Printf("  创建文章成功: ID=%d, Title=%s\n", article.ID, article.Title)

	// 预加载查询
	var queryUser User
	db.Preload("Articles").First(&queryUser, user.ID)
	fmt.Printf("  预加载查询: 用户 %s 有 %d 篇文章\n", queryUser.Name, len(queryUser.Articles))

	// Scope 查询
	var users []User
	db.Scopes(func(db *gorm.DB) *gorm.DB {
		return db.Where("age >= ?", 18)
	}).Find(&users)
	fmt.Printf("  Scope 查询: 找到 %d 个成年用户\n", len(users))

	// 软删除
	db.Delete(&user)
	var count int64
	db.Model(&User{}).Count(&count)
	fmt.Printf("  软删除后用户数: %d\n", count)

	// 清理（物理删除）
	db.Unscoped().Where("1 = 1").Delete(&Article{})
	db.Unscoped().Where("1 = 1").Delete(&User{})

	// 删除表
	db.Migrator().DropTable(&Article{}, &Tag{}, &User{}, "article_tags")
	fmt.Printf("  ✅ %s 清理完成\n", dbType)
}

// ============================================================
// 主函数
// ============================================================

func main() {
	// Part A：纯内存模拟，直接运行
	partA()

	// Part B：连接真实数据库，需传入参数 'real'
	if len(os.Args) > 1 && os.Args[1] == "real" {
		partB()
	} else {
		fmt.Println("\n" + strings.Repeat("-", 60))
		fmt.Println("💡 运行 'go run main.go real' 可连接真实 MySQL/PostgreSQL")
		fmt.Println("   前置条件: docker compose -f docker/docker-compose.yml up -d mysql postgresql")
	}
}
