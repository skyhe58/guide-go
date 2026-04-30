// 数据库迁移概念演示
// 演示：golang-migrate 和 goose 的迁移概念、迁移文件结构、版本管理
// Go 1.22+ | 验证日期 2025-01-01
//
// Part A：纯内存模拟，直接运行理解数据库迁移核心概念
// 本示例不需要连接真实数据库
//
// 运行方式：
//   go run main.go

package main

import (
	"fmt"
	"strings"
	"time"
)

// ============================================================
// 数据模型
// ============================================================

// Migration 迁移记录
type Migration struct {
	Version   int
	Name      string
	UpSQL     string
	DownSQL   string
	AppliedAt time.Time
	Applied   bool
}

// MigrationManager 迁移管理器（内存模拟）
type MigrationManager struct {
	migrations     []Migration
	currentVersion int
	history        []string // 操作历史
}

// ============================================================
// 迁移管理器实现
// ============================================================

// NewMigrationManager 创建迁移管理器
func NewMigrationManager() *MigrationManager {
	return &MigrationManager{
		migrations: []Migration{
			{
				Version: 1,
				Name:    "create_users",
				UpSQL: `CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`,
				DownSQL: "DROP TABLE IF EXISTS users;",
			},
			{
				Version: 2,
				Name:    "add_users_age",
				UpSQL:   "ALTER TABLE users ADD COLUMN age INT DEFAULT 0;",
				DownSQL: "ALTER TABLE users DROP COLUMN age;",
			},
			{
				Version: 3,
				Name:    "create_articles",
				UpSQL: `CREATE TABLE articles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    content TEXT,
    user_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);`,
				DownSQL: "DROP TABLE IF EXISTS articles;",
			},
			{
				Version: 4,
				Name:    "add_email_index",
				UpSQL:   "CREATE INDEX idx_users_email ON users(email);",
				DownSQL: "DROP INDEX idx_users_email ON users;",
			},
			{
				Version: 5,
				Name:    "create_tags",
				UpSQL: `CREATE TABLE tags (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL
);
CREATE TABLE article_tags (
    article_id INT NOT NULL,
    tag_id INT NOT NULL,
    PRIMARY KEY (article_id, tag_id),
    FOREIGN KEY (article_id) REFERENCES articles(id),
    FOREIGN KEY (tag_id) REFERENCES tags(id)
);`,
				DownSQL: "DROP TABLE IF EXISTS article_tags;\nDROP TABLE IF EXISTS tags;",
			},
		},
		currentVersion: 0,
	}
}

// Up 执行向上迁移
func (m *MigrationManager) Up(steps int) {
	applied := 0
	for i := range m.migrations {
		if m.migrations[i].Version > m.currentVersion {
			m.migrations[i].Applied = true
			m.migrations[i].AppliedAt = time.Now()
			m.currentVersion = m.migrations[i].Version
			m.history = append(m.history, fmt.Sprintf("UP: %03d_%s", m.migrations[i].Version, m.migrations[i].Name))
			fmt.Printf("  ✅ 执行迁移 %03d_%s.up.sql\n", m.migrations[i].Version, m.migrations[i].Name)
			applied++
			if steps > 0 && applied >= steps {
				break
			}
		}
	}
	if applied == 0 {
		fmt.Println("  没有待执行的迁移")
	}
}

// Down 执行向下回滚
func (m *MigrationManager) Down(steps int) {
	applied := 0
	for i := len(m.migrations) - 1; i >= 0; i-- {
		if m.migrations[i].Applied && m.migrations[i].Version == m.currentVersion {
			m.migrations[i].Applied = false
			m.history = append(m.history, fmt.Sprintf("DOWN: %03d_%s", m.migrations[i].Version, m.migrations[i].Name))
			fmt.Printf("  ⬇️ 回滚迁移 %03d_%s.down.sql\n", m.migrations[i].Version, m.migrations[i].Name)
			m.currentVersion--
			applied++
			if steps > 0 && applied >= steps {
				break
			}
		}
	}
	if applied == 0 {
		fmt.Println("  没有可回滚的迁移")
	}
}

// Status 显示迁移状态
func (m *MigrationManager) Status() {
	fmt.Printf("  当前版本: %d\n", m.currentVersion)
	fmt.Println("  迁移文件列表：")
	for _, mg := range m.migrations {
		status := "⬜ 未执行"
		if mg.Applied {
			status = "✅ 已执行"
		}
		fmt.Printf("    %s %03d_%s\n", status, mg.Version, mg.Name)
	}
}

// ShowMigrationFile 显示迁移文件内容
func (m *MigrationManager) ShowMigrationFile(version int) {
	for _, mg := range m.migrations {
		if mg.Version == version {
			fmt.Printf("\n  文件: %03d_%s.up.sql\n", mg.Version, mg.Name)
			fmt.Println("  " + strings.Repeat("-", 40))
			for _, line := range strings.Split(mg.UpSQL, "\n") {
				fmt.Printf("  %s\n", line)
			}
			fmt.Printf("\n  文件: %03d_%s.down.sql\n", mg.Version, mg.Name)
			fmt.Println("  " + strings.Repeat("-", 40))
			for _, line := range strings.Split(mg.DownSQL, "\n") {
				fmt.Printf("  %s\n", line)
			}
			return
		}
	}
}

// ============================================================
// 演示函数
// ============================================================

func main() {
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println("数据库迁移概念演示（内存模拟）")
	fmt.Println("=" + strings.Repeat("=", 59))

	mgr := NewMigrationManager()

	// 1. 迁移文件结构
	demoMigrationFiles(mgr)

	// 2. 执行迁移
	demoMigrateUp(mgr)

	// 3. 回滚迁移
	demoMigrateDown(mgr)

	// 4. 部分迁移
	demoPartialMigration(mgr)

	// 5. golang-migrate CLI 命令
	demoGolangMigrateCLI()

	// 6. goose CLI 命令
	demoGooseCLI()

	// 7. GORM AutoMigrate 对比
	demoAutoMigrateComparison()

	// 8. 最佳实践
	demoBestPractices()
}

// demoMigrationFiles 演示迁移文件结构
func demoMigrationFiles(mgr *MigrationManager) {
	fmt.Println("\n--- 1. 迁移文件结构 ---")

	fmt.Println("目录结构：")
	fmt.Println("  migrations/")
	fmt.Println("  ├── 001_create_users.up.sql")
	fmt.Println("  ├── 001_create_users.down.sql")
	fmt.Println("  ├── 002_add_users_age.up.sql")
	fmt.Println("  ├── 002_add_users_age.down.sql")
	fmt.Println("  ├── 003_create_articles.up.sql")
	fmt.Println("  ├── 003_create_articles.down.sql")
	fmt.Println("  ├── 004_add_email_index.up.sql")
	fmt.Println("  ├── 004_add_email_index.down.sql")
	fmt.Println("  ├── 005_create_tags.up.sql")
	fmt.Println("  └── 005_create_tags.down.sql")

	fmt.Println("\n迁移文件内容示例：")
	mgr.ShowMigrationFile(1)
}

// demoMigrateUp 演示执行迁移
func demoMigrateUp(mgr *MigrationManager) {
	fmt.Println("\n--- 2. 执行所有迁移 (migrate up) ---")
	mgr.Status()
	fmt.Println("\n  执行 migrate up:")
	mgr.Up(0) // 执行所有
	fmt.Println()
	mgr.Status()
}

// demoMigrateDown 演示回滚迁移
func demoMigrateDown(mgr *MigrationManager) {
	fmt.Println("\n--- 3. 回滚迁移 (migrate down 2) ---")
	fmt.Println("  回滚最近 2 个迁移:")
	mgr.Down(2)
	fmt.Println()
	mgr.Status()
}

// demoPartialMigration 演示部分迁移
func demoPartialMigration(mgr *MigrationManager) {
	fmt.Println("\n--- 4. 部分迁移 (migrate up 1) ---")
	fmt.Println("  执行下一个迁移:")
	mgr.Up(1)
	fmt.Println()
	mgr.Status()

	// 恢复到最新
	fmt.Println("\n  执行剩余迁移:")
	mgr.Up(0)
	mgr.Status()
}

// demoGolangMigrateCLI 演示 golang-migrate CLI
func demoGolangMigrateCLI() {
	fmt.Println("\n--- 5. golang-migrate CLI 命令 ---")

	fmt.Println("安装：")
	fmt.Println("  go install -tags 'mysql postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest")

	fmt.Println("\n常用命令：")
	fmt.Println("  # 创建迁移文件")
	fmt.Println("  migrate create -ext sql -dir migrations -seq create_users")
	fmt.Println()
	fmt.Println("  # 执行所有迁移")
	fmt.Println("  migrate -path migrations -database \"mysql://root:root123@tcp(localhost:3306)/testdb\" up")
	fmt.Println()
	fmt.Println("  # 回滚一步")
	fmt.Println("  migrate -path migrations -database \"mysql://root:root123@tcp(localhost:3306)/testdb\" down 1")
	fmt.Println()
	fmt.Println("  # 查看当前版本")
	fmt.Println("  migrate -path migrations -database \"mysql://root:root123@tcp(localhost:3306)/testdb\" version")
	fmt.Println()
	fmt.Println("  # 强制设置版本（修复脏状态）")
	fmt.Println("  migrate -path migrations -database \"...\" force 3")

	fmt.Println("\nGo 库方式：")
	fmt.Println("  m, _ := migrate.New(\"file://migrations\", \"mysql://root:root123@tcp(localhost:3306)/testdb\")")
	fmt.Println("  m.Up()       // 执行所有迁移")
	fmt.Println("  m.Down()     // 回滚所有迁移")
	fmt.Println("  m.Steps(-1)  // 回滚一步")
}

// demoGooseCLI 演示 goose CLI
func demoGooseCLI() {
	fmt.Println("\n--- 6. goose CLI 命令 ---")

	fmt.Println("安装：")
	fmt.Println("  go install github.com/pressly/goose/v3/cmd/goose@latest")

	fmt.Println("\n常用命令：")
	fmt.Println("  # 创建 SQL 迁移")
	fmt.Println("  goose -dir migrations create create_users sql")
	fmt.Println()
	fmt.Println("  # 创建 Go 迁移（支持复杂逻辑）")
	fmt.Println("  goose -dir migrations create seed_data go")
	fmt.Println()
	fmt.Println("  # 执行迁移")
	fmt.Println("  goose -dir migrations mysql \"root:root123@tcp(localhost:3306)/testdb\" up")
	fmt.Println()
	fmt.Println("  # 回滚")
	fmt.Println("  goose -dir migrations mysql \"root:root123@tcp(localhost:3306)/testdb\" down")
	fmt.Println()
	fmt.Println("  # 查看状态")
	fmt.Println("  goose -dir migrations mysql \"root:root123@tcp(localhost:3306)/testdb\" status")

	fmt.Println("\ngoose 特色：支持 Go 代码迁移（适合数据迁移等复杂场景）")
}

// demoAutoMigrateComparison 演示 GORM AutoMigrate 对比
func demoAutoMigrateComparison() {
	fmt.Println("\n--- 7. GORM AutoMigrate vs 专业迁移工具 ---")

	fmt.Println("GORM AutoMigrate：")
	fmt.Println("  db.AutoMigrate(&User{}, &Article{}, &Tag{})")
	fmt.Println()
	fmt.Println("  ✅ 优点：简单方便，适合开发环境")
	fmt.Println("  ❌ 局限：")
	fmt.Println("     - 只能添加字段和索引")
	fmt.Println("     - 不能删除字段")
	fmt.Println("     - 不能修改字段类型")
	fmt.Println("     - 不支持回滚")
	fmt.Println("     - 不适合生产环境")

	fmt.Println("\n专业迁移工具（golang-migrate / goose）：")
	fmt.Println("  ✅ 版本化管理，可追溯")
	fmt.Println("  ✅ 支持回滚（down 文件）")
	fmt.Println("  ✅ 支持任意 DDL 操作")
	fmt.Println("  ✅ 适合生产环境")
	fmt.Println("  ✅ 可集成到 CI/CD 流水线")

	fmt.Println("\n推荐方案：")
	fmt.Println("  开发环境：GORM AutoMigrate（快速迭代）")
	fmt.Println("  生产环境：golang-migrate 或 goose（版本化管理）")
}

// demoBestPractices 演示最佳实践
func demoBestPractices() {
	fmt.Println("\n--- 8. 数据库迁移最佳实践 ---")

	fmt.Println("1. 每个迁移文件必须有对应的 down 文件（可回滚）")
	fmt.Println("2. 迁移文件一旦提交不可修改，只能创建新的迁移")
	fmt.Println("3. 大表变更使用 pt-online-schema-change 或 gh-ost 避免锁表")
	fmt.Println("4. 迁移文件纳入版本控制，与代码一起 review")
	fmt.Println("5. CI/CD 中自动执行迁移，避免手动操作")
	fmt.Println("6. 迁移文件命名规范：序号_描述.up.sql / 序号_描述.down.sql")
	fmt.Println("7. 每个迁移文件只做一件事，保持原子性")
	fmt.Println("8. 测试环境先执行迁移，验证通过后再上生产")
}
