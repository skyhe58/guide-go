# 贡献指南

感谢你对 Go 从入门到精通知识库的关注！本文档说明如何为本项目贡献内容。

## 项目结构概览

```
guide-go/
├── docs/                        # 知识文档（Markdown）
│   ├── {层级}-{分组名}/
│   │   └── {模块编号}-{模块名}/
│   │       ├── index.md         # 模块索引
│   │       ├── 01-xxx.md        # 知识条目
│   │       └── interview.md     # 面试指南
│   └── templates/
│       └── entry-template.md    # 知识条目模板
├── code-examples/               # 代码示例（Go Workspace）
│   ├── go.work
│   └── {分组编号}-{分组名}/
│       └── {子模块名}/
│           ├── go.mod
│           └── {知识点}/main.go
└── docker/                      # Docker Compose 配置
```

## 如何添加新的知识条目

### 1. 编写文档

1. 复制模板文件 `docs/templates/entry-template.md` 到目标模块目录
2. 按模块内的编号顺序重命名，如 `05-control-flow.md`
3. 填写 YAML frontmatter 元数据（title、module、difficulty、tags 等）
4. 按模板中的标准章节结构编写内容：
   - 概念说明
   - 核心原理（复杂流程配 Mermaid 图）
   - 标准库方案（Go 哲学：标准库优先）
   - 第三方库方案
   - 代码示例（链接到对应代码目录）
   - 常见面试题
   - 常见陷阱
   - 参考资料

### 2. 编写代码示例

1. 在对应的 `code-examples/` 子模块下创建知识点目录
2. 编写 `main.go`，确保包含：
   - `func main()` 入口函数，可通过 `go run` 直接运行
   - 详细的中文注释
   - 文件头注释标注 Go 版本要求和验证日期
3. 如果依赖外部服务，采用 Part A + Part B 混合模式：
   - Part A：纯内存模拟，直接运行
   - Part B：连接真实中间件，需传入参数 `real`
   - 文件头注释标注 Docker 启动命令

### 3. 更新索引

1. 更新模块的 `index.md`，添加新知识条目的链接
2. 如果是新模块，更新根目录 `README.md` 的完成度追踪表

## 如何添加新模块

### 1. 创建文档目录

```bash
# 在对应层级下创建模块目录
mkdir -p docs/{层级}-{分组名}/{模块编号}-{模块名}/
```

### 2. 创建代码模块

```bash
# 在对应分组下创建 Go Module
mkdir -p {分组编号}-{分组名}/{子模块名}/
cd {分组编号}-{分组名}/{子模块名}/
go mod init guide-go/{子模块名}
```

### 3. 更新 Go Workspace

在 `code-examples/go.work` 中添加新模块路径：

```go
use (
    // ... 已有模块
    ./{分组编号}-{分组名}/{子模块名}
)
```

### 4. 创建必要文件

- `index.md` — 模块索引，列出所有知识点
- `interview.md` — 面试指南
- 各知识条目文档（使用 `entry-template.md` 模板）

## 编写规范

### 文档规范

- 所有文档使用**中文**编写
- 技术术语保留英文，首次出现时附中文解释
- 复杂流程必须包含 Mermaid 流程图或时序图
- 面试题标注难度（⭐）和频率（🔥）

### 代码规范

- Go 版本要求：Go 1.22+
- 代码中包含详细的中文注释
- 遵循 Go 官方代码风格（gofmt）
- 通过 golangci-lint 检查
- 文件头注释格式：

```go
// 知识点名称 - 简要说明
//
// Go 版本要求: Go 1.22+
// 最后验证日期: 2025-01-01
//
// 运行方式:
//   go run main.go
//
// Part B（需 Docker）:
//   docker compose -f docker/docker-compose.yml up -d redis
//   go run main.go real
package main
```

### Git 提交规范

```
feat(模块名): 添加 xxx 知识点文档和代码示例
fix(模块名): 修复 xxx 代码示例中的错误
docs(模块名): 更新 xxx 文档内容
style: 格式化代码/文档
```

## 本地开发

```bash
# 编译所有代码示例
cd code-examples && go build ./...

# 本地预览文档站点
cd docs && pnpm install && pnpm run dev

# 运行测试
cd tests && go test -v ./...
```

## 问题反馈

如果发现文档错误、代码 bug 或有改进建议，欢迎提交 Issue 或 Pull Request。
