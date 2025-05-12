package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"todo-project/user-service/internal/config"
	"todo-project/user-service/internal/database"
	"todo-project/user-service/internal/mq"
	"todo-project/user-service/internal/service"
	pb "todo-project/user-service/proto/user"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()
	log.Printf("当前运行环境: %s", cfg.AppEnv)

	if string(cfg.JWTSecretKey) == "" {
		log.Fatal("JWT_SECRET_KEY 环境变量未设置!")
	}

	// 初始化数据库连接
	if err := database.InitDB(cfg); err != nil {
		log.Fatalf("无法初始化数据库: %v", err)
	}

	// 初始化RabbitMQ连接
	if err := mq.ConnectRabbitMQ(cfg); err != nil {
		// 记录错误但允许服务继续运行
		log.Printf("无法连接到RabbitMQ: %v. 事件发布功能将不可用。", err)
	} else {
		// 仅在成功连接时延迟关闭
		defer mq.CloseConnections()
	}

	// 设置gRPC服务器
	port := ":50051"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	userService := service.NewUserService(cfg.JWTSecretKey)
	pb.RegisterUserServiceServer(s, userService)

	// 注册反射服务，便于调试
	reflection.Register(s)

	log.Printf("User service listening on %s (with reflection)", port)

	// 启动gRPC服务器
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
