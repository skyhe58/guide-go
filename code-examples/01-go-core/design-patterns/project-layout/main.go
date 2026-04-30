// 标准 Go 项目布局演示 — cmd/internal/pkg 目录规范
// 本示例展示 Go 项目的标准目录结构和各目录的职责
//
// Go 版本要求: Go 1.22+
// 最后验证日期: 2025-01-01
//
// 运行方式: go run ./project-layout/

package main

import "fmt"

// =============================================================================
// Part A: 标准 Go 项目布局说明
// =============================================================================

// ProjectLayout 描述一个标准 Go 项目的目录结构
type ProjectLayout struct {
	Name        string
	Description string
	Children    []ProjectLayout
}

// Print 递归打印目录结构
func (p ProjectLayout) Print(indent string, isLast bool) {
	// 选择连接符
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	// 打印当前节点
	if indent == "" {
		fmt.Printf("%s/\n", p.Name)
	} else {
		desc := ""
		if p.Description != "" {
			desc = fmt.Sprintf("  # %s", p.Description)
		}
		fmt.Printf("%s%s%s%s\n", indent, connector, p.Name, desc)
	}

	// 计算子节点的缩进
	childIndent := indent
	if indent != "" {
		if isLast {
			childIndent += "    "
		} else {
			childIndent += "│   "
		}
	}

	// 递归打印子节点
	for i, child := range p.Children {
		child.Print(childIndent, i == len(p.Children)-1)
	}
}

func main() {
	fmt.Println("=== 标准 Go 项目布局演示 ===")
	fmt.Println()

	// --- 1. 大型项目布局 ---
	fmt.Println("--- 1. 大型项目布局（如微服务、多入口应用）---")
	fmt.Println()

	largeProject := ProjectLayout{
		Name: "myproject",
		Children: []ProjectLayout{
			{Name: "cmd/", Description: "应用入口（每个子目录一个可执行文件）", Children: []ProjectLayout{
				{Name: "server/", Description: "API 服务入口", Children: []ProjectLayout{
					{Name: "main.go", Description: "保持简洁，只做初始化和启动"},
				}},
				{Name: "worker/", Description: "后台任务入口", Children: []ProjectLayout{
					{Name: "main.go"},
				}},
				{Name: "cli/", Description: "命令行工具入口", Children: []ProjectLayout{
					{Name: "main.go"},
				}},
			}},
			{Name: "internal/", Description: "私有代码（Go 编译器强制限制外部导入）", Children: []ProjectLayout{
				{Name: "config/", Description: "配置加载", Children: []ProjectLayout{
					{Name: "config.go"},
				}},
				{Name: "handler/", Description: "HTTP 处理器（Controller 层）", Children: []ProjectLayout{
					{Name: "user.go"},
					{Name: "article.go"},
				}},
				{Name: "service/", Description: "业务逻辑层", Children: []ProjectLayout{
					{Name: "user.go"},
					{Name: "article.go"},
				}},
				{Name: "repository/", Description: "数据访问层", Children: []ProjectLayout{
					{Name: "user.go"},
					{Name: "article.go"},
				}},
				{Name: "model/", Description: "数据模型", Children: []ProjectLayout{
					{Name: "user.go"},
					{Name: "article.go"},
				}},
				{Name: "middleware/", Description: "HTTP 中间件", Children: []ProjectLayout{
					{Name: "auth.go"},
					{Name: "logging.go"},
				}},
			}},
			{Name: "pkg/", Description: "可被外部项目导入的公共库", Children: []ProjectLayout{
				{Name: "logger/", Description: "日志工具", Children: []ProjectLayout{
					{Name: "logger.go"},
				}},
				{Name: "validator/", Description: "校验工具", Children: []ProjectLayout{
					{Name: "validator.go"},
				}},
			}},
			{Name: "api/", Description: "API 定义文件", Children: []ProjectLayout{
				{Name: "proto/", Description: "Protocol Buffers 定义"},
				{Name: "openapi/", Description: "OpenAPI/Swagger 定义"},
			}},
			{Name: "configs/", Description: "配置文件模板", Children: []ProjectLayout{
				{Name: "config.yaml"},
				{Name: "config.example.yaml"},
			}},
			{Name: "deployments/", Description: "部署配置", Children: []ProjectLayout{
				{Name: "Dockerfile", Description: "多阶段构建"},
				{Name: "docker-compose.yml"},
				{Name: "k8s/", Description: "Kubernetes 部署清单"},
			}},
			{Name: "migrations/", Description: "数据库迁移文件"},
			{Name: "scripts/", Description: "构建和工具脚本"},
			{Name: "test/", Description: "集成测试"},
			{Name: "go.mod"},
			{Name: "go.sum"},
			{Name: "Makefile", Description: "构建自动化"},
			{Name: "README.md"},
		},
	}
	largeProject.Print("", false)
	fmt.Println()

	// --- 2. 中型项目布局 ---
	fmt.Println("--- 2. 中型项目布局（单个微服务）---")
	fmt.Println()

	mediumProject := ProjectLayout{
		Name: "myservice",
		Children: []ProjectLayout{
			{Name: "cmd/", Children: []ProjectLayout{
				{Name: "server/", Children: []ProjectLayout{
					{Name: "main.go"},
				}},
			}},
			{Name: "internal/", Children: []ProjectLayout{
				{Name: "handler/"},
				{Name: "service/"},
				{Name: "repository/"},
			}},
			{Name: "go.mod"},
			{Name: "Makefile"},
			{Name: "README.md"},
		},
	}
	mediumProject.Print("", false)
	fmt.Println()

	// --- 3. 小型项目布局 ---
	fmt.Println("--- 3. 小型项目布局（CLI 工具、简单服务）---")
	fmt.Println()

	smallProject := ProjectLayout{
		Name: "mytool",
		Children: []ProjectLayout{
			{Name: "main.go", Description: "所有代码在一个文件中"},
			{Name: "config.go", Description: "配置相关"},
			{Name: "handler.go", Description: "业务逻辑"},
			{Name: "go.mod"},
			{Name: "README.md"},
		},
	}
	smallProject.Print("", false)
	fmt.Println()

	// --- 4. 关键规则 ---
	fmt.Println("=== 项目布局关键规则 ===")
	fmt.Println()
	fmt.Println("1. cmd/ — 应用入口，main.go 保持简洁（只做初始化和启动）")
	fmt.Println("2. internal/ — 私有代码，Go 编译器强制限制外部包不能导入")
	fmt.Println("3. pkg/ — 可被外部导入的公共库（不是所有项目都需要）")
	fmt.Println("4. 小项目不要过度工程化，一个 main.go 就够了")
	fmt.Println("5. 参考: Kubernetes、Docker、etcd 的项目结构")
	fmt.Println()
	fmt.Println("实际应用:")
	fmt.Println("  - Kubernetes: cmd/ 下有 kube-apiserver、kubectl 等多个入口")
	fmt.Println("  - Docker: cmd/dockerd/ 是守护进程入口")
	fmt.Println("  - etcd: cmd/etcd/ 和 cmd/etcdctl/")
	fmt.Println("  - B 站 Kratos: 推荐 cmd/internal/pkg 标准布局")
}
