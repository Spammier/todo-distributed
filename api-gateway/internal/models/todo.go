package models

import (
	"time"

	todopb "todo-project/api-gateway/proto/todo"
)

// TodoResponse 定义用于API响应的Todo结构体
type TodoResponse struct {
	Id          uint32 `json:"id"`
	UserId      uint32 `json:"user_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ConvertProtoTodoToResponse 将protobuf的Todo转换为TodoResponse
func ConvertProtoTodoToResponse(protoTodo *todopb.Todo) TodoResponse {
	createdAt := ""
	if protoTodo.CreatedAt != nil && protoTodo.CreatedAt.IsValid() {
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
