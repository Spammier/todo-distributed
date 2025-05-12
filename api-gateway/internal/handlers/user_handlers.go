package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	userpb "todo-project/api-gateway/proto/user"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// RegisterHandler 处理用户注册请求
func RegisterHandler(userClient userpb.UserServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req userpb.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据: " + err.Error()})
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		res, err := userClient.Register(ctx, &req)
		if err != nil {
			HandleGrpcError(c, err, "注册失败")
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "注册成功", "user_id": res.GetUserId()})
	}
}

// LoginHandler 处理用户登录请求
func LoginHandler(userClient userpb.UserServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req userpb.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据: " + err.Error()})
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		res, err := userClient.Login(ctx, &req)
		if err != nil {
			HandleGrpcError(c, err, "登录失败")
			return
		}

		tokenString := res.GetToken()

		// 解析Token获取username
		var username string
		token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
		if err == nil {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if name, ok := claims["username"].(string); ok {
					username = name
					log.Printf("从登录token中解析到用户名: %s", username)
				} else {
					log.Println("警告：登录token中username claim类型不正确或不存在")
				}
			} else {
				log.Println("警告：无法断言登录token的claims为MapClaims")
			}
		} else {
			log.Printf("警告：解析登录token (unverified) 失败: %v", err)
		}

		// 返回token和username
		c.JSON(http.StatusOK, gin.H{
			"token":    tokenString,
			"username": username,
		})
	}
}

// ChangePasswordHandler 处理密码修改请求
func ChangePasswordHandler(userClient userpb.UserServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var reqBody struct {
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据: " + err.Error()})
			return
		}

		// 从上下文中获取user_id
		userID, _ := c.Get("user_id")

		// 准备gRPC请求
		grpcReq := &userpb.ChangePasswordRequest{
			UserId:      userID.(uint32),
			OldPassword: reqBody.OldPassword,
			NewPassword: reqBody.NewPassword,
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		_, err := userClient.ChangePassword(ctx, grpcReq)
		if err != nil {
			HandleGrpcError(c, err, "修改密码失败")
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
	}
}
