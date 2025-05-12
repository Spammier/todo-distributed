package util

import (
	"todo-project/todo-service/internal/model"
	pb "todo-project/todo-service/proto/todo"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ConvertToProtoTodo(todoModel *model.Todo) *pb.Todo {
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

func ConvertToProtoTodos(todoModels []*model.Todo) []*pb.Todo {
	protoTodos := make([]*pb.Todo, len(todoModels))
	for i, model := range todoModels {
		protoTodos[i] = ConvertToProtoTodo(model)
	}
	return protoTodos
}
