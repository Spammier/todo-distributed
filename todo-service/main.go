package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	// 导入项目内部包
	pb "todo-project/todo-service/proto/todo" // 导入生成的 protobuf 代码

	// 导入外部依赖
	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection" // 启用 gRPC Server Reflection
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// --- 全局变量 ---
var (
	db  *gorm.DB
	rdb *redis.Client
	ctx = context.Background()
)

// --- 模型定义 --- (直接在这里定义，或者可以拆分到 models 包)
type Todo struct {
	ID          uint   `gorm:"primaryKey"`
	UserID      uint   `gorm:"not null;index"`
	Title       string `gorm:"not null"`
	Description string
	Completed   bool `gorm:"default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// --- gRPC 服务器实现 ---
type server struct {
	pb.UnimplementedTodoServiceServer // 嵌入未实现的 server
}

// --- 辅助函数 ---
// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// convertToProtoTodo 将 GORM 模型转换为 Protobuf 消息
func convertToProtoTodo(todoModel *Todo) *pb.Todo {
	return &pb.Todo{
		Id:          uint32(todoModel.ID),
		UserId:      uint32(todoModel.UserID),
		Title:       todoModel.Title,
		Description: todoModel.Description,
		Completed:   todoModel.Completed,
		CreatedAt:   timestamppb.New(todoModel.CreatedAt),
		UpdatedAt:   timestamppb.New(todoModel.UpdatedAt),
	}
}

// convertToProtoTodos 转换 Todo 模型切片
func convertToProtoTodos(todoModels []*Todo) []*pb.Todo {
	protoTodos := make([]*pb.Todo, len(todoModels))
	for i, model := range todoModels {
		protoTodos[i] = convertToProtoTodo(model)
	}
	return protoTodos
}

// --- 实现 TodoService gRPC 方法 ---

func (s *server) CreateTodo(ctx context.Context, req *pb.CreateTodoRequest) (*pb.Todo, error) {
	log.Printf("Received CreateTodo request for user_id: %d, title: %s", req.GetUserId(), req.GetTitle())

	if req.GetUserId() == 0 || req.GetTitle() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "用户 ID 和标题不能为空")
	}

	newTodo := Todo{
		UserID:      uint(req.GetUserId()),
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		Completed:   false, // 默认未完成
	}

	result := db.Create(&newTodo)
	if result.Error != nil {
		log.Printf("创建 Todo 失败: %v", result.Error)
		return nil, status.Errorf(codes.Internal, "创建 Todo 失败")
	}

	log.Printf("Todo 创建成功: ID=%d", newTodo.ID)
	// TODO: 清除 Redis 缓存 (用户列表缓存)

	return convertToProtoTodo(&newTodo), nil
}

func (s *server) GetTodos(ctx context.Context, req *pb.GetTodosRequest) (*pb.GetTodosResponse, error) {
	log.Printf("Received GetTodos request for user_id: %d", req.GetUserId())
	userID := req.GetUserId()
	if userID == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "无效的用户 ID")
	}

	// TODO: 实现 Redis 缓存读取逻辑

	var todos []*Todo
	result := db.Where("user_id = ?", userID).Find(&todos)
	if result.Error != nil {
		log.Printf("获取 Todos 失败 for user %d: %v", userID, result.Error)
		return nil, status.Errorf(codes.Internal, "获取待办事项失败")
	}

	// TODO: 实现 Redis 缓存写入逻辑

	log.Printf("找到 %d 个 Todos for user %d", len(todos), userID)
	return &pb.GetTodosResponse{Todos: convertToProtoTodos(todos)}, nil
}

func (s *server) GetTodoByID(ctx context.Context, req *pb.GetTodoByIDRequest) (*pb.Todo, error) {
	log.Printf("Received GetTodoByID request for user_id: %d, todo_id: %d", req.GetUserId(), req.GetTodoId())
	userID := req.GetUserId()
	todoID := req.GetTodoId()

	if userID == 0 || todoID == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "无效的用户 ID 或 Todo ID")
	}

	// TODO: 实现 Redis 缓存读取逻辑

	var todo Todo
	err := db.Where("id = ? AND user_id = ?", todoID, userID).First(&todo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Todo 未找到: user_id=%d, todo_id=%d", userID, todoID)
			return nil, status.Errorf(codes.NotFound, "待办事项未找到或无权访问")
		}
		log.Printf("获取 Todo %d 失败 for user %d: %v", todoID, userID, err)
		return nil, status.Errorf(codes.Internal, "获取待办事项失败")
	}

	// TODO: 实现 Redis 缓存写入逻辑

	log.Printf("找到 Todo: ID=%d", todo.ID)
	return convertToProtoTodo(&todo), nil
}

func (s *server) UpdateTodo(ctx context.Context, req *pb.UpdateTodoRequest) (*pb.Todo, error) {
	log.Printf("Received UpdateTodo request for user_id: %d, todo_id: %d", req.GetUserId(), req.GetTodoId())
	userID := req.GetUserId()
	todoID := req.GetTodoId()

	if userID == 0 || todoID == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "无效的用户 ID 或 Todo ID")
	}

	// 查找原始 Todo 以进行权限检查
	var originalTodo Todo
	err := db.Where("id = ? AND user_id = ?", todoID, userID).First(&originalTodo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("更新时 Todo 未找到: user_id=%d, todo_id=%d", userID, todoID)
			return nil, status.Errorf(codes.NotFound, "待办事项未找到或无权更新")
		}
		log.Printf("查找待更新 Todo 失败: %v", err)
		return nil, status.Errorf(codes.Internal, "获取待办事项失败")
	}

	// 构建更新映射，只更新请求中提供的字段
	updates := map[string]interface{}{
		"title":       req.GetTitle(),
		"description": req.GetDescription(),
		"completed":   req.GetCompleted(),
		// 更新时间会自动处理
	}

	// 执行更新
	result := db.Model(&originalTodo).Updates(updates)
	if result.Error != nil {
		log.Printf("更新 Todo %d 失败: %v", todoID, result.Error)
		return nil, status.Errorf(codes.Internal, "更新待办事项失败")
	}
	if result.RowsAffected == 0 {
		// 理论上不应该发生，因为我们已经先查到了记录
		log.Printf("更新 Todo %d 时影响行数为 0", todoID)
		// 可以选择返回 NotFound 或 Internal 错误
		return nil, status.Errorf(codes.Internal, "更新失败，记录可能已不存在")
	}

	log.Printf("Todo %d 更新成功", todoID)
	// TODO: 清除相关 Redis 缓存 (用户列表和单个 Todo 缓存)

	// 返回更新后的 Todo (从数据库重新获取以确保数据最新)
	var updatedTodo Todo
	db.First(&updatedTodo, todoID)

	return convertToProtoTodo(&updatedTodo), nil
}

func (s *server) DeleteTodo(ctx context.Context, req *pb.DeleteTodoRequest) (*emptypb.Empty, error) {
	log.Printf("Received DeleteTodo request for user_id: %d, todo_id: %d", req.GetUserId(), req.GetTodoId())
	userID := req.GetUserId()
	todoID := req.GetTodoId()

	if userID == 0 || todoID == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "无效的用户 ID 或 Todo ID")
	}

	// 直接尝试删除，GORM 的 Delete 会返回影响的行数
	result := db.Where("id = ? AND user_id = ?", todoID, userID).Delete(&Todo{})

	if result.Error != nil {
		log.Printf("删除 Todo %d 失败: %v", todoID, result.Error)
		return nil, status.Errorf(codes.Internal, "删除待办事项失败")
	}

	if result.RowsAffected == 0 {
		log.Printf("删除时 Todo 未找到: user_id=%d, todo_id=%d", userID, todoID)
		return nil, status.Errorf(codes.NotFound, "待办事项未找到或无权删除")
	}

	log.Printf("Todo %d 删除成功", todoID)
	// TODO: 清除相关 Redis 缓存

	return &emptypb.Empty{}, nil
}

// --- 主函数 ---
func main() {
	// ---- 1. 初始化数据库连接 ----
	dbUser := getEnvOrDefault("DB_USER", "root")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := getEnvOrDefault("DB_HOST", "localhost")
	dbPort := getEnvOrDefault("DB_PORT", "3306")
	dbName := getEnvOrDefault("DB_NAME", "todo")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("无法初始化数据库: %v", err)
	}
	fmt.Println("成功连接到数据库")

	// ---- 2. 自动迁移数据库表结构 ----
	err = db.AutoMigrate(&Todo{})
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// ---- 3. 初始化 Redis 连接 ----
	redisAddr := getEnvOrDefault("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnvOrDefault("REDIS_PASSWORD", "")
	redisDBStr := getEnvOrDefault("REDIS_DB", "0")
	redisDBNum, err := strconv.Atoi(redisDBStr)
	if err != nil {
		redisDBNum = 0
		log.Printf("Redis DB 配置无效，使用默认值 0: %v\n", err)
	}
	rdb = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDBNum,
	})
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		// Redis 连接失败不应阻塞服务启动，打印警告即可
		log.Printf("警告: Redis 连接失败: %v", err)
	} else {
		fmt.Println("成功连接到 Redis")
	}

	// ---- 4. 设置 gRPC 服务器 ----
	port := ":50052" // 为 todo-service 使用不同的端口，例如 50052
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterTodoServiceServer(s, &server{})
	reflection.Register(s) // 启用反射

	log.Printf("Todo service listening on %s (with reflection)", port)

	// ---- 5. 启动 gRPC 服务器 ----
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
