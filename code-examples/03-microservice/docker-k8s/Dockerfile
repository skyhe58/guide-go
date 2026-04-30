# ============================================
# Go 应用多阶段构建 Dockerfile
# 最佳实践：scratch 基础镜像 + CGO_ENABLED=0
# 最终镜像大小：≤ 10MB
#
# 构建命令：docker build -t my-go-app:latest .
# 运行命令：docker run -p 8080:8080 my-go-app:latest
#
# Go 版本要求：Go 1.22+
# 最后验证日期：2024-12-01
# ============================================

# ============================================
# 阶段 1：构建阶段（Builder）
# 使用 golang:1.22-alpine 作为构建环境
# alpine 版本比 debian 版本小很多（~250MB vs ~800MB）
# ============================================
FROM golang:1.22-alpine AS builder

# 安装构建所需的系统依赖
# ca-certificates: HTTPS 请求需要的 CA 证书
# git: 如果 go.mod 中有私有仓库依赖需要 git
# tzdata: 时区数据，scratch 镜像中没有
RUN apk add --no-cache ca-certificates git tzdata

# 设置工作目录
WORKDIR /build

# ---- 依赖缓存层 ----
# 先复制依赖文件，利用 Docker 层缓存
# 只有 go.mod 或 go.sum 变化时才会重新下载依赖
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# ---- 编译层 ----
# 复制全部源码
COPY . .

# 编译 Go 应用
# CGO_ENABLED=0: 禁用 CGO，生成纯静态二进制文件（可在 scratch 上运行）
# GOOS=linux:    目标操作系统为 Linux
# GOARCH=amd64:  目标 CPU 架构（生产环境常用 amd64，ARM 服务器用 arm64）
# -ldflags:
#   -s: 去掉符号表（减小体积）
#   -w: 去掉 DWARF 调试信息（减小体积）
#   -X main.version=...: 注入版本信息（可选）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
    -trimpath \
    -o /app \
    ./cmd/server

# ============================================
# 阶段 2：运行阶段（Runtime）
# 使用 scratch 空镜像，最终镜像只包含：
# - 编译好的二进制文件
# - CA 证书（HTTPS 请求需要）
# - 时区数据（time.LoadLocation 需要）
# ============================================
FROM scratch

# 元数据标签
LABEL maintainer="go-knowledge-base"
LABEL description="Go 应用生产级 Docker 镜像"

# 从构建阶段复制 CA 证书（HTTPS 请求需要）
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 从构建阶段复制时区数据（time.LoadLocation 需要）
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# 设置时区环境变量
ENV TZ=Asia/Shanghai

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app /app

# 暴露服务端口
EXPOSE 8080

# 设置入口点
# 使用 exec 格式（JSON 数组），确保 Go 进程是 PID 1
# 这样 Docker stop 发送的 SIGTERM 信号能直接到达 Go 进程
ENTRYPOINT ["/app"]
