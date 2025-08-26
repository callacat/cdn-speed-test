# --- Build Stage ---
FROM golang:1.25-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 Go 模块文件
COPY go.mod go.sum ./
# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
# -ldflags="-w -s" 用于减小二进制文件大小
# CGO_ENABLED=0 禁用CGO，确保静态链接
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /cdn-speed-test ./cmd/cdn-speed-test

# --- Final Stage ---
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件和必要文件
COPY --from=builder /cdn-speed-test /app/cdn-speed-test
# 如果需要默认的ip.txt或config.yaml，也在这里复制
# COPY ip.txt .
# COPY config.yaml .

# 暴露端口 (如果需要)
# EXPOSE 8080

# 容器启动时执行的命令
ENTRYPOINT ["/app/cdn-speed-test"]
