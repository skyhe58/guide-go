// RBAC 权限控制 — Casbin 权限框架
// 演示：RBAC 模型定义、策略管理、权限检查、角色继承、RESTful 路径匹配
// Go 1.22+ | 验证日期 2025-01-01
//
// 运行方式：
//   go run ./rbac-casbin/
//
// 依赖：github.com/casbin/casbin/v2

package main

import (
	"fmt"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

// ============================================================
// Casbin RBAC 模型定义
// ============================================================

// RBAC 模型配置（通常写在 model.conf 文件中）
// 这里使用字符串加载，方便演示
const rbacModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

// RESTful RBAC 模型（支持路径匹配）
const restfulRBACModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && r.act == p.act
`

// ============================================================
// 演示 1：基础 RBAC 权限控制
// ============================================================

func demoBasicRBAC() {
	fmt.Println("\n--- 1. 基础 RBAC 权限控制 ---")

	// 从字符串加载模型
	m, err := model.NewModelFromString(rbacModel)
	if err != nil {
		fmt.Printf("加载模型失败: %v\n", err)
		return
	}

	// 创建 Enforcer（使用内存适配器）
	e, err := casbin.NewEnforcer(m)
	if err != nil {
		fmt.Printf("创建 Enforcer 失败: %v\n", err)
		return
	}

	// --- 添加策略规则 ---
	// 格式：角色, 资源, 操作

	// admin 角色权限
	_, _ = e.AddPolicy("admin", "article", "read")
	_, _ = e.AddPolicy("admin", "article", "create")
	_, _ = e.AddPolicy("admin", "article", "edit")
	_, _ = e.AddPolicy("admin", "article", "delete")
	_, _ = e.AddPolicy("admin", "user", "manage")
	_, _ = e.AddPolicy("admin", "system", "config")

	// author 角色权限
	_, _ = e.AddPolicy("author", "article", "read")
	_, _ = e.AddPolicy("author", "article", "create")
	_, _ = e.AddPolicy("author", "article", "edit")
	_, _ = e.AddPolicy("author", "comment", "create")

	// reader 角色权限
	_, _ = e.AddPolicy("reader", "article", "read")
	_, _ = e.AddPolicy("reader", "comment", "create")

	// --- 分配用户角色 ---
	_, _ = e.AddGroupingPolicy("zhangsan", "admin")
	_, _ = e.AddGroupingPolicy("lisi", "author")
	_, _ = e.AddGroupingPolicy("wangwu", "reader")

	// --- 权限检查 ---
	fmt.Println("\n  权限检查结果：")

	checks := []struct {
		user     string
		resource string
		action   string
	}{
		{"zhangsan", "article", "delete"},  // admin → ✅
		{"zhangsan", "user", "manage"},     // admin → ✅
		{"lisi", "article", "create"},      // author → ✅
		{"lisi", "article", "delete"},      // author → ❌
		{"lisi", "user", "manage"},         // author → ❌
		{"wangwu", "article", "read"},      // reader → ✅
		{"wangwu", "article", "create"},    // reader → ❌
		{"wangwu", "comment", "create"},    // reader → ✅
	}

	for _, c := range checks {
		allowed, _ := e.Enforce(c.user, c.resource, c.action)
		status := "❌ 拒绝"
		if allowed {
			status = "✅ 允许"
		}
		fmt.Printf("  %s → %s:%s → %s\n", c.user, c.resource, c.action, status)
	}

	// --- 查询角色信息 ---
	fmt.Println("\n  角色信息：")
	for _, user := range []string{"zhangsan", "lisi", "wangwu"} {
		roles, _ := e.GetRolesForUser(user)
		fmt.Printf("  %s 的角色: %v\n", user, roles)
	}
}

// ============================================================
// 演示 2：角色继承（层级 RBAC）
// ============================================================

func demoRoleHierarchy() {
	fmt.Println("\n--- 2. 角色继承（层级 RBAC） ---")

	m, _ := model.NewModelFromString(rbacModel)
	e, _ := casbin.NewEnforcer(m)

	// 只给最低级别角色定义权限
	_, _ = e.AddPolicy("reader", "article", "read")
	_, _ = e.AddPolicy("reader", "comment", "create")
	_, _ = e.AddPolicy("author", "article", "create")
	_, _ = e.AddPolicy("author", "article", "edit")
	_, _ = e.AddPolicy("admin", "article", "delete")
	_, _ = e.AddPolicy("admin", "user", "manage")

	// 角色继承：admin → author → reader
	// admin 继承 author 的所有权限，author 继承 reader 的所有权限
	_, _ = e.AddGroupingPolicy("author", "reader") // author 继承 reader
	_, _ = e.AddGroupingPolicy("admin", "author")  // admin 继承 author

	// 分配用户角色
	_, _ = e.AddGroupingPolicy("zhangsan", "admin")
	_, _ = e.AddGroupingPolicy("lisi", "author")
	_, _ = e.AddGroupingPolicy("wangwu", "reader")

	fmt.Println("\n  角色继承链：admin → author → reader")
	fmt.Println("  权限检查结果：")

	checks := []struct {
		user     string
		resource string
		action   string
		expect   string
	}{
		{"zhangsan", "article", "read", "admin 继承 reader 的 read 权限"},
		{"zhangsan", "article", "create", "admin 继承 author 的 create 权限"},
		{"zhangsan", "article", "delete", "admin 自身的 delete 权限"},
		{"zhangsan", "user", "manage", "admin 自身的 manage 权限"},
		{"lisi", "article", "read", "author 继承 reader 的 read 权限"},
		{"lisi", "article", "create", "author 自身的 create 权限"},
		{"lisi", "article", "delete", "author 没有 delete 权限"},
		{"wangwu", "article", "read", "reader 自身的 read 权限"},
		{"wangwu", "article", "create", "reader 没有 create 权限"},
	}

	for _, c := range checks {
		allowed, _ := e.Enforce(c.user, c.resource, c.action)
		status := "❌"
		if allowed {
			status = "✅"
		}
		fmt.Printf("  %s %s:%s → %s（%s）\n", status, c.user, c.action, c.resource, c.expect)
	}
}

// ============================================================
// 演示 3：RESTful API 权限控制
// ============================================================

func demoRESTfulRBAC() {
	fmt.Println("\n--- 3. RESTful API 权限控制（路径匹配） ---")

	m, _ := model.NewModelFromString(restfulRBACModel)
	e, _ := casbin.NewEnforcer(m)

	// RESTful 策略：角色, URL 路径, HTTP 方法
	_, _ = e.AddPolicy("admin", "/api/users", "GET")
	_, _ = e.AddPolicy("admin", "/api/users/:id", "DELETE")
	_, _ = e.AddPolicy("admin", "/api/articles", "GET")
	_, _ = e.AddPolicy("admin", "/api/articles", "POST")
	_, _ = e.AddPolicy("admin", "/api/articles/:id", "PUT")
	_, _ = e.AddPolicy("admin", "/api/articles/:id", "DELETE")

	_, _ = e.AddPolicy("author", "/api/articles", "GET")
	_, _ = e.AddPolicy("author", "/api/articles", "POST")
	_, _ = e.AddPolicy("author", "/api/articles/:id", "PUT")

	_, _ = e.AddPolicy("reader", "/api/articles", "GET")
	_, _ = e.AddPolicy("reader", "/api/articles/:id", "GET")

	// 分配用户角色
	_, _ = e.AddGroupingPolicy("zhangsan", "admin")
	_, _ = e.AddGroupingPolicy("lisi", "author")
	_, _ = e.AddGroupingPolicy("wangwu", "reader")

	fmt.Println("\n  RESTful API 权限检查：")

	checks := []struct {
		user   string
		path   string
		method string
	}{
		{"zhangsan", "/api/users", "GET"},          // admin → ✅
		{"zhangsan", "/api/users/123", "DELETE"},    // admin → ✅（keyMatch2 匹配 :id）
		{"lisi", "/api/articles", "POST"},           // author → ✅
		{"lisi", "/api/articles/456", "PUT"},        // author → ✅
		{"lisi", "/api/articles/456", "DELETE"},     // author → ❌
		{"lisi", "/api/users", "GET"},               // author → ❌
		{"wangwu", "/api/articles", "GET"},          // reader → ✅
		{"wangwu", "/api/articles/789", "GET"},      // reader → ✅
		{"wangwu", "/api/articles", "POST"},         // reader → ❌
	}

	for _, c := range checks {
		allowed, _ := e.Enforce(c.user, c.path, c.method)
		status := "❌ 拒绝"
		if allowed {
			status = "✅ 允许"
		}
		fmt.Printf("  %s %s %s → %s\n", c.user, c.method, c.path, status)
	}
}

// ============================================================
// 演示 4：动态策略管理
// ============================================================

func demoPolicyManagement() {
	fmt.Println("\n--- 4. 动态策略管理 ---")

	m, _ := model.NewModelFromString(rbacModel)
	e, _ := casbin.NewEnforcer(m)

	// 初始策略
	_, _ = e.AddPolicy("author", "article", "read")
	_, _ = e.AddGroupingPolicy("lisi", "author")

	// 检查初始权限
	allowed, _ := e.Enforce("lisi", "article", "read")
	fmt.Printf("  初始状态 — lisi 读取文章: %v\n", allowed)

	allowed, _ = e.Enforce("lisi", "article", "delete")
	fmt.Printf("  初始状态 — lisi 删除文章: %v\n", allowed)

	// 动态添加权限
	fmt.Println("\n  [操作] 给 author 角色添加 delete 权限...")
	_, _ = e.AddPolicy("author", "article", "delete")

	allowed, _ = e.Enforce("lisi", "article", "delete")
	fmt.Printf("  添加后 — lisi 删除文章: %v\n", allowed)

	// 动态移除权限
	fmt.Println("\n  [操作] 移除 author 角色的 delete 权限...")
	_, _ = e.RemovePolicy("author", "article", "delete")

	allowed, _ = e.Enforce("lisi", "article", "delete")
	fmt.Printf("  移除后 — lisi 删除文章: %v\n", allowed)

	// 动态修改用户角色
	fmt.Println("\n  [操作] 将 lisi 从 author 升级为 admin...")
	_, _ = e.AddPolicy("admin", "article", "delete")
	_, _ = e.AddPolicy("admin", "article", "read")
	_, _ = e.RemoveGroupingPolicy("lisi", "author")
	_, _ = e.AddGroupingPolicy("lisi", "admin")

	roles, _ := e.GetRolesForUser("lisi")
	fmt.Printf("  lisi 当前角色: %v\n", roles)

	allowed, _ = e.Enforce("lisi", "article", "delete")
	fmt.Printf("  升级后 — lisi 删除文章: %v\n", allowed)

	// 查询所有策略
	fmt.Println("\n  当前所有策略规则：")
	policies, _ := e.GetPolicy()
	for _, p := range policies {
		fmt.Printf("    %v\n", p)
	}

	fmt.Println("\n  当前角色分配：")
	grouping, _ := e.GetGroupingPolicy()
	for _, g := range grouping {
		fmt.Printf("    %s → %s\n", g[0], g[1])
	}
}

// ============================================================
// 演示入口
// ============================================================

func main() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("RBAC 权限控制 — Casbin 权限框架")
	fmt.Println(strings.Repeat("=", 60))

	// 1. 基础 RBAC
	demoBasicRBAC()

	// 2. 角色继承
	demoRoleHierarchy()

	// 3. RESTful API 权限
	demoRESTfulRBAC()

	// 4. 动态策略管理
	demoPolicyManagement()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Casbin RBAC 权限控制演示完成")
	fmt.Println(strings.Repeat("=", 60))
}
