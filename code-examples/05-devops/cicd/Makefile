# =============================================================================
# Go 项目 Makefile 示例
# 适用于：标准 Go 后端项目的构建自动化
# 功能：编译构建 / 单元测试 / 代码检查 / Docker 镜像 / 代码生成
#
# 使用说明：
#   make help      - 查看所有可用命令
#   make build     - 编译 Go 二进制文件
#   make test      - 运行单元测试
#   make lint      - 运行代码检查
#   make docker    - 构建 Docker 镜像
#
# Go 版本：Go 1.22+
# 最后验证：2024-12-01
# =============================================================================

# ============================================================================
# 变量定义
# ============================================================================

# 项目信息
APP_NAME    := myapp
CMD_DIR     := ./cmd/server
BIN_DIR     := ./bin
DIST_DIR    := ./dist

# 版本信息（从 Git 自动获取）
VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT      := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE  := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GO_VERSION  := $(shell go version | awk '{print $$3}')

# 编译参数
LDFLAGS     := -s -w \
               -X main.version=$(VERSION) \
               -X main.commit=$(COMMIT) \
               -X main.date=$(BUILD_DATE)

# Docker 配置
DOCKER_IMAGE := $(APP_NAME)
DOCKER_TAG   := $(VERSION)

# ============================================================================
# 构建相关
# ============================================================================

.PHONY: build
build: ## 编译 Go 二进制文件
	@echo "==> 编译 $(APP_NAME) $(VERSION)..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME) $(CMD_DIR)
	@echo "==> 编译完成: $(BIN_DIR)/$(APP_NAME)"

.PHONY: build-all
build-all: ## 交叉编译多平台二进制文件
	@echo "==> 交叉编译 $(APP_NAME) $(VERSION)..."
	@mkdir -p $(BIN_DIR)
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME)-linux-amd64 $(CMD_DIR)
	GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME)-linux-arm64 $(CMD_DIR)
	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME)-darwin-amd64 $(CMD_DIR)
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME)-darwin-arm64 $(CMD_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME)-windows-amd64.exe $(CMD_DIR)
	@echo "==> 交叉编译完成"

.PHONY: run
run: ## 启动服务（开发模式）
	go run $(CMD_DIR)

# ============================================================================
# 测试相关
# ============================================================================

.PHONY: test
test: ## 运行单元测试（含竞态检测和覆盖率）
	@echo "==> 运行单元测试..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo "==> 测试完成"

.PHONY: test-short
test-short: ## 运行短测试（跳过集成测试）
	go test -v -short ./...

.PHONY: coverage
coverage: test ## 生成并打开覆盖率报告
	go tool cover -html=coverage.out -o coverage.html
	@echo "==> 覆盖率报告: coverage.html"

.PHONY: bench
bench: ## 运行基准测试
	go test -bench=. -benchmem ./...

# ============================================================================
# 代码质量
# ============================================================================

.PHONY: lint
lint: ## 运行 golangci-lint 代码检查
	@echo "==> 运行代码检查..."
	golangci-lint run --timeout=5m ./...

.PHONY: fmt
fmt: ## 格式化代码
	@echo "==> 格式化代码..."
	gofmt -s -w .
	goimports -w .

.PHONY: vet
vet: ## 运行 go vet 静态分析
	go vet ./...

.PHONY: tidy
tidy: ## 整理 Go Module 依赖
	go mod tidy
	go mod verify

# ============================================================================
# 代码生成
# ============================================================================

.PHONY: generate
generate: ## 运行 go generate（protobuf、mock、wire 等）
	@echo "==> 运行代码生成..."
	go generate ./...

.PHONY: swagger
swagger: ## 生成 Swagger API 文档
	@echo "==> 生成 Swagger 文档..."
	swag init -g $(CMD_DIR)/main.go -o ./docs/swagger

# ============================================================================
# Docker 相关
# ============================================================================

.PHONY: docker
docker: ## 构建 Docker 镜像
	@echo "==> 构建 Docker 镜像 $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-t $(DOCKER_IMAGE):latest \
		.

.PHONY: docker-push
docker-push: ## 推送 Docker 镜像
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest

.PHONY: docker-up
docker-up: ## 启动依赖中间件（Docker Compose）
	docker compose -f docker/docker-compose.yml up -d

.PHONY: docker-down
docker-down: ## 停止依赖中间件
	docker compose -f docker/docker-compose.yml down

# ============================================================================
# 数据库迁移
# ============================================================================

.PHONY: migrate-up
migrate-up: ## 执行数据库迁移
	migrate -path migrations -database "$(DB_URL)" up

.PHONY: migrate-down
migrate-down: ## 回滚数据库迁移（最近一次）
	migrate -path migrations -database "$(DB_URL)" down 1

.PHONY: migrate-create
migrate-create: ## 创建新的迁移文件（用法：make migrate-create NAME=create_users）
	migrate create -ext sql -dir migrations -seq $(NAME)

# ============================================================================
# 清理与帮助
# ============================================================================

.PHONY: clean
clean: ## 清理构建产物
	@echo "==> 清理构建产物..."
	rm -rf $(BIN_DIR) $(DIST_DIR) coverage.out coverage.html

.PHONY: info
info: ## 显示项目信息
	@echo "应用名称:  $(APP_NAME)"
	@echo "版本:      $(VERSION)"
	@echo "Commit:    $(COMMIT)"
	@echo "构建日期:  $(BUILD_DATE)"
	@echo "Go 版本:   $(GO_VERSION)"

.PHONY: ci
ci: lint test build ## CI 流水线（lint + test + build）
	@echo "==> CI 流水线完成"

.PHONY: help
help: ## 显示帮助信息
	@echo "$(APP_NAME) Makefile 命令列表："
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# 默认目标
.DEFAULT_GOAL := help
