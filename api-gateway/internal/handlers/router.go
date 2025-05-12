package handlers

import (
	"todo-project/api-gateway/internal/middleware"
	todopb "todo-project/api-gateway/proto/todo"
	userpb "todo-project/api-gateway/proto/user"

	"github.com/gin-gonic/gin"
)

// SetupRouter 配置API路由
func SetupRouter(router *gin.Engine, userClient userpb.UserServiceClient, todoClient todopb.TodoServiceClient, jwtKey []byte) {
	// API路由组
	api := router.Group("/api")
	{
		// 公开路由
		api.POST("/register", RegisterHandler(userClient))
		api.POST("/login", LoginHandler(userClient))

		// 需要认证的路由组
		auth := api.Group("")
		auth.Use(middleware.AuthMiddleware(jwtKey))
		{
			// 用户相关认证路由
			auth.POST("/change-password", ChangePasswordHandler(userClient))

			// Todo相关认证路由
			todos := auth.Group("/todos")
			{
				todos.POST("", CreateTodoHandler(todoClient))
				todos.GET("", GetTodosHandler(todoClient))
				todos.GET("/:id", GetTodoByIDHandler(todoClient))
				todos.PUT("/:id", UpdateTodoHandler(todoClient))
				todos.DELETE("/:id", DeleteTodoHandler(todoClient))
				todos.PATCH("/batch", BatchUpdateTodosHandler(todoClient))
			}
		}
	}
}
