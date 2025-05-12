package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"todo-project/api-gateway/internal/config"
	"todo-project/api-gateway/internal/handlers"
	todopb "todo-project/api-gateway/proto/todo"
	userpb "todo-project/api-gateway/proto/user"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()
	if string(cfg.JWTSecretKey) == "" {
		log.Fatal("JWT_SECRET_KEY 环境变量未设置!")
	}

	// 建立gRPC连接
	// 连接User Service
	userConn, err := grpc.Dial(cfg.UserServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("无法连接到User Service: %v", err)
	}
	defer userConn.Close()
	userClient := userpb.NewUserServiceClient(userConn)
	log.Printf("已连接到User Service at %s", cfg.UserServiceAddr)

	// 连接Todo Service
	todoConn, err := grpc.Dial(cfg.TodoServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("无法连接到Todo Service: %v", err)
	}
	defer todoConn.Close()
	todoClient := todopb.NewTodoServiceClient(todoConn)
	log.Printf("已连接到Todo Service at %s", cfg.TodoServiceAddr)

	// 设置Gin HTTP服务器
	router := gin.Default()

	// 设置路由
	handlers.SetupRouter(router, userClient, todoClient, cfg.JWTSecretKey)

	// 启动HTTP服务器
	log.Printf("API Gateway listening on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to run API Gateway: %v", err)
	}
}
