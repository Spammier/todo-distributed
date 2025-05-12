package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"todo-project/api-gateway/internal/models"
	todopb "todo-project/api-gateway/proto/todo"

	"github.com/gin-gonic/gin"
)

// CreateTodoHandler 处理创建待办事项请求
func CreateTodoHandler(todoClient todopb.TodoServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var reqBody struct {
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
			HandleGrpcError(c, err, "创建待办事项失败")
			return
		}
		c.JSON(http.StatusCreated, models.ConvertProtoTodoToResponse(res))
	}
}

// GetTodosHandler 处理获取所有待办事项请求
func GetTodosHandler(todoClient todopb.TodoServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		req := &todopb.GetTodosRequest{UserId: userID.(uint32)}
		res, err := todoClient.GetTodos(c.Request.Context(), req)

		if err != nil {
			HandleGrpcError(c, err, "获取待办事项失败")
			return
		}

		responseList := make([]models.TodoResponse, len(res.Todos))
		for i, protoTodo := range res.Todos {
			responseList[i] = models.ConvertProtoTodoToResponse(protoTodo)
		}
		c.JSON(http.StatusOK, responseList)
	}
}

// GetTodoByIDHandler 处理获取单个待办事项请求
func GetTodoByIDHandler(todoClient todopb.TodoServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		todoIDStr := c.Param("id")
		todoID, err := strconv.ParseUint(todoIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的待办事项ID"})
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
			HandleGrpcError(c, err, "获取待办事项失败")
			return
		}
		c.JSON(http.StatusOK, models.ConvertProtoTodoToResponse(res))
	}
}

// UpdateTodoHandler 处理更新待办事项请求
func UpdateTodoHandler(todoClient todopb.TodoServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		todoIDStr := c.Param("id")
		todoID, err := strconv.ParseUint(todoIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的待办事项ID"})
			return
		}

		var reqBody struct {
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
			HandleGrpcError(c, err, "更新待办事项失败")
			return
		}
		c.JSON(http.StatusOK, models.ConvertProtoTodoToResponse(res))
	}
}

// DeleteTodoHandler 处理删除待办事项请求
func DeleteTodoHandler(todoClient todopb.TodoServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		todoIDStr := c.Param("id")
		todoID, err := strconv.ParseUint(todoIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的待办事项ID"})
			return
		}

		grpcReq := &todopb.DeleteTodoRequest{
			UserId: userID.(uint32),
			TodoId: uint32(todoID),
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		_, err = todoClient.DeleteTodo(ctx, grpcReq)
		if err != nil {
			HandleGrpcError(c, err, "删除待办事项失败")
			return
		}
		c.Status(http.StatusNoContent)
	}
}

// BatchUpdateTodosHandler 处理批量更新待办事项请求
func BatchUpdateTodosHandler(todoClient todopb.TodoServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var reqBody struct {
			TodoIDs []uint32 `json:"todo_ids" binding:"required"`
			Action  string   `json:"action" binding:"required,oneof=MARK_AS_COMPLETED MARK_AS_INCOMPLETE"`
		}
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据: " + err.Error()})
			return
		}

		userID, _ := c.Get("user_id")

		var actionEnum todopb.BatchUpdateTodosRequest_ActionType
		switch reqBody.Action {
		case "MARK_AS_COMPLETED":
			actionEnum = todopb.BatchUpdateTodosRequest_MARK_AS_COMPLETED
		case "MARK_AS_INCOMPLETE":
			actionEnum = todopb.BatchUpdateTodosRequest_MARK_AS_INCOMPLETE
		default:
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
			HandleGrpcError(c, err, "批量更新待办事项失败")
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "批量更新成功"})
	}
}
