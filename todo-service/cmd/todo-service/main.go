package main

import (
	"log"
	"net"

	"todo-project/todo-service/internal/config"
	"todo-project/todo-service/internal/db"
	"todo-project/todo-service/internal/service"
	pb "todo-project/todo-service/proto/todo"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库和Redis
	dbConn := db.InitDB(cfg)
	redisClient := db.InitRedis(cfg)

	// 启动gRPC服务
	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterTodoServiceServer(s, service.NewTodoService(dbConn, redisClient))
	reflection.Register(s)

	log.Printf("Todo service listening on %s (with reflection)", cfg.GRPCPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
