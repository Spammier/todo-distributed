# ---- 构建阶段 ----
FROM golang:1.24-alpine as builder

ENV CGO_ENABLED=0 GOOS=linux
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

# 复制Go模块文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源代码
COPY . .

# 编译Go应用
RUN go build -a -installsuffix cgo -o /app/api-gateway ./cmd/api-gateway

# ---- 运行阶段 ----
FROM alpine:latest

WORKDIR /root/

# 从构建阶段复制编译好的可执行文件
COPY --from=builder /app/api-gateway .

# 设置环境变量
ENV APP_ENV=container

# API网关监听8080端口
EXPOSE 8080

# 运行应用
CMD ["./api-gateway"] 