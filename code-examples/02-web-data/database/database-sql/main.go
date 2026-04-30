// database/sql 标准库 CRUD 示例
// 演示：连接池配置、CRUD 操作、预编译语句、事务、Null 类型处理、Scanner/Valuer 接口
// Go 1.22+ | 验证日期 2025-01-01
//
// Part A：纯内存模拟，直接运行理解 database/sql 核心概念
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
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// ============================================================
// 数据模型
// ============================================================

// User 用户模型
type User struct {
	ID        int
	Name      string
	Email     string
	Age       int
	CreatedAt time.Time
}

// NullableUser 包含可空字段的用户模型
type NullableUser struct {
	ID    int
	Name  string
	Email sql.NullString // 可空字段
	Age   sql.NullInt64  // 可空字段
}

// ============================================================
// Part A：纯内存模拟 database/sql 核心概念
// ============================================================

func partA() {
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println("Part A：database/sql 核心概念（内存模拟）")
	fmt.Println("=" + strings.Repeat("=", 59))

	// 1. 连接池概念演示
	demoConnectionPool()

	// 2. CRUD 操作概念
	demoCRUDConcepts()

	// 3. 预编译语句概念
	demoPreparedStatement()

	// 4. 事务概念
	demoTransaction()

	// 5. Null 类型处理
	demoNullTypes()

	// 6. Scanner 和 Valuer 接口
	demoScannerValuer()
}

// demoConnectionPool 演示连接池配置概念
func demoConnectionPool() {
	fmt.Println("\n--- 1. 连接池配置概念 ---")

	// 模拟连接池配置参数
	type PoolConfig struct {
		MaxOpenConns    int           // 最大打开连接数
		MaxIdleConns    int           // 最大空闲连接数
		ConnMaxLifetime time.Duration // 连接最大生命周期
		ConnMaxIdleTime time.Duration // 空闲连接最大存活时间
	}

	config := PoolConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 3 * time.Minute,
	}

	fmt.Printf("连接池配置：\n")
	fmt.Printf("  最大打开连接数: %d\n", config.MaxOpenConns)
	fmt.Printf("  最大空闲连接数: %d\n", config.MaxIdleConns)
	fmt.Printf("  连接最大生命周期: %v\n", config.ConnMaxLifetime)
	fmt.Printf("  空闲连接最大存活时间: %v\n", config.ConnMaxIdleTime)

	fmt.Println("\n实际代码：")
	fmt.Println(`  db.SetMaxOpenConns(25)`)
	fmt.Println(`  db.SetMaxIdleConns(10)`)
	fmt.Println(`  db.SetConnMaxLifetime(5 * time.Minute)`)
	fmt.Println(`  db.SetConnMaxIdleTime(3 * time.Minute)`)
}

// demoCRUDConcepts 演示 CRUD 操作概念
func demoCRUDConcepts() {
	fmt.Println("\n--- 2. CRUD 操作概念 ---")

	// 模拟内存数据库
	users := []User{
		{ID: 1, Name: "张三", Email: "zhangsan@example.com", Age: 25, CreatedAt: time.Now()},
		{ID: 2, Name: "李四", Email: "lisi@example.com", Age: 30, CreatedAt: time.Now()},
		{ID: 3, Name: "王五", Email: "wangwu@example.com", Age: 28, CreatedAt: time.Now()},
	}

	// 模拟 QueryRow（查询单行）
	fmt.Println("\nQueryRow - 查询单行：")
	fmt.Printf("  SELECT * FROM users WHERE id = 1 → %+v\n", users[0])

	// 模拟 Query（查询多行）
	fmt.Println("\nQuery - 查询多行：")
	fmt.Println("  SELECT * FROM users WHERE age > 25:")
	for _, u := range users {
		if u.Age > 25 {
			fmt.Printf("    → ID=%d, Name=%s, Age=%d\n", u.ID, u.Name, u.Age)
		}
	}

	// 模拟 Exec（写操作）
	fmt.Println("\nExec - 写操作：")
	newUser := User{ID: 4, Name: "赵六", Email: "zhaoliu@example.com", Age: 35}
	users = append(users, newUser)
	fmt.Printf("  INSERT INTO users ... → LastInsertId=%d, RowsAffected=1\n", newUser.ID)

	// 模拟 Update
	users[0].Name = "张三（已更新）"
	fmt.Printf("  UPDATE users SET name='张三（已更新）' WHERE id=1 → RowsAffected=1\n")

	// 模拟 Delete
	fmt.Printf("  DELETE FROM users WHERE id=4 → RowsAffected=1\n")
	fmt.Printf("  当前用户数: %d\n", len(users)-1)
}

// demoPreparedStatement 演示预编译语句概念
func demoPreparedStatement() {
	fmt.Println("\n--- 3. 预编译语句概念 ---")

	fmt.Println("预编译语句的优势：")
	fmt.Println("  1. 防止 SQL 注入：参数与 SQL 语句分离")
	fmt.Println("  2. 性能提升：SQL 只解析一次，多次执行复用执行计划")
	fmt.Println("  3. 类型安全：数据库驱动自动处理参数类型转换")

	fmt.Println("\n示例代码：")
	fmt.Println(`  stmt, err := db.Prepare("SELECT * FROM users WHERE id = ?")`)
	fmt.Println(`  defer stmt.Close()`)
	fmt.Println(`  stmt.QueryRow(1)  // 复用预编译语句`)
	fmt.Println(`  stmt.QueryRow(2)  // 复用预编译语句`)

	// 演示 SQL 注入防护
	fmt.Println("\nSQL 注入防护：")
	maliciousInput := "1 OR 1=1; DROP TABLE users; --"
	fmt.Printf("  恶意输入: %s\n", maliciousInput)
	fmt.Println("  ❌ 拼接 SQL: SELECT * FROM users WHERE id = " + maliciousInput)
	fmt.Println("  ✅ 预编译: SELECT * FROM users WHERE id = ? → 参数作为值处理，不会被解析为 SQL")
}

// demoTransaction 演示事务概念
func demoTransaction() {
	fmt.Println("\n--- 4. 事务概念 ---")

	// 模拟转账事务
	type Account struct {
		ID      int
		Name    string
		Balance float64
	}

	accounts := []Account{
		{ID: 1, Name: "账户A", Balance: 1000},
		{ID: 2, Name: "账户B", Balance: 500},
	}

	fmt.Println("转账前：")
	for _, a := range accounts {
		fmt.Printf("  %s: %.2f 元\n", a.Name, a.Balance)
	}

	// 模拟事务：从账户A转100元到账户B
	amount := 100.0
	fmt.Printf("\n执行事务：从 %s 转 %.0f 元到 %s\n", accounts[0].Name, amount, accounts[1].Name)
	fmt.Println("  tx.Begin()")
	fmt.Println("  tx.Exec(UPDATE accounts SET balance = balance - 100 WHERE id = 1)")
	fmt.Println("  tx.Exec(UPDATE accounts SET balance = balance + 100 WHERE id = 2)")
	fmt.Println("  tx.Commit()")

	accounts[0].Balance -= amount
	accounts[1].Balance += amount

	fmt.Println("\n转账后：")
	for _, a := range accounts {
		fmt.Printf("  %s: %.2f 元\n", a.Name, a.Balance)
	}

	fmt.Println("\n事务隔离级别设置：")
	fmt.Println(`  tx, err := db.BeginTx(ctx, &sql.TxOptions{`)
	fmt.Println(`      Isolation: sql.LevelRepeatableRead,`)
	fmt.Println(`      ReadOnly:  false,`)
	fmt.Println(`  })`)
}

// demoNullTypes 演示 Null 类型处理
func demoNullTypes() {
	fmt.Println("\n--- 5. Null 类型处理 ---")

	// 模拟可空字段
	nullEmail := sql.NullString{String: "", Valid: false}
	validEmail := sql.NullString{String: "test@example.com", Valid: true}
	nullAge := sql.NullInt64{Int64: 0, Valid: false}
	validAge := sql.NullInt64{Int64: 25, Valid: true}

	fmt.Println("sql.NullString 示例：")
	fmt.Printf("  NULL email: Valid=%v, String=%q\n", nullEmail.Valid, nullEmail.String)
	fmt.Printf("  有效 email: Valid=%v, String=%q\n", validEmail.Valid, validEmail.String)

	fmt.Println("\nsql.NullInt64 示例：")
	fmt.Printf("  NULL age: Valid=%v, Int64=%d\n", nullAge.Valid, nullAge.Int64)
	fmt.Printf("  有效 age: Valid=%v, Int64=%d\n", validAge.Valid, validAge.Int64)

	fmt.Println("\n常见陷阱：")
	fmt.Println("  ❌ var name string; row.Scan(&name)  // 遇到 NULL 会报错")
	fmt.Println("  ✅ var name sql.NullString; row.Scan(&name)  // 正确处理 NULL")
}

// demoScannerValuer 演示 Scanner 和 Valuer 接口
func demoScannerValuer() {
	fmt.Println("\n--- 6. Scanner 和 Valuer 接口 ---")

	fmt.Println("sql.Scanner 接口（从数据库读取时自定义解析）：")
	fmt.Println(`  type Scanner interface {`)
	fmt.Println(`      Scan(src interface{}) error`)
	fmt.Println(`  }`)

	fmt.Println("\ndriver.Valuer 接口（写入数据库时自定义序列化）：")
	fmt.Println(`  type Valuer interface {`)
	fmt.Println(`      Value() (driver.Value, error)`)
	fmt.Println(`  }`)

	// 模拟自定义类型
	fmt.Println("\n应用场景示例 — JSON 字段映射：")
	fmt.Println(`  type JSONMap map[string]interface{}`)
	fmt.Println(`  func (j *JSONMap) Scan(src interface{}) error { ... }  // 从 DB 读取 JSON`)
	fmt.Println(`  func (j JSONMap) Value() (driver.Value, error) { ... } // 写入 DB 为 JSON`)

	fmt.Println("\n应用场景示例 — 枚举类型：")
	fmt.Println(`  type Status int`)
	fmt.Println(`  const (Active Status = 1; Inactive Status = 2)`)
	fmt.Println(`  func (s *Status) Scan(src interface{}) error { ... }`)
	fmt.Println(`  func (s Status) Value() (driver.Value, error) { return int64(s), nil }`)
}

// ============================================================
// Part B：连接真实 MySQL 和 PostgreSQL
// ============================================================

func partB() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Part B：连接真实数据库（MySQL + PostgreSQL）")
	fmt.Println(strings.Repeat("=", 60))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// MySQL 连接
	fmt.Println("\n--- MySQL 示例 ---")
	mysqlDSN := "root:root123@tcp(localhost:3306)/testdb?parseTime=true&charset=utf8mb4"
	mysqlDB, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		fmt.Printf("MySQL 连接失败: %v\n", err)
		fmt.Println("请确保已启动 MySQL: docker compose -f docker/docker-compose.yml up -d mysql")
	} else {
		defer mysqlDB.Close()
		// 配置连接池
		mysqlDB.SetMaxOpenConns(25)
		mysqlDB.SetMaxIdleConns(10)
		mysqlDB.SetConnMaxLifetime(5 * time.Minute)

		if err := mysqlDB.PingContext(ctx); err != nil {
			fmt.Printf("MySQL Ping 失败: %v\n", err)
			fmt.Println("请确保已启动 MySQL: docker compose -f docker/docker-compose.yml up -d mysql")
		} else {
			fmt.Println("✅ MySQL 连接成功")
			runMySQLDemo(ctx, mysqlDB)
		}
	}

	// PostgreSQL 连接
	fmt.Println("\n--- PostgreSQL 示例 ---")
	pgDSN := "postgres://postgres:postgres123@localhost:5432/testdb?sslmode=disable"
	pgDB, err := sql.Open("pgx", pgDSN)
	if err != nil {
		fmt.Printf("PostgreSQL 连接失败: %v\n", err)
		fmt.Println("请确保已启动 PostgreSQL: docker compose -f docker/docker-compose.yml up -d postgresql")
	} else {
		defer pgDB.Close()
		pgDB.SetMaxOpenConns(25)
		pgDB.SetMaxIdleConns(10)

		if err := pgDB.PingContext(ctx); err != nil {
			fmt.Printf("PostgreSQL Ping 失败: %v\n", err)
			fmt.Println("请确保已启动 PostgreSQL: docker compose -f docker/docker-compose.yml up -d postgresql")
		} else {
			fmt.Println("✅ PostgreSQL 连接成功")
			runPostgreSQLDemo(ctx, pgDB)
		}
	}
}

// runMySQLDemo 执行 MySQL CRUD 示例
func runMySQLDemo(ctx context.Context, db *sql.DB) {
	// 创建表
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS demo_users (
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
	fmt.Println("  创建表 demo_users 成功")

	// 插入数据
	result, err := db.ExecContext(ctx,
		"INSERT INTO demo_users (name, email, age) VALUES (?, ?, ?)",
		"张三", "zhangsan@example.com", 25)
	if err != nil {
		fmt.Printf("插入失败: %v\n", err)
		return
	}
	lastID, _ := result.LastInsertId()
	fmt.Printf("  插入成功, LastInsertId=%d\n", lastID)

	// 查询
	var name string
	var age int
	err = db.QueryRowContext(ctx, "SELECT name, age FROM demo_users WHERE id = ?", lastID).Scan(&name, &age)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	fmt.Printf("  查询结果: name=%s, age=%d\n", name, age)

	// 清理
	db.ExecContext(ctx, "DROP TABLE IF EXISTS demo_users")
	fmt.Println("  清理完成")
}

// runPostgreSQLDemo 执行 PostgreSQL CRUD 示例
func runPostgreSQLDemo(ctx context.Context, db *sql.DB) {
	// 创建表（PostgreSQL 使用 SERIAL 和 $1 占位符）
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS demo_users (
			id SERIAL PRIMARY KEY,
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
	fmt.Println("  创建表 demo_users 成功")

	// 插入数据（PostgreSQL 使用 RETURNING 获取自增 ID）
	var lastID int
	err = db.QueryRowContext(ctx,
		"INSERT INTO demo_users (name, email, age) VALUES ($1, $2, $3) RETURNING id",
		"李四", "lisi@example.com", 30).Scan(&lastID)
	if err != nil {
		fmt.Printf("插入失败: %v\n", err)
		return
	}
	fmt.Printf("  插入成功, id=%d\n", lastID)

	// 查询
	var name string
	var age int
	err = db.QueryRowContext(ctx, "SELECT name, age FROM demo_users WHERE id = $1", lastID).Scan(&name, &age)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	fmt.Printf("  查询结果: name=%s, age=%d\n", name, age)

	// 清理
	db.ExecContext(ctx, "DROP TABLE IF EXISTS demo_users")
	fmt.Println("  清理完成")
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
