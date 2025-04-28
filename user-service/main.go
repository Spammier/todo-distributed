package main

import (
	"context"
	"encoding/json" // Import encoding/json
	"fmt"
	"log"
	"net"
	"os"
	"time"

	// 导入项目内部包
	"todo-project/user-service/database"
	"todo-project/user-service/models"
	pb "todo-project/user-service/proto/user" // 导入生成的 protobuf 代码

	// 导入外部依赖
	"github.com/golang-jwt/jwt/v4"
	"github.com/rabbitmq/amqp091-go" // Import RabbitMQ package
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection" // 导入 reflection 包
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// Add RabbitMQ connection and channel variables
var (
	rabbitConn    *amqp091.Connection
	rabbitChannel *amqp091.Channel
	jwtKey        []byte // JWT 密钥 - Keep this
)

// server 结构体将实现 UserServiceServer 接口
type server struct {
	pb.UnimplementedUserServiceServer // 嵌入未实现的 server 以确保向前兼容
	// jwtKey is now a global variable
}

// Define the queue name
const userRegisteredQueue = "user_registered_queue"

// --- 新增：RabbitMQ 连接重试参数 ---
const (
	maxRabbitMQRetries = 5               // 最大重试次数
	rabbitMQRetryDelay = 5 * time.Second // 重试间隔时间
)

// Function to connect to RabbitMQ with retries
func connectRabbitMQ() error {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		log.Println("RABBITMQ_URL not set, skipping RabbitMQ connection.")
		return nil // Allow service to run without RabbitMQ for local dev if needed
	}

	var err error
	var attempt int // 重试次数计数器

	for attempt = 1; attempt <= maxRabbitMQRetries; attempt++ {
		log.Printf("尝试连接 RabbitMQ (第 %d 次)... URL: %s", attempt, rabbitURL)
		rabbitConn, err = amqp091.Dial(rabbitURL)
		if err == nil {
			// 连接成功，跳出循环
			log.Println("成功连接到 RabbitMQ。")
			break
		}

		log.Printf("连接 RabbitMQ 失败 (第 %d 次): %v", attempt, err)
		if attempt == maxRabbitMQRetries {
			// 达到最大重试次数，返回最后一次错误
			log.Printf("达到最大重试次数 (%d)，放弃连接 RabbitMQ。", maxRabbitMQRetries)
			return fmt.Errorf("failed to connect to RabbitMQ after %d attempts: %w", maxRabbitMQRetries, err)
		}

		log.Printf("等待 %v 后重试...", rabbitMQRetryDelay)
		time.Sleep(rabbitMQRetryDelay) // 等待一段时间后重试
	}

	// --- 连接成功后，继续打开通道和声明队列 ---

	// 如果连接失败（虽然理论上前面的 break 和 return 会处理，但作为保险）
	if rabbitConn == nil {
		return fmt.Errorf("RabbitMQ connection is nil after retry loop")
	}

	rabbitChannel, err = rabbitConn.Channel()
	if err != nil {
		rabbitConn.Close() // Close connection if channel fails
		return fmt.Errorf("failed to open a channel: %w", err)
	}

	// Declare the queue (idempotent)
	_, err = rabbitChannel.QueueDeclare(
		userRegisteredQueue, // name
		true,                // durable (queue survives broker restart)
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		rabbitChannel.Close()
		rabbitConn.Close()
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	log.Println("成功打开 RabbitMQ 通道并声明队列")
	return nil
}

// Register 实现用户注册方法
func (s *server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	log.Printf("Received Register request for user: %s, email: %s", req.GetUsername(), req.GetEmail())

	username := req.GetUsername()
	password := req.GetPassword()
	email := req.GetEmail() // Get email from request

	// 检查用户名、密码和邮箱是否为空
	if username == "" || password == "" || email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "用户名、密码和邮箱不能为空")
	}

	// 检查用户名是否已存在
	var existingUser models.User
	err := database.DB.Where("username = ?", username).First(&existingUser).Error
	if err == nil {
		log.Printf("用户名已存在: %s", username)
		return nil, status.Errorf(codes.AlreadyExists, "用户名已存在")
	} else if err != gorm.ErrRecordNotFound {
		log.Printf("检查用户名时数据库错误: %v", err)
		return nil, status.Errorf(codes.Internal, "检查用户名时出错")
	}

	// --- 注释掉检查邮箱是否已存在的代码块 ---
	/*
		// 检查邮箱是否已存在
		err = database.DB.Where("email = ?", email).First(&existingUser).Error
		if err == nil {
			log.Printf("邮箱已存在: %s", email)
			return nil, status.Errorf(codes.AlreadyExists, "邮箱已被注册")
		} else if err != gorm.ErrRecordNotFound {
			log.Printf("检查邮箱时数据库错误: %v", err)
			return nil, status.Errorf(codes.Internal, "检查邮箱时出错")
		}
	*/
	// --- 结束注释 ---

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("密码加密失败: %v", err)
		return nil, status.Errorf(codes.Internal, "密码加密失败")
	}

	newUser := models.User{
		Username: username,
		Password: string(hashedPassword),
		Email:    email, // Save email
	}

	// 创建用户
	result := database.DB.Create(&newUser)
	if result.Error != nil {
		log.Printf("创建用户失败: %v", result.Error)
		return nil, status.Errorf(codes.Internal, "创建用户失败")
	}

	log.Printf("用户注册成功: ID=%d, Username=%s, Email=%s", newUser.ID, newUser.Username, newUser.Email)

	// --- Publish UserRegistered event to RabbitMQ --- (Best effort)
	if rabbitChannel != nil { // 现在这个检查更有意义，因为我们会重试连接
		messageBody := map[string]interface{}{
			"user_id":  newUser.ID,
			"username": newUser.Username,
			"email":    newUser.Email,
		}
		body, err := json.Marshal(messageBody)
		if err != nil {
			log.Printf("无法序列化注册事件消息: %v", err)
			// Don't block registration if marshalling fails
		} else {
			publishCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Timeout for publishing
			defer cancel()

			err = rabbitChannel.PublishWithContext(publishCtx,
				"",                  // exchange (empty string for default exchange)
				userRegisteredQueue, // routing key (queue name when using default exchange)
				false,               // mandatory
				false,               // immediate
				amqp091.Publishing{
					ContentType:  "application/json",
					Body:         body,
					DeliveryMode: amqp091.Persistent, // Make message persistent
				},
			)
			if err != nil {
				log.Printf("发布 UserRegistered 事件失败: %v", err)
				// Log error but don't fail the registration
			} else {
				log.Printf("UserRegistered 事件已发布到队列 %s", userRegisteredQueue)
			}
		}
	} else {
		log.Println("RabbitMQ channel 未初始化或连接失败，跳过事件发布。") // 更新日志信息
	}

	return &pb.RegisterResponse{UserId: uint32(newUser.ID)}, nil
}

// Login 实现用户登录方法
func (s *server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Printf("Received Login request for user: %s", req.GetUsername())

	username := req.GetUsername()
	password := req.GetPassword()

	if username == "" || password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "用户名和密码不能为空")
	}

	// 查找用户
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("用户未找到: %s", username)
		} else {
			log.Printf("查找用户失败: %v", err)
		}
		return nil, status.Errorf(codes.Unauthenticated, "用户名或密码错误")
	}

	// 验证密码
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Printf("密码验证失败 for user %s: %v", username, err)
		return nil, status.Errorf(codes.Unauthenticated, "用户名或密码错误")
	}

	// 生成JWT令牌
	expirationTime := time.Now().Add(24 * time.Hour) // 24小时过期
	claims := &jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Use the global jwtKey
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Printf("JWT令牌生成失败: %v", err)
		return nil, status.Errorf(codes.Internal, "生成令牌失败")
	}

	log.Printf("用户登录成功: %s", username)
	return &pb.LoginResponse{Token: tokenString}, nil
}

// ChangePassword 实现修改密码方法
func (s *server) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	log.Printf("Received ChangePassword request for user ID: %d", req.GetUserId())

	userID := req.GetUserId()
	oldPassword := req.GetOldPassword()
	newPassword := req.GetNewPassword()

	// 基本验证
	if userID == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "无效的用户 ID")
	}
	if oldPassword == "" || newPassword == "" {
		return nil, status.Errorf(codes.InvalidArgument, "旧密码和新密码不能为空")
	}
	if oldPassword == newPassword {
		return nil, status.Errorf(codes.InvalidArgument, "新旧密码不能相同")
	}

	// 获取用户信息
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("修改密码时用户未找到: ID=%d", userID)
			return nil, status.Errorf(codes.NotFound, "用户不存在")
		}
		log.Printf("获取用户信息失败: %v", err)
		return nil, status.Errorf(codes.Internal, "获取用户信息失败")
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		log.Printf("旧密码验证失败 for user ID %d: %v", userID, err)
		// 注意：出于安全考虑，不要明确提示是"旧密码错误"，统一返回认证失败
		return nil, status.Errorf(codes.Unauthenticated, "旧密码不正确")
	}

	// 加密新密码
	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("新密码加密失败: %v", err)
		return nil, status.Errorf(codes.Internal, "新密码加密失败")
	}

	// 更新密码
	result := database.DB.Model(&user).Update("password", string(hashedNewPassword))
	if result.Error != nil {
		log.Printf("更新密码失败: %v", result.Error)
		return nil, status.Errorf(codes.Internal, "更新密码失败")
	}
	if result.RowsAffected == 0 {
		log.Printf("更新密码时未找到用户 (可能并发删除?): ID=%d", userID)
		return nil, status.Errorf(codes.NotFound, "用户不存在或更新失败")
	}

	log.Printf("用户 ID %d 密码修改成功", userID)
	// 返回空的响应表示成功
	return &pb.ChangePasswordResponse{}, nil
}

func main() {
	// ---- 1. 初始化数据库连接 ----
	if err := database.InitDB(); err != nil {
		log.Fatalf("无法初始化数据库: %v", err)
	}

	// ---- 2. 自动迁移数据库表结构 ----
	err := database.DB.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// ---- 3. 获取 JWT 密钥 ----
	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET_KEY 环境变量未设置!")
	}
	jwtKey = []byte(jwtSecret) // Assign to global variable

	// ---- 3.5 初始化 RabbitMQ 连接 (带重试) ----
	if err := connectRabbitMQ(); err != nil {
		// Log the error but allow the service to potentially continue without MQ
		log.Printf("无法连接到 RabbitMQ (经过重试): %v. 事件发布功能将不可用。", err)
		// 注意：这里不再调用 defer 关闭，因为连接/通道可能未成功打开
	} else {
		// Defer closing connection and channel only if successfully connected
		defer func() {
			if rabbitChannel != nil {
				rabbitChannel.Close()
			}
			if rabbitConn != nil {
				rabbitConn.Close()
			}
			log.Println("RabbitMQ 连接和通道已关闭")
		}()
	}

	// ---- 4. 设置 gRPC 服务器 ----
	port := ":50051" // gRPC 服务端口
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	// Pass the server struct instance, jwtKey is now global
	pb.RegisterUserServiceServer(s, &server{})

	reflection.Register(s)

	log.Printf("User service listening on %s (with reflection)", port)

	// ---- 5. 启动 gRPC 服务器 ----
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
