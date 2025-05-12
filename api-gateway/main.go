package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4" // 导入 JWT 包
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes" // 导入 gRPC codes
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status" // 导入 gRPC status

	// 导入 proto 定义
	todopb "todo-project/api-gateway/proto/todo"
	userpb "todo-project/api-gateway/proto/user"
)

var jwtKey []byte // JWT 密钥

// --- 新增：定义用于 API 响应的 Todo 结构体 ---
type TodoResponse struct {
	Id          uint32 `json:"id"`
	UserId      uint32 `json:"user_id"` // 注意：实际 API 可能不需要返回 user_id
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	CreatedAt   string `json:"created_at"` // 使用 string 类型
	UpdatedAt   string `json:"updated_at"` // 使用 string 类型
}

// --- 新增：辅助函数，将 todopb.Todo 转换为 TodoResponse ---
func convertProtoTodoToResponse(protoTodo *todopb.Todo) TodoResponse {
	createdAt := "" // 默认为空字符串
	if protoTodo.CreatedAt != nil && protoTodo.CreatedAt.IsValid() {
		// 转换为 UTC 时间并格式化为 RFC3339Nano (近似 ISO 8601)
		createdAt = protoTodo.CreatedAt.AsTime().UTC().Format(time.RFC3339Nano)
	}
	updatedAt := ""
	if protoTodo.UpdatedAt != nil && protoTodo.UpdatedAt.IsValid() {
		updatedAt = protoTodo.UpdatedAt.AsTime().UTC().Format(time.RFC3339Nano)
	}
	return TodoResponse{
		Id:          protoTodo.Id,
		UserId:      protoTodo.UserId,
		Title:       protoTodo.Title,
		Description: protoTodo.Description,
		Completed:   protoTodo.Completed,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

// AuthMiddleware 实现 JWT 认证
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证令牌"})
			c.Abort()
			return
		}

		// 检查是否为 Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证令牌格式错误"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 解析 Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 确保签名方法是 HS256
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("非预期的签名方法: %v", token.Header["alg"])
			}
			return jwtKey, nil
		})

		if err != nil {
			log.Printf("Token 解析错误: %v", err)
			errorMsg := "无效的认证令牌"
			if errors.Is(err, jwt.ErrTokenExpired) {
				errorMsg = "认证令牌已过期"
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": errorMsg})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// 从 claims 中提取 user_id (注意类型断言可能需要调整)
			userIDFloat, okUserID := claims["user_id"].(float64) // JWT 标准库解析数字为 float64
			username, _ := claims["username"].(string)           // 忽略 username 的 ok 检查
			if !okUserID {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌声明 (缺少 user_id)"})
				c.Abort()
				return
			}
			// 将 user_id 存储到 Gin 上下文中，供后续处理器使用
			c.Set("user_id", uint32(userIDFloat)) // gRPC 需要 uint32
			c.Set("username", username)           // 顺便存储 username
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌"})
			c.Abort()
		}
	}
}

func main() {
	// ---- 0. 获取 JWT 密钥 ----
	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET_KEY 环境变量未设置!")
	}
	jwtKey = []byte(jwtSecret)

	// ---- 1. 获取后端服务地址 ----
	userServiceAddr := getEnvOrDefault("USER_SERVICE_ADDR", "user-service:50051") // Docker Compose 环境下使用服务名
	todoServiceAddr := getEnvOrDefault("TODO_SERVICE_ADDR", "todo-service:50052") // Docker Compose 环境下使用服务名

	// ---- 2. 建立 gRPC 连接 ----
	// 连接 User Service
	userConn, err := grpc.NewClient(userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("无法连接到 User Service: %v", err)
	}
	defer userConn.Close()
	userClient := userpb.NewUserServiceClient(userConn)
	log.Printf("已连接到 User Service at %s", userServiceAddr)

	// 连接 Todo Service
	todoConn, err := grpc.NewClient(todoServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("无法连接到 Todo Service: %v", err)
	}
	defer todoConn.Close()
	todoClient := todopb.NewTodoServiceClient(todoConn)
	log.Printf("已连接到 Todo Service at %s", todoServiceAddr)

	// ---- 3. 设置 Gin HTTP 服务器 ----
	router := gin.Default()
	// (可以添加 CORS 中间件)

	// ---- 4. 定义 API 路由组 ----
	api := router.Group("/api")
	{
		// --- 公开路由 ---
		api.POST("/register", func(c *gin.Context) {
			var req userpb.RegisterRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据: " + err.Error()})
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*15) // 增加超时到15秒
			defer cancel()
			res, err := userClient.Register(ctx, &req)
			if err != nil {
				handleGrpcError(c, err, "注册失败")
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "注册成功", "user_id": res.GetUserId()})
		})

		api.POST("/login", func(c *gin.Context) {
			var req userpb.LoginRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据: " + err.Error()})
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			res, err := userClient.Login(ctx, &req)
			if err != nil {
				handleGrpcError(c, err, "登录失败") // 登录失败也可能返回特定状态码
				return
			}

			tokenString := res.GetToken()

			// --- 新增：解析 Token 获取 username ---
			var username string
			// 注意：这里我们不验证签名，因为 token 是刚由 user-service 生成的，我们信任它。
			// 如果需要验证，应该使用 jwt.Parse 并提供密钥。
			token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
			if err == nil {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					if name, ok := claims["username"].(string); ok {
						username = name
						log.Printf("从登录 token 中解析到用户名: %s", username)
					} else {
						log.Println("警告：登录 token 中 username claim 类型不正确或不存在")
					}
				} else {
					log.Println("警告：无法断言登录 token 的 claims 为 MapClaims")
				}
			} else {
				log.Printf("警告：解析登录 token (unverified) 失败: %v", err)
				// 即使解析失败，仍然可以继续返回 token，只是 username 为空
			}
			// --- 结束新增 ---

			// 返回 token 和 username
			c.JSON(http.StatusOK, gin.H{
				"token":    tokenString,
				"username": username, // 添加 username 字段
			})
		})

		// --- 需要认证的路由组 ---
		auth := api.Group("")
		auth.Use(AuthMiddleware()) // 应用认证中间件
		{
			// --- 用户相关认证路由 ---
			auth.POST("/change-password", func(c *gin.Context) {
				var reqBody struct { // 定义临时的请求体结构
					OldPassword string `json:"old_password"`
					NewPassword string `json:"new_password"`
				}
				if err := c.ShouldBindJSON(&reqBody); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据: " + err.Error()})
					return
				}

				// 从上下文中获取 user_id
				userID, _ := c.Get("user_id")

				// 准备 gRPC 请求
				grpcReq := &userpb.ChangePasswordRequest{
					UserId:      userID.(uint32), // 类型断言
					OldPassword: reqBody.OldPassword,
					NewPassword: reqBody.NewPassword,
				}

				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10) // 增加超时到10秒
				defer cancel()

				_, err := userClient.ChangePassword(ctx, grpcReq)
				if err != nil {
					handleGrpcError(c, err, "修改密码失败")
					return
				}

				c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
			})

			// --- Todo 相关认证路由 ---
			todos := auth.Group("/todos")
			{
				// 创建 Todo
				todos.POST("", func(c *gin.Context) {
					var reqBody struct { // 请求体可以只包含需要用户输入的部分
						Title       string `json:"title"`
						Description string `json:"description"`
					}
					if err := c.ShouldBindJSON(&reqBody); err != nil {
						c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据: " + err.Error()})
						return
					}

					userID, _ := c.Get("user_id")

					grpcReq := &todopb.CreateTodoRequest{
						UserId:      userID.(uint32),
						Title:       reqBody.Title,
						Description: reqBody.Description,
					}

					ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
					defer cancel()

					res, err := todoClient.CreateTodo(ctx, grpcReq)
					if err != nil {
						handleGrpcError(c, err, "创建待办事项失败")
						return
					}
					// 返回转换后的结构体
					c.JSON(http.StatusCreated, convertProtoTodoToResponse(res))
				})

				// 获取所有 Todos
				todos.GET("", func(c *gin.Context) {
					userID, _ := c.Get("user_id")

					// 调用 Todo Service 获取该用户的所有 Todos
					req := &todopb.GetTodosRequest{UserId: userID.(uint32)} // 使用 GetTodosRequest 和 uint32
					// 使用 GetTodos 方法，并声明 getErr
					res, getErr := todoClient.GetTodos(c.Request.Context(), req)

					if getErr != nil {
						handleGrpcError(c, getErr, "获取待办事项失败") // 使用新的错误变量
						return
					}
					// 转换列表中的每个 Todo
					responseList := make([]TodoResponse, len(res.Todos))
					for i, protoTodo := range res.Todos {
						responseList[i] = convertProtoTodoToResponse(protoTodo)
					}
					c.JSON(http.StatusOK, responseList) // 返回转换后的列表
				})

				// 获取单个 Todo
				todos.GET("/:id", func(c *gin.Context) {
					userID, _ := c.Get("user_id")
					todoIDStr := c.Param("id")
					todoID, err := strconv.ParseUint(todoIDStr, 10, 32)
					if err != nil {
						c.JSON(http.StatusBadRequest, gin.H{"error": "无效的待办事项 ID"})
						return
					}

					grpcReq := &todopb.GetTodoByIDRequest{
						UserId: userID.(uint32),
						TodoId: uint32(todoID),
					}

					ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
					defer cancel()

					res, err := todoClient.GetTodoByID(ctx, grpcReq)
					if err != nil {
						handleGrpcError(c, err, "获取待办事项失败")
						return
					}
					// 返回转换后的结构体
					c.JSON(http.StatusOK, convertProtoTodoToResponse(res))
				})

				// 更新 Todo
				todos.PUT("/:id", func(c *gin.Context) {
					userID, _ := c.Get("user_id")
					todoIDStr := c.Param("id")
					todoID, err := strconv.ParseUint(todoIDStr, 10, 32)
					if err != nil {
						c.JSON(http.StatusBadRequest, gin.H{"error": "无效的待办事项 ID"})
						return
					}

					var reqBody struct { // 请求体包含可更新字段
						Title       string `json:"title"`
						Description string `json:"description"`
						Completed   bool   `json:"completed"`
					}
					if err := c.ShouldBindJSON(&reqBody); err != nil {
						c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据: " + err.Error()})
						return
					}

					grpcReq := &todopb.UpdateTodoRequest{
						UserId:      userID.(uint32),
						TodoId:      uint32(todoID),
						Title:       reqBody.Title,
						Description: reqBody.Description,
						Completed:   reqBody.Completed,
					}

					ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
					defer cancel()

					res, err := todoClient.UpdateTodo(ctx, grpcReq)
					if err != nil {
						handleGrpcError(c, err, "更新待办事项失败")
						return
					}
					// 返回转换后的结构体
					c.JSON(http.StatusOK, convertProtoTodoToResponse(res))
				})

				// 删除 Todo
				todos.DELETE("/:id", func(c *gin.Context) {
					userID, _ := c.Get("user_id")
					todoIDStr := c.Param("id")
					todoID, err := strconv.ParseUint(todoIDStr, 10, 32)
					if err != nil {
						c.JSON(http.StatusBadRequest, gin.H{"error": "无效的待办事项 ID"})
						return
					}

					grpcReq := &todopb.DeleteTodoRequest{
						UserId: userID.(uint32),
						TodoId: uint32(todoID),
					}

					ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
					defer cancel()

					// 使用新的 deleteErr 变量
					_, deleteErr := todoClient.DeleteTodo(ctx, grpcReq)
					if deleteErr != nil {
						handleGrpcError(c, deleteErr, "删除待办事项失败") // 使用新的错误变量
						return
					}
					c.Status(http.StatusNoContent) // 删除成功返回 204
				})

				// 批量更新 Todos 状态
				todos.PATCH("/batch", func(c *gin.Context) {
					var reqBody struct {
						TodoIDs []uint32 `json:"todo_ids" binding:"required"`
						Action  string   `json:"action" binding:"required,oneof=MARK_AS_COMPLETED MARK_AS_INCOMPLETE"`
					}
					if err := c.ShouldBindJSON(&reqBody); err != nil {
						c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据: " + err.Error()})
						return
					}

					userID, _ := c.Get("user_id")

					// 将字符串 action 转换为枚举值
					var actionEnum todopb.BatchUpdateTodosRequest_ActionType
					switch reqBody.Action {
					case "MARK_AS_COMPLETED":
						actionEnum = todopb.BatchUpdateTodosRequest_MARK_AS_COMPLETED
					case "MARK_AS_INCOMPLETE":
						actionEnum = todopb.BatchUpdateTodosRequest_MARK_AS_INCOMPLETE
					default:
						// 理论上不会到这里，因为绑定验证已经检查了
						c.JSON(http.StatusBadRequest, gin.H{"error": "无效的操作类型"})
						return
					}

					grpcReq := &todopb.BatchUpdateTodosRequest{
						UserId:  userID.(uint32),
						TodoIds: reqBody.TodoIDs,
						Action:  actionEnum,
					}

					ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
					defer cancel()

					_, err := todoClient.BatchUpdateTodos(ctx, grpcReq)
					if err != nil {
						handleGrpcError(c, err, "批量更新待办事项失败")
						return
					}

					c.JSON(http.StatusOK, gin.H{"message": "批量更新成功"})
				})
			}
		}
	}

	// ---- 5. 启动 HTTP 服务器 ----
	port := getEnvOrDefault("PORT", "8080")
	log.Printf("API Gateway listening on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to run API Gateway: %v", err)
	}
}

// getEnvOrDefault 获取环境变量
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// handleGrpcError 将 gRPC 错误转换为 Gin 的 HTTP 响应
func handleGrpcError(c *gin.Context, err error, defaultMessage string) {
	log.Printf("gRPC Error: %v", err)
	st, ok := status.FromError(err)
	if ok {
		// 根据 gRPC 状态码映射到 HTTP 状态码
		httpCode := http.StatusInternalServerError // 默认为 500
		switch st.Code() {
		case codes.InvalidArgument:
			httpCode = http.StatusBadRequest
		case codes.Unauthenticated:
			httpCode = http.StatusUnauthorized
		case codes.PermissionDenied:
			httpCode = http.StatusForbidden
		case codes.NotFound:
			httpCode = http.StatusNotFound
		case codes.AlreadyExists:
			httpCode = http.StatusConflict
		case codes.DeadlineExceeded:
			httpCode = http.StatusGatewayTimeout // 或者 500
		}
		c.JSON(httpCode, gin.H{"error": st.Message()}) // 使用 gRPC 的错误消息
	} else {
		// 如果不是标准的 gRPC 错误
		c.JSON(http.StatusInternalServerError, gin.H{"error": defaultMessage, "details": err.Error()})
	}
}
