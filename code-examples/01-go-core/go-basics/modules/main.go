// Go 版本要求: Go 1.22+
// 最后验证日期: 2025-01-01
// Go Module 使用示例：模块信息、构建信息、包的组织
package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

func main() {
	fmt.Println("========== Go Module 使用示例 ==========")

	// ========== 1. 运行时信息 ==========
	fmt.Println("\n--- 1. Go 运行时信息 ---")
	fmt.Printf("Go 版本: %s\n", runtime.Version())
	fmt.Printf("操作系统: %s\n", runtime.GOOS)
	fmt.Printf("架构: %s\n", runtime.GOARCH)
	fmt.Printf("CPU 核数: %d\n", runtime.NumCPU())
	fmt.Printf("GOROOT: %s\n", runtime.GOROOT())

	// ========== 2. 构建信息（Go Module 元数据） ==========
	fmt.Println("\n--- 2. 构建信息 ---")
	info, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Println("无法读取构建信息")
		return
	}

	fmt.Printf("模块路径: %s\n", info.Main.Path)
	fmt.Printf("Go 版本: %s\n", info.GoVersion)

	if info.Main.Version != "" {
		fmt.Printf("模块版本: %s\n", info.Main.Version)
	}

	// 打印依赖信息
	if len(info.Deps) > 0 {
		fmt.Println("\n依赖列表:")
		for _, dep := range info.Deps {
			fmt.Printf("  %s %s\n", dep.Path, dep.Version)
		}
	} else {
		fmt.Println("\n当前模块无外部依赖（纯标准库）")
	}

	// 构建设置
	fmt.Println("\n构建设置:")
	for _, setting := range info.Settings {
		if setting.Key == "GOOS" || setting.Key == "GOARCH" || setting.Key == "vcs" || setting.Key == "vcs.revision" {
			fmt.Printf("  %s = %s\n", setting.Key, setting.Value)
		}
	}

	// ========== 3. Go Module 常用命令速查 ==========
	fmt.Println("\n--- 3. Go Module 常用命令 ---")
	commands := []struct {
		cmd  string
		desc string
	}{
		{"go mod init <module>", "初始化模块"},
		{"go mod tidy", "整理依赖（添加缺失/移除多余）"},
		{"go mod download", "下载依赖到本地缓存"},
		{"go mod vendor", "将依赖复制到 vendor 目录"},
		{"go mod graph", "打印依赖图"},
		{"go mod why <pkg>", "解释为什么需要某个依赖"},
		{"go get <pkg>@latest", "添加/更新依赖到最新版本"},
		{"go get <pkg>@v1.2.3", "添加/更新依赖到指定版本"},
		{"go work init", "初始化 Go Workspace"},
		{"go work use ./path", "将模块添加到 Workspace"},
	}

	for _, c := range commands {
		fmt.Printf("  %-30s  %s\n", c.cmd, c.desc)
	}

	// ========== 4. 包的组织原则 ==========
	fmt.Println("\n--- 4. Go 包的组织原则 ---")
	principles := []string{
		"一个目录 = 一个包（同目录下所有 .go 文件属于同一个包）",
		"包名 = 目录名（惯例，非强制）",
		"main 包是程序入口，必须有 main() 函数",
		"首字母大写 = 导出（public），小写 = 未导出（private）",
		"避免循环依赖（Go 编译器不允许）",
		"internal/ 目录下的包只能被父目录及其子目录导入",
		"cmd/ 目录存放可执行程序的 main 包",
	}

	for i, p := range principles {
		fmt.Printf("  %d. %s\n", i+1, p)
	}

	// ========== 5. go.mod 文件结构说明 ==========
	fmt.Println("\n--- 5. go.mod 文件结构 ---")
	fmt.Println(`  module github.com/yourname/project  // 模块路径
  go 1.22                              // Go 版本要求
  
  require (
      github.com/gin-gonic/gin v1.9.1  // 直接依赖
  )
  
  require (
      github.com/some/pkg v1.0.0 // indirect  // 间接依赖
  )
  
  replace (
      github.com/old/pkg => ../local-pkg  // 本地替换
  )`)

	fmt.Println("\n========== 示例结束 ==========")
}
