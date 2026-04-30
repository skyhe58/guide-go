# GoBlog Makefile
# 博客平台 REST API 后端服务构建自动化

# 变量定义
APP_NAME := goblog
BUILD_DIR := build
MAIN_FILE := ./cmd/goblog/main.go
MIGRATE_DIR := migrations
DB_DSN ?= "postgres://postgres:postgres123@localhost:5432/goblog?sslmode=disable"

# 默认目标
.PHONY: all
all: build

# 编译项目
.PHONY: build
build:
	@echo ">>> 编译 $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo ">>> 编译完成: $(BUILD_DIR)/$(APP_NAME)"

# 启动服务
.PHONY: run
run:
	@echo ">>> 启动 $(APP_NAME)..."
	go run $(MAIN_FILE)

# 运行测试
.PHONY: test
test:
	@echo ">>> 运行测试..."
	go test -v -race -cover ./...

# 代码检查（golangci-lint）
.PHONY: lint
lint:
	@echo ">>> 运行代码检查..."
	golangci-lint run ./...

# 数据库迁移 - 升级
.PHONY: migrate-up
migrate-up:
	@echo ">>> 执行数据库迁移（升级）..."
	migrate -path $(MIGRATE_DIR) -database $(DB_DSN) up

# 数据库迁移 - 回滚
.PHONY: migrate-down
migrate-down:
	@echo ">>> 执行数据库迁移（回滚）..."
	migrate -path $(MIGRATE_DIR) -database $(DB_DSN) down

# 启动 Docker 依赖中间件（PostgreSQL + Redis）
.PHONY: docker-up
docker-up:
	@echo ">>> 启动 Docker 依赖中间件..."
	docker compose up -d

# 停止 Docker 依赖中间件
.PHONY: docker-down
docker-down:
	@echo ">>> 停止 Docker 依赖中间件..."
	docker compose down

# 生成 Swagger API 文档
.PHONY: swagger
swagger:
	@echo ">>> 生成 Swagger 文档..."
	swag init -g $(MAIN_FILE) -o docs

# 生成 Wire 依赖注入代码
.PHONY: wire
wire:
	@echo ">>> 生成 Wire 依赖注入代码..."
	cd internal/wire && wire

# 清理构建产物
.PHONY: clean
clean:
	@echo ">>> 清理构建产物..."
	rm -rf $(BUILD_DIR)

.PHONY: help
help:
	@echo "可用的 Make 目标："
	@echo "  build         - 编译项目"
	@echo "  run           - 启动服务"
	@echo "  test          - 运行测试"
	@echo "  lint          - 代码检查（golangci-lint）"
	@echo "  migrate-up    - 数据库迁移（升级）"
	@echo "  migrate-down  - 数据库迁移（回滚）"
	@echo "  docker-up     - 启动 Docker 依赖中间件"
	@echo "  docker-down   - 停止 Docker 依赖中间件"
	@echo "  swagger       - 生成 Swagger API 文档"
	@echo "  wire          - 生成 Wire 依赖注入代码"
	@echo "  clean         - 清理构建产物"
