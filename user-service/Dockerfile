# ---- 构建阶段 ----
FROM golang:1.24-alpine as builder

# 设置 Go 环境变量
ENV CGO_ENABLED=0 GOOS=linux
ENV GOPROXY https://goproxy.cn,direct

WORKDIR /app

# 复制 Go 模块文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源代码
# 注意：这里将整个 /app 目录复制，如果 .proto 文件在别处，需要调整
COPY . .

# 编译 Go 应用
# 将可执行文件命名为 'user-service'
RUN go build -a -installsuffix cgo -o /app/user-service .

# ---- 运行阶段 ----
FROM alpine:latest

WORKDIR /root/

# 从构建阶段复制编译好的可执行文件
COPY --from=builder /app/user-service .

# (可选) 复制任何需要的配置文件，例如 CA 证书
# COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 暴露 gRPC 服务监听的端口
EXPOSE 50051

# 运行 Go 应用
CMD ["./user-service"] 