# 快速开始

## 环境准备

### 1. 安装 Go

本知识库要求 **Go 1.22+**，推荐使用最新稳定版本。

```bash
# macOS (Homebrew)
brew install go

# 验证安装
go version
```

前往 [Go 官方下载页](https://go.dev/dl/) 获取其他平台安装包。

### 2. 克隆项目

```bash
git clone https://github.com/your-username/guide-go.git
cd guide-go
```

### 3. 编译代码示例

项目使用 Go Workspace（`go.work`）管理所有代码模块，一条命令即可编译全部示例：

```bash
cd code-examples && go build ./...
```

### 4. 本地预览文档站点

```bash
cd docs
pnpm install
pnpm run dev
```

浏览器访问 `http://localhost:5173` 即可预览知识库站点。

## 项目结构

```
guide-go/
├── docs/                    # VitePress 知识文档
│   ├── 1-go-core/          # 第一层：语言核心
│   ├── 2-web-data/         # 第二层：Web 开发与数据
│   ├── 3-microservice/     # 第三层：微服务与云原生
│   ├── 4-distributed/      # 第四层：分布式与架构
│   ├── 5-devops/           # 第五层：运维与部署
│   ├── interview/          # 面试汇总
│   └── learning-paths/     # 学习路径
├── code-examples/           # Go Workspace 多模块代码
│   ├── go.work             # Go Workspace 配置
│   ├── 01-go-core/         # 语言核心代码示例
│   ├── 02-web-data/        # Web 与数据代码示例
│   ├── 03-microservice/    # 微服务代码示例
│   ├── 04-distributed/     # 分布式代码示例
│   └── 05-devops/          # 运维部署配置示例
├── docker/                  # Docker Compose 中间件配置
└── .github/                 # GitHub Actions CI/CD
```

## 代码示例运行方式

代码示例采用 **Part A + Part B 混合模式**：

- **Part A（直接运行）**：纯内存模拟，无需任何外部依赖，直接 `go run` 理解原理
- **Part B（需要 Docker）**：连接真实中间件，传入参数 `real` 启动

```bash
# Part A：直接运行，理解原理
cd code-examples/01-go-core/go-basics
go run ./slice/main.go

# Part B：连接真实中间件（需先启动 Docker）
docker compose -f docker/docker-compose.yml up -d redis
cd code-examples/02-web-data/cache-search
go run ./redis/main.go real
```

## 推荐学习路径

根据你的经验水平选择合适的学习路径：

| 路径 | 适合人群 | 预计时间 |
|------|---------|---------|
| [初学者路径](/learning-paths/beginner) | 零基础或其他语言转 Go | 8-12 周 |
| [中级进阶路径](/learning-paths/intermediate) | 有 Go 基础，想深入 | 6-8 周 |
| [高级深入路径](/learning-paths/advanced) | 有项目经验，想精通 | 4-6 周 |
| [面试突击路径](/learning-paths/interview-sprint) | 准备 Go 面试 | 2-4 周 |
| [云原生工程师路径](/learning-paths/cloud-native) | 目标云原生方向 | 6-10 周 |

## 下一步

- 📖 阅读 [使用指南](/guide/how-to-use) 了解知识库的组织方式
- 🚀 从 [Go 基础语法](/1-go-core/1.1-go-basics/) 开始你的学习之旅
- 🎯 查看 [面试知识图谱](/interview/knowledge-map) 了解面试重点
