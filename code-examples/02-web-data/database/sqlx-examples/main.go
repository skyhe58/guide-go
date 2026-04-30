// sqlx 示例
// 演示：结构体映射、命名参数、Get/Select、批量操作、IN 子句、事务
// Go 1.22+ | 验证日期 2025-01-01
//
// Part A：纯内存模拟，直接运行理解 sqlx 核心概念
// Part B：连接真实 MySQL，需传入参数 'real'
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

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// ============================================================
// 数据模型（sqlx 使用 db 标签）
// ============================================================

// User 用户模型 — 使用 db 标签映射数据库列名
type User struct {
	ID        int       `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Age       int       `db:"age"`
	CreatedAt time.Time `db:"created_at"`
}

// ============================================================
// Part A：纯内存模拟 sqlx 核心概念
// ============================================================

func partA() {
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println("Part A：sqlx 核心概念（内存模拟）")
	fmt.Println("=" + strings.Repeat("=", 59))

	// 1. 结构体映射
	demoStructMapping()

	// 2. Get 和 Select
	demoGetSelect()

	// 3. 命名参数
	demoNamedParams()

	// 4. 批量操作
	demoBatchOperations()

	// 5. IN 子句展开
	demoInClause()

	// 6. 事务
	demoTransaction()

	// 7. sqlx vs database/sql 对比
	demoComparison()
}

// demoStructMapping 演示结构体映射
func demoStructMapping() {
	fmt.Println("\n--- 1. 结构体映射 ---")

	fmt.Println("sqlx 通过 db 标签将数据库列名映射到结构体字段：")
	fmt.Println(`  type User struct {`)
	fmt.Println(`      ID        int       ` + "`db:\"id\"`")
	fmt.Println(`      Name      string    ` + "`db:\"name\"`")
	fmt.Println(`      Email     string    ` + "`db:\"email\"`")
	fmt.Println(`      CreatedAt time.Time ` + "`db:\"created_at\"`")
	fmt.Println(`  }`)

	fmt.Println("\n映射原理：")
	fmt.Println("  1. 解析结构体的 db 标签")
	fmt.Println("  2. 构建列名到字段的映射表（内部缓存）")
	fmt.Println("  3. 查询结果按列名自动 Scan 到对应字段")
	fmt.Println("  4. 没有 db 标签时，使用字段名的小写形式匹配")

	// 模拟映射
	user := User{ID: 1, Name: "张三", Email: "zhangsan@example.com", Age: 25, CreatedAt: time.Now()}
	fmt.Printf("\n模拟映射结果: %+v\n", user)
}

// demoGetSelect 演示 Get 和 Select
func demoGetSelect() {
	fmt.Println("\n--- 2. Get 和 Select ---")

	// 模拟数据
	users := []User{
		{ID: 1, Name: "张三", Email: "zhangsan@example.com", Age: 25},
		{ID: 2, Name: "李四", Email: "lisi@example.com", Age: 30},
		{ID: 3, Name: "王五", Email: "wangwu@example.com", Age: 28},
	}

	fmt.Println("Get — 查询单行，映射到单个结构体：")
	fmt.Println("  var user User")
	fmt.Println("  err = db.Get(&user, \"SELECT * FROM users WHERE id = ?\", 1)")
	fmt.Printf("  结果: %+v\n", users[0])

	fmt.Println("\nSelect — 查询多行，映射到结构体切片：")
	fmt.Println("  var users []User")
	fmt.Println("  err = db.Select(&users, \"SELECT * FROM users WHERE age > ?\", 25)")
	fmt.Println("  结果:")
	for _, u := range users {
		if u.Age > 25 {
			fmt.Printf("    → ID=%d, Name=%s, Age=%d\n", u.ID, u.Name, u.Age)
		}
	}

	fmt.Println("\n区别：")
	fmt.Println("  Get:    底层调用 QueryRow，无结果返回 sql.ErrNoRows")
	fmt.Println("  Select: 底层调用 Query，无结果返回空切片（不报错）")
}

// demoNamedParams 演示命名参数
func demoNamedParams() {
	fmt.Println("\n--- 3. 命名参数 ---")

	fmt.Println("使用 :name 风格的命名参数：")
	fmt.Println("  query := `INSERT INTO users (name, email) VALUES (:name, :email)`")

	fmt.Println("\n从结构体提取参数：")
	fmt.Println("  user := User{Name: \"张三\", Email: \"zhangsan@example.com\"}")
	fmt.Println("  db.NamedExec(query, user)")

	fmt.Println("\n从 map 提取参数：")
	fmt.Println("  params := map[string]interface{}{")
	fmt.Println("      \"name\":  \"张三\",")
	fmt.Println("      \"email\": \"zhangsan@example.com\",")
	fmt.Println("  }")
	fmt.Println("  db.NamedExec(query, params)")

	// 模拟命名参数解析
	user := User{Name: "张三", Email: "zhangsan@example.com"}
	fmt.Printf("\n模拟: INSERT INTO users (name, email) VALUES ('%s', '%s')\n", user.Name, user.Email)
}

// demoBatchOperations 演示批量操作
func demoBatchOperations() {
	fmt.Println("\n--- 4. 批量操作 ---")

	users := []User{
		{Name: "用户A", Email: "a@example.com"},
		{Name: "用户B", Email: "b@example.com"},
		{Name: "用户C", Email: "c@example.com"},
	}

	fmt.Println("NamedExec 批量插入：")
	fmt.Println("  users := []User{{Name: \"A\"}, {Name: \"B\"}, {Name: \"C\"}}")
	fmt.Println("  db.NamedExec(`INSERT INTO users (name, email) VALUES (:name, :email)`, users)")
	fmt.Printf("  模拟批量插入 %d 条记录\n", len(users))

	for _, u := range users {
		fmt.Printf("    → INSERT (name='%s', email='%s')\n", u.Name, u.Email)
	}
}

// demoInClause 演示 IN 子句展开
func demoInClause() {
	fmt.Println("\n--- 5. IN 子句展开 ---")

	ids := []int{1, 3, 5, 7}

	fmt.Println("sqlx.In 展开 IN 子句：")
	fmt.Println("  query, args, err := sqlx.In(\"SELECT * FROM users WHERE id IN (?)\", []int{1, 3, 5, 7})")
	fmt.Println("  query = db.Rebind(query)  // 转换占位符风格")
	fmt.Println("  db.Select(&users, query, args...)")

	fmt.Printf("\n模拟展开: SELECT * FROM users WHERE id IN (%s)\n",
		strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ids)), ","), "[]"))

	fmt.Println("\n⚠️ 注意：使用 sqlx.In 后必须调用 db.Rebind 转换占位符")
	fmt.Println("  MySQL:      ? → ?")
	fmt.Println("  PostgreSQL: ? → $1, $2, ...")
}

// demoTransaction 演示事务
func demoTransaction() {
	fmt.Println("\n--- 6. 事务 ---")

	fmt.Println("sqlx 事务（与 database/sql 类似，增加命名参数支持）：")
	fmt.Println("  tx, err := db.Beginx()")
	fmt.Println("  defer tx.Rollback()")
	fmt.Println()
	fmt.Println("  tx.NamedExec(`UPDATE accounts SET balance = balance - :amount WHERE id = :id`,")
	fmt.Println("      map[string]interface{}{\"amount\": 100, \"id\": 1})")
	fmt.Println()
	fmt.Println("  tx.NamedExec(`UPDATE accounts SET balance = balance + :amount WHERE id = :id`,")
	fmt.Println("      map[string]interface{}{\"amount\": 100, \"id\": 2})")
	fmt.Println()
	fmt.Println("  tx.Commit()")

	// 模拟转账
	fmt.Println("\n模拟转账事务：")
	fmt.Println("  账户A: 1000 → 900 (-100)")
	fmt.Println("  账户B: 500  → 600 (+100)")
	fmt.Println("  事务提交成功 ✅")
}

// demoComparison 演示 sqlx vs database/sql 对比
func demoComparison() {
	fmt.Println("\n--- 7. sqlx vs database/sql 对比 ---")

	fmt.Println("查询多行对比：")
	fmt.Println()
	fmt.Println("database/sql（手动 Scan）：")
	fmt.Println("  rows, _ := db.Query(\"SELECT id, name, email FROM users\")")
	fmt.Println("  defer rows.Close()")
	fmt.Println("  for rows.Next() {")
	fmt.Println("      var u User")
	fmt.Println("      rows.Scan(&u.ID, &u.Name, &u.Email)  // 逐字段 Scan")
	fmt.Println("      users = append(users, u)")
	fmt.Println("  }")
	fmt.Println()
	fmt.Println("sqlx（自动映射）：")
	fmt.Println("  var users []User")
	fmt.Println("  db.Select(&users, \"SELECT id, name, email FROM users\")  // 一行搞定")
	fmt.Println()
	fmt.Println("结论：sqlx 减少了 80% 的样板代码，同时保持 SQL 完全可控")
}

// ============================================================
// Part B：连接真实 MySQL
// ============================================================

func partB() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Part B：sqlx 连接真实 MySQL")
	fmt.Println(strings.Repeat("=", 60))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 连接 MySQL
	dsn := "root:root123@tcp(localhost:3306)/testdb?parseTime=true&charset=utf8mb4"
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		fmt.Printf("MySQL 连接失败: %v\n", err)
		fmt.Println("请确保已启动: docker compose -f docker/docker-compose.yml up -d mysql")
		return
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	fmt.Println("✅ MySQL 连接成功")

	// 创建表
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS sqlx_users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(255),
			age INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		fmt.Printf("创建表失败: %v\n", err)
		return
	}
	fmt.Println("  创建表 sqlx_users 成功")

	// NamedExec 插入
	users := []User{
		{Name: "sqlx张三", Email: "sqlx_zhangsan@example.com", Age: 25},
		{Name: "sqlx李四", Email: "sqlx_lisi@example.com", Age: 30},
		{Name: "sqlx王五", Email: "sqlx_wangwu@example.com", Age: 28},
	}

	for _, u := range users {
		_, err = db.NamedExecContext(ctx,
			"INSERT INTO sqlx_users (name, email, age) VALUES (:name, :email, :age)", u)
		if err != nil {
			fmt.Printf("  插入失败: %v\n", err)
			continue
		}
	}
	fmt.Printf("  批量插入 %d 条记录成功\n", len(users))

	// Get 查询单行
	var user User
	err = db.GetContext(ctx, &user, "SELECT id, name, email, age, created_at FROM sqlx_users WHERE name = ?", "sqlx张三")
	if err != nil {
		fmt.Printf("  Get 查询失败: %v\n", err)
	} else {
		fmt.Printf("  Get 查询: %+v\n", user)
	}

	// Select 查询多行
	var allUsers []User
	err = db.SelectContext(ctx, &allUsers, "SELECT id, name, email, age, created_at FROM sqlx_users WHERE age >= ?", 25)
	if err != nil {
		fmt.Printf("  Select 查询失败: %v\n", err)
	} else {
		fmt.Printf("  Select 查询: 找到 %d 条记录\n", len(allUsers))
		for _, u := range allUsers {
			fmt.Printf("    → ID=%d, Name=%s, Age=%d\n", u.ID, u.Name, u.Age)
		}
	}

	// IN 子句
	query, args, err := sqlx.In("SELECT id, name, email, age, created_at FROM sqlx_users WHERE age IN (?)", []int{25, 30})
	if err != nil {
		fmt.Printf("  IN 展开失败: %v\n", err)
	} else {
		query = db.Rebind(query)
		var inUsers []User
		err = db.SelectContext(ctx, &inUsers, query, args...)
		if err != nil {
			fmt.Printf("  IN 查询失败: %v\n", err)
		} else {
			fmt.Printf("  IN 查询 (age IN [25, 30]): 找到 %d 条\n", len(inUsers))
		}
	}

	// 事务
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		fmt.Printf("  开启事务失败: %v\n", err)
	} else {
		_, err = tx.NamedExecContext(ctx,
			"UPDATE sqlx_users SET age = age + 1 WHERE name = :name",
			map[string]interface{}{"name": "sqlx张三"})
		if err != nil {
			tx.Rollback()
			fmt.Printf("  事务执行失败: %v\n", err)
		} else {
			tx.Commit()
			fmt.Println("  事务提交成功: sqlx张三 age + 1")
		}
	}

	// 清理
	db.ExecContext(ctx, "DROP TABLE IF EXISTS sqlx_users")
	fmt.Println("  ✅ 清理完成")
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
		fmt.Println("💡 运行 'go run main.go real' 可连接真实 MySQL")
		fmt.Println("   前置条件: docker compose -f docker/docker-compose.yml up -d mysql")
	}
}
