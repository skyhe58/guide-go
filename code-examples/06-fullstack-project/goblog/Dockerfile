# GoBlog 多阶段构建 Dockerfile
# 构建产物为静态链接的单一二进制文件，基于 scratch 镜像运行
# 最终镜像大小 ≤ 20MB
#
# 构建命令：docker build -t goblog .
# 运行命令：docker run -p 8080:8080 goblog

# ==================== 阶段一：编译 ====================
FROM golang:1.22-alpine AS builder

# 安装必要的构建工具（用于 CGO 或证书）
RUN apk add --no-cache ca-certificates tzdata

# 设置工作目录
WORKDIR /build

# 先复制依赖文件，利用 Docker 缓存层加速构建
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译：禁用 CGO，生成静态链接二进制文件
# -ldflags="-s -w" 去除调试信息和符号表，减小二进制体积
# -trimpath 去除编译路径信息，提高安全性
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -trimpath \
    -o goblog \
    ./cmd/goblog/

# ==================== 阶段二：运行 ====================
FROM scratch

# 从 builder 阶段复制 CA 证书（HTTPS 请求需要）
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 从 builder 阶段复制时区数据
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# 复制配置文件
COPY --from=builder /build/configs/config.yaml /configs/config.yaml

# 复制数据库迁移文件
COPY --from=builder /build/migrations /migrations

# 复制编译好的二进制文件
COPY --from=builder /build/goblog /goblog

# 设置时区环境变量
ENV TZ=Asia/Shanghai

# 暴露服务端口
EXPOSE 8080

# 启动服务
ENTRYPOINT ["/goblog"]
