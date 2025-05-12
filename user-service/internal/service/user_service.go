package service

import (
	"context"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"todo-project/user-service/internal/database"
	"todo-project/user-service/internal/models"
	"todo-project/user-service/internal/mq"
	pb "todo-project/user-service/proto/user"
)

// UserService 实现UserServiceServer接口
type UserService struct {
	pb.UnimplementedUserServiceServer
	jwtKey []byte
}

// NewUserService 创建UserService实例
func NewUserService(jwtKey []byte) *UserService {
	return &UserService{
		jwtKey: jwtKey,
	}
}

// Register 实现用户注册方法
func (s *UserService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	log.Printf("Received Register request for user: %s, email: %s", req.GetUsername(), req.GetEmail())

	username := req.GetUsername()
	password := req.GetPassword()
	email := req.GetEmail()

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

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("密码加密失败: %v", err)
		return nil, status.Errorf(codes.Internal, "密码加密失败")
	}

	newUser := models.User{
		Username: username,
		Password: string(hashedPassword),
		Email:    email,
	}

	// 创建用户
	result := database.DB.Create(&newUser)
	if result.Error != nil {
		log.Printf("创建用户失败: %v", result.Error)
		return nil, status.Errorf(codes.Internal, "创建用户失败")
	}

	log.Printf("用户注册成功: ID=%d, Username=%s, Email=%s", newUser.ID, newUser.Username, newUser.Email)

	// 发布用户注册事件
	if err := mq.PublishUserRegistered(&newUser); err != nil {
		log.Printf("发布用户注册事件失败: %v", err)
		// 不阻止注册成功
	}

	return &pb.RegisterResponse{UserId: uint32(newUser.ID)}, nil
}

// Login 实现用户登录方法
func (s *UserService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
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
	tokenString, err := token.SignedString(s.jwtKey)
	if err != nil {
		log.Printf("JWT令牌生成失败: %v", err)
		return nil, status.Errorf(codes.Internal, "生成令牌失败")
	}

	log.Printf("用户登录成功: %s", username)
	return &pb.LoginResponse{Token: tokenString}, nil
}

// ChangePassword 实现修改密码方法
func (s *UserService) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	log.Printf("Received ChangePassword request for user ID: %d", req.GetUserId())

	userID := req.GetUserId()
	oldPassword := req.GetOldPassword()
	newPassword := req.GetNewPassword()

	// 基本验证
	if userID == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "无效的用户ID")
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

	log.Printf("用户ID %d 密码修改成功", userID)
	return &pb.ChangePasswordResponse{}, nil
}
