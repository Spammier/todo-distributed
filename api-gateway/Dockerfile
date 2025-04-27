# ---- 构建阶段 ----
FROM golang:1.24-alpine as builder

ENV CGO_ENABLED=0 GOOS=linux
ENV GOPROXY https://goproxy.cn,direct

WORKDIR /app

# 复制 Go 模块文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源代码 (包括生成的 proto 代码)
COPY . .

# 编译 Go 应用
RUN go build -a -installsuffix cgo -o /app/api-gateway .

# ---- 运行阶段 ----
FROM alpine:latest

WORKDIR /root/

# 从构建阶段复制编译好的可执行文件
COPY --from=builder /app/api-gateway .

# API 网关监听 8080 端口 (由代码中的 PORT 环境变量决定)
EXPOSE 8080

# 运行应用
CMD ["./api-gateway"] 