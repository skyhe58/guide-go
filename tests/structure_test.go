package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// projectRoot 返回项目根目录路径（tests/ 的上一级）
func projectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("获取工作目录失败: %v", err)
	}
	return filepath.Dir(wd)
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// dirExists 检查目录是否存在
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// TestTopLevelDirectories 验证顶层目录结构存在
// Requirements: 1.1
func TestTopLevelDirectories(t *testing.T) {
	root := projectRoot(t)

	dirs := []string{
		"docs",
		"code-examples",
		"docker",
	}

	for _, dir := range dirs {
		t.Run(dir, func(t *testing.T) {
			path := filepath.Join(root, dir)
			if !dirExists(path) {
				t.Errorf("顶层目录 %s 不存在", dir)
			}
		})
	}
}

// TestREADMEExists 验证 README.md 存在
// Requirements: 1.2
func TestREADMEExists(t *testing.T) {
	root := projectRoot(t)
	path := filepath.Join(root, "README.md")
	if !fileExists(path) {
		t.Error("README.md 不存在")
	}
}

// TestCONTRIBUTINGExists 验证 CONTRIBUTING.md 存在
// Requirements: 24.2
func TestCONTRIBUTINGExists(t *testing.T) {
	root := projectRoot(t)
	path := filepath.Join(root, "CONTRIBUTING.md")
	if !fileExists(path) {
		t.Error("CONTRIBUTING.md 不存在")
	}
}

// TestGoWorkExists 验证 code-examples/go.work 存在
// Requirements: 24.4
func TestGoWorkExists(t *testing.T) {
	root := projectRoot(t)
	path := filepath.Join(root, "code-examples", "go.work")
	if !fileExists(path) {
		t.Error("code-examples/go.work 不存在")
	}
}

// TestDockerComposeFiles 验证所有 Docker Compose 文件存在
// Requirements: 25.1
func TestDockerComposeFiles(t *testing.T) {
	root := projectRoot(t)

	files := []string{
		"docker-compose.yml",
		"docker-compose.mq.yml",
		"docker-compose.es.yml",
		"docker-compose.etcd.yml",
		"docker-compose.consul.yml",
		"docker-compose.nginx.yml",
		"docker-compose.auth.yml",
		"docker-compose.localstack.yml",
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			path := filepath.Join(root, "docker", file)
			if !fileExists(path) {
				t.Errorf("Docker Compose 文件 docker/%s 不存在", file)
			}
		})
	}
}

// TestEntryTemplateExists 验证知识条目模板存在
// Requirements: 24.1
func TestEntryTemplateExists(t *testing.T) {
	root := projectRoot(t)
	path := filepath.Join(root, "docs", "templates", "entry-template.md")
	if !fileExists(path) {
		t.Error("docs/templates/entry-template.md 不存在")
	}
}

// TestAllModuleDirectoriesExist 验证所有 24 个知识模块目录存在
// Requirements: 1.1
func TestAllModuleDirectoriesExist(t *testing.T) {
	root := projectRoot(t)

	modules := []string{
		// 第一层：语言核心（7 个）
		"docs/1-go-core/1.1-go-basics",
		"docs/1-go-core/1.2-go-advanced",
		"docs/1-go-core/1.3-concurrent",
		"docs/1-go-core/1.4-runtime",
		"docs/1-go-core/1.5-testing",
		"docs/1-go-core/1.6-patterns",
		"docs/1-go-core/1.7-algorithm",
		// 第二层：Web 开发与数据（7 个）
		"docs/2-web-data/2.1-web-framework",
		"docs/2-web-data/2.2-database",
		"docs/2-web-data/2.3-cache-search",
		"docs/2-web-data/2.4-message-queue",
		"docs/2-web-data/2.5-object-storage",
		"docs/2-web-data/2.6-auth",
		"docs/2-web-data/2.7-observability",
		// 第三层：微服务与云原生（4 个）
		"docs/3-microservice/3.1-microservice",
		"docs/3-microservice/3.2-service-governance",
		"docs/3-microservice/3.3-docker-k8s",
		"docs/3-microservice/3.4-aws",
		// 第四层：分布式与架构（3 个）
		"docs/4-distributed/4.1-distributed",
		"docs/4-distributed/4.2-architecture",
		"docs/4-distributed/4.3-ai",
		// 第五层：运维与部署（3 个）
		"docs/5-devops/5.1-cicd",
		"docs/5-devops/5.2-linux",
		"docs/5-devops/5.3-nginx",
	}

	for _, mod := range modules {
		t.Run(mod, func(t *testing.T) {
			path := filepath.Join(root, mod)
			if !dirExists(path) {
				t.Errorf("知识模块目录 %s 不存在", mod)
			}
		})
	}

	// 验证总数为 24
	if len(modules) != 24 {
		t.Errorf("期望 24 个知识模块目录，实际定义了 %d 个", len(modules))
	}
}

// TestAllModulesHaveIndexMD 验证所有 24 个模块目录下存在 index.md
// Requirements: 1.3
func TestAllModulesHaveIndexMD(t *testing.T) {
	root := projectRoot(t)

	modules := []string{
		"docs/1-go-core/1.1-go-basics",
		"docs/1-go-core/1.2-go-advanced",
		"docs/1-go-core/1.3-concurrent",
		"docs/1-go-core/1.4-runtime",
		"docs/1-go-core/1.5-testing",
		"docs/1-go-core/1.6-patterns",
		"docs/1-go-core/1.7-algorithm",
		"docs/2-web-data/2.1-web-framework",
		"docs/2-web-data/2.2-database",
		"docs/2-web-data/2.3-cache-search",
		"docs/2-web-data/2.4-message-queue",
		"docs/2-web-data/2.5-object-storage",
		"docs/2-web-data/2.6-auth",
		"docs/2-web-data/2.7-observability",
		"docs/3-microservice/3.1-microservice",
		"docs/3-microservice/3.2-service-governance",
		"docs/3-microservice/3.3-docker-k8s",
		"docs/3-microservice/3.4-aws",
		"docs/4-distributed/4.1-distributed",
		"docs/4-distributed/4.2-architecture",
		"docs/4-distributed/4.3-ai",
		"docs/5-devops/5.1-cicd",
		"docs/5-devops/5.2-linux",
		"docs/5-devops/5.3-nginx",
	}

	for _, mod := range modules {
		t.Run(mod+"/index.md", func(t *testing.T) {
			path := filepath.Join(root, mod, "index.md")
			if !fileExists(path) {
				t.Errorf("模块 %s 缺少 index.md", mod)
			}
		})
	}
}

// TestAllModulesHaveInterviewMD 验证所有 24 个模块目录下存在 interview.md
// Requirements: 2.6
func TestAllModulesHaveInterviewMD(t *testing.T) {
	root := projectRoot(t)

	modules := []string{
		"docs/1-go-core/1.1-go-basics",
		"docs/1-go-core/1.2-go-advanced",
		"docs/1-go-core/1.3-concurrent",
		"docs/1-go-core/1.4-runtime",
		"docs/1-go-core/1.5-testing",
		"docs/1-go-core/1.6-patterns",
		"docs/1-go-core/1.7-algorithm",
		"docs/2-web-data/2.1-web-framework",
		"docs/2-web-data/2.2-database",
		"docs/2-web-data/2.3-cache-search",
		"docs/2-web-data/2.4-message-queue",
		"docs/2-web-data/2.5-object-storage",
		"docs/2-web-data/2.6-auth",
		"docs/2-web-data/2.7-observability",
		"docs/3-microservice/3.1-microservice",
		"docs/3-microservice/3.2-service-governance",
		"docs/3-microservice/3.3-docker-k8s",
		"docs/3-microservice/3.4-aws",
		"docs/4-distributed/4.1-distributed",
		"docs/4-distributed/4.2-architecture",
		"docs/4-distributed/4.3-ai",
		"docs/5-devops/5.1-cicd",
		"docs/5-devops/5.2-linux",
		"docs/5-devops/5.3-nginx",
	}

	for _, mod := range modules {
		t.Run(mod+"/interview.md", func(t *testing.T) {
			path := filepath.Join(root, mod, "interview.md")
			if !fileExists(path) {
				t.Errorf("模块 %s 缺少 interview.md", mod)
			}
		})
	}
}

// TestREADMEHasCompletionTable 验证 README.md 包含完成度追踪表
// Requirements: 24.5
func TestREADMEHasCompletionTable(t *testing.T) {
	root := projectRoot(t)
	path := filepath.Join(root, "README.md")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取 README.md 失败: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "完成度追踪") {
		t.Error("README.md 缺少完成度追踪表")
	}
}

// TestCodeExampleGroupDirs 验证代码示例分组目录存在
// Requirements: 26.2
func TestCodeExampleGroupDirs(t *testing.T) {
	root := projectRoot(t)

	groups := []string{
		"code-examples/01-go-core",
		"code-examples/02-web-data",
		"code-examples/03-microservice",
		"code-examples/04-distributed",
		"code-examples/05-devops",
	}

	for _, group := range groups {
		t.Run(group, func(t *testing.T) {
			path := filepath.Join(root, group)
			if !dirExists(path) {
				t.Errorf("代码示例分组目录 %s 不存在", group)
			}
		})
	}
}
