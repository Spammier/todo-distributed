package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"todo-project/todo-service/internal/model"
	"todo-project/todo-service/internal/util"
	pb "todo-project/todo-service/proto/todo"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

// 定义缓存持续时间
const cacheDuration = 10 * time.Minute

// server 实现了 pb.TodoServiceServer 接口
type server struct {
	db  *gorm.DB
	rdb *redis.Client
	pb.UnimplementedTodoServiceServer
}

// NewTodoService 创建一个新的 TodoService
func NewTodoService(db *gorm.DB, rdb *redis.Client) pb.TodoServiceServer {
	return &server{db: db, rdb: rdb}
}

// 实现 gRPC 方法
func (s *server) CreateTodo(ctx context.Context, req *pb.CreateTodoRequest) (*pb.Todo, error) {
	log.Printf("Received CreateTodo request for user_id: %d, title: %s", req.GetUserId(), req.GetTitle())

	if req.GetUserId() == 0 || req.GetTitle() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "用户 ID 和标题不能为空")
	}

	newTodo := model.Todo{
		UserID:      uint(req.GetUserId()),
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		Completed:   false,
	}

	result := s.db.Create(&newTodo)
	if result.Error != nil {
		log.Printf("创建 Todo 失败: %v", result.Error)
		return nil, status.Errorf(codes.Internal, "创建 Todo 失败")
	}

	log.Printf("Todo 创建成功: ID=%d", newTodo.ID)
	userCacheKey := fmt.Sprintf("user_todos:%d", newTodo.UserID)
	err := s.rdb.Del(ctx, userCacheKey).Err()
	if err != nil {
		log.Printf("警告: 清除用户 %d 的 Todos 列表缓存 (%s) 失败: %v", newTodo.UserID, userCacheKey, err)
	} else {
		log.Printf("Redis 用户 Todos 列表缓存已清除: %s", userCacheKey)
	}

	return util.ConvertToProtoTodo(&newTodo), nil
}

func (s *server) GetTodos(ctx context.Context, req *pb.GetTodosRequest) (*pb.GetTodosResponse, error) {
	log.Printf("Received GetTodos request for user_id: %d", req.GetUserId())
	userID := req.GetUserId()
	if userID == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "无效的用户 ID")
	}

	cacheKey := fmt.Sprintf("user_todos:%d", userID)

	// 尝试从 Redis 读取缓存
	cachedTodosJSON, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil { // 缓存命中
		var cachedTodos []*pb.Todo
		if unmarshalErr := json.Unmarshal([]byte(cachedTodosJSON), &cachedTodos); unmarshalErr == nil {
			log.Printf("从 Redis 缓存获取用户 %d 的 Todos 成功 (%d 条)", userID, len(cachedTodos))
			return &pb.GetTodosResponse{Todos: cachedTodos}, nil
		} else { // 将日志记录移到此 else 块中
			// 反序列化失败，记录日志并继续从数据库读取
			log.Printf("警告: 反序列化用户 %d 的 Todos 缓存失败: %v。将从数据库获取。", userID, unmarshalErr)
		}
	} else if err != redis.Nil { // Redis 出错 (非 key 不存在)
		log.Printf("警告: 从 Redis 获取用户 %d 的 Todos 缓存失败: %v。将从数据库获取。", userID, err)
	} else { // 缓存未命中 (err == redis.Nil)
		log.Printf("用户 %d 的 Todos 缓存未命中: %s。将从数据库获取。", userID, cacheKey)
	}

	var todos []*model.Todo
	result := s.db.Where("user_id = ?", userID).Find(&todos)
	if result.Error != nil {
		log.Printf("获取 Todos 失败 for user %d: %v", userID, result.Error)
		return nil, status.Errorf(codes.Internal, "获取待办事项失败")
	}

	// 将从数据库获取的数据转换为 Protobuf 格式
	protoTodos := util.ConvertToProtoTodos(todos)

	// 尝试将结果写入 Redis 缓存
	todosJSON, errMarshal := json.Marshal(protoTodos)
	if errMarshal == nil {
		errSet := s.rdb.Set(ctx, cacheKey, todosJSON, cacheDuration).Err()
		if errSet != nil {
			log.Printf("警告: 写入用户 %d 的 Todos (%d 条) 到 Redis 缓存 (%s) 失败: %v", userID, len(protoTodos), cacheKey, errSet)
		} else {
			log.Printf("用户 %d 的 Todos (%d 条) 已写入 Redis 缓存: %s", userID, len(protoTodos), cacheKey)
		}
	} else {
		log.Printf("警告: 序列化用户 %d 的 Todos 以进行缓存失败: %v", userID, errMarshal)
	}

	log.Printf("找到 %d 个 Todos for user %d (从数据库)", len(todos), userID)
	return &pb.GetTodosResponse{Todos: protoTodos}, nil
}

func (s *server) GetTodoByID(ctx context.Context, req *pb.GetTodoByIDRequest) (*pb.Todo, error) {
	log.Printf("Received GetTodoByID request for user_id: %d, todo_id: %d", req.GetUserId(), req.GetTodoId())
	userID := req.GetUserId()
	todoID := req.GetTodoId()

	if userID == 0 || todoID == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "无效的用户 ID 或 Todo ID")
	}

	cacheKey := fmt.Sprintf("todo:%d", todoID)

	// 尝试从 Redis 读取缓存
	cachedTodoJSON, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil { // 缓存命中
		var cachedTodo pb.Todo
		if unmarshalErr := json.Unmarshal([]byte(cachedTodoJSON), &cachedTodo); unmarshalErr == nil {
			// 关键: 验证缓存的 Todo 是否属于请求的用户
			if cachedTodo.GetUserId() == userID {
				log.Printf("从 Redis 缓存获取 Todo %d (用户 %d) 成功", todoID, userID)
				return &cachedTodo, nil
			}
			log.Printf("警告: 缓存的 Todo %d (属于用户 %d) 与请求用户 %d 不匹配。将从数据库获取。", todoID, cachedTodo.GetUserId(), userID)
			// 用户不匹配，视为缓存未命中，将从数据库重新验证和获取
		} else {
			log.Printf("警告: 反序列化 Todo %d 缓存失败: %v。将从数据库获取。", todoID, unmarshalErr)
		}
	} else if err != redis.Nil { // Redis 出错 (非 key 不存在)
		log.Printf("警告: 从 Redis 获取 Todo %d 缓存失败: %v。将从数据库获取。", todoID, err)
	} else { // 缓存未命中 (err == redis.Nil)
		log.Printf("Todo %d 缓存未命中: %s。将从数据库获取。", todoID, cacheKey)
	}

	var todo model.Todo
	dbErr := s.db.Where("id = ? AND user_id = ?", todoID, userID).First(&todo).Error
	if dbErr != nil {
		if dbErr == gorm.ErrRecordNotFound {
			log.Printf("Todo 未找到: user_id=%d, todo_id=%d", userID, todoID)
			return nil, status.Errorf(codes.NotFound, "待办事项未找到或无权访问")
		}
		log.Printf("获取 Todo %d 失败 for user %d: %v", todoID, userID, dbErr)
		return nil, status.Errorf(codes.Internal, "获取待办事项失败")
	}

	// 将从数据库获取的数据转换为 Protobuf 格式
	protoTodo := util.ConvertToProtoTodo(&todo)

	// 尝试将结果写入 Redis 缓存
	todoJSON, errMarshal := json.Marshal(protoTodo)
	if errMarshal == nil {
		errSet := s.rdb.Set(ctx, cacheKey, todoJSON, cacheDuration).Err()
		if errSet != nil {
			log.Printf("警告: 写入 Todo %d 到 Redis 缓存 (%s) 失败: %v", todoID, cacheKey, errSet)
		} else {
			log.Printf("Todo %d 已写入 Redis 缓存: %s", todoID, cacheKey)
		}
	} else {
		log.Printf("警告: 序列化 Todo %d 以进行缓存失败: %v", todoID, errMarshal)
	}

	log.Printf("找到 Todo: ID=%d (从数据库)", todo.ID)
	return protoTodo, nil
}

func (s *server) UpdateTodo(ctx context.Context, req *pb.UpdateTodoRequest) (*pb.Todo, error) {
	log.Printf("Received UpdateTodo request for user_id: %d, todo_id: %d", req.GetUserId(), req.GetTodoId())
	userID := req.GetUserId()
	todoID := req.GetTodoId()

	if userID == 0 || todoID == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "无效的用户 ID 或 Todo ID")
	}

	// 查找原始 Todo 以进行权限检查
	var originalTodo model.Todo
	err := s.db.Where("id = ? AND user_id = ?", todoID, userID).First(&originalTodo).Error
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
	result := s.db.Model(&originalTodo).Updates(updates)
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
	// 清除相关 Redis 缓存 (用户列表和单个 Todo 缓存)
	userCacheKey := fmt.Sprintf("user_todos:%d", req.GetUserId())
	todoCacheKey := fmt.Sprintf("todo:%d", todoID)

	errDelTodo := s.rdb.Del(ctx, todoCacheKey).Err()
	if errDelTodo != nil {
		log.Printf("警告: 清除单个 Todo %d 的缓存 (%s) 失败: %v", todoID, todoCacheKey, errDelTodo)
	} else {
		log.Printf("Redis 单个 Todo 缓存已清除: %s", todoCacheKey)
	}

	errDelUserTodos := s.rdb.Del(ctx, userCacheKey).Err()
	if errDelUserTodos != nil {
		log.Printf("警告: 清除用户 %d 的 Todos 列表缓存 (%s) 失败: %v", req.GetUserId(), userCacheKey, errDelUserTodos)
	} else {
		log.Printf("Redis 用户 Todos 列表缓存已清除: %s", userCacheKey)
	}

	// 返回更新后的 Todo (从数据库重新获取以确保数据最新)
	var updatedTodo model.Todo
	s.db.First(&updatedTodo, todoID)

	return util.ConvertToProtoTodo(&updatedTodo), nil
}

func (s *server) DeleteTodo(ctx context.Context, req *pb.DeleteTodoRequest) (*emptypb.Empty, error) {
	log.Printf("Received DeleteTodo request for user_id: %d, todo_id: %d", req.GetUserId(), req.GetTodoId())
	userID := req.GetUserId()
	todoID := req.GetTodoId()

	if userID == 0 || todoID == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "无效的用户 ID 或 Todo ID")
	}

	// 直接尝试删除，GORM 的 Delete 会返回影响的行数
	result := s.db.Where("id = ? AND user_id = ?", todoID, userID).Delete(&model.Todo{})

	if result.Error != nil {
		log.Printf("删除 Todo %d 失败: %v", todoID, result.Error)
		return nil, status.Errorf(codes.Internal, "删除待办事项失败")
	}

	if result.RowsAffected == 0 {
		log.Printf("删除时 Todo 未找到: user_id=%d, todo_id=%d", userID, todoID)
		return nil, status.Errorf(codes.NotFound, "待办事项未找到或无权删除")
	}

	log.Printf("Todo %d 删除成功", todoID)
	// 清除相关 Redis 缓存
	userCacheKey := fmt.Sprintf("user_todos:%d", userID)
	todoCacheKey := fmt.Sprintf("todo:%d", todoID)

	errDelTodo := s.rdb.Del(ctx, todoCacheKey).Err()
	if errDelTodo != nil {
		log.Printf("警告: 清除单个 Todo %d 的缓存 (%s) 失败: %v", todoID, todoCacheKey, errDelTodo)
	} else {
		log.Printf("Redis 单个 Todo 缓存已清除: %s", todoCacheKey)
	}

	errDelUserTodos := s.rdb.Del(ctx, userCacheKey).Err()
	if errDelUserTodos != nil {
		log.Printf("警告: 清除用户 %d 的 Todos 列表缓存 (%s) 失败: %v", userID, userCacheKey, errDelUserTodos)
	} else {
		log.Printf("Redis 用户 Todos 列表缓存已清除: %s", userCacheKey)
	}

	return &emptypb.Empty{}, nil
}

func (s *server) BatchUpdateTodos(ctx context.Context, req *pb.BatchUpdateTodosRequest) (*emptypb.Empty, error) {
	userID := req.GetUserId()
	todoIDs := req.GetTodoIds() // 这是 []uint32
	action := req.GetAction()

	if userID == 0 || len(todoIDs) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "用户 ID 和待办事项 ID 列表不能为空")
	}
	if action == pb.BatchUpdateTodosRequest_ACTION_TYPE_UNSPECIFIED {
		return nil, status.Errorf(codes.InvalidArgument, "未指定有效的操作类型")
	}

	log.Printf("Received BatchUpdateTodos request for user_id: %d, todo_ids: %v, action: %s", userID, todoIDs, action.String())

	var operationLog model.BatchOperationLog
	operationLog.UserID = uint(userID)
	operationLog.OperationType = action.String()
	operationLog.Status = "FAILURE" // 默认失败，成功时更新

	// 将 []uint32 转换为 JSON 字符串以便存储
	affectedTodoIDsBytes, err := json.Marshal(todoIDs)
	if err != nil {
		log.Printf("序列化 affected_todo_ids 失败: %v", err)
		// 即使序列化失败，也尝试记录日志，但 affected_todo_ids 可能为空
		operationLog.Details = fmt.Sprintf("序列化 affected_todo_ids 失败: %v", err)
		s.db.Create(&operationLog) // 尽力记录
		return nil, status.Errorf(codes.Internal, "处理请求失败: 序列化ID失败")
	}
	operationLog.AffectedTodoIDs = string(affectedTodoIDsBytes)

	dbErr := s.db.Transaction(func(tx *gorm.DB) error {
		var updates map[string]interface{}
		var operationDetail string
		var actualAffectedRows int64

		switch action {
		case pb.BatchUpdateTodosRequest_MARK_AS_COMPLETED:
			updates = map[string]interface{}{"completed": true}
			operationDetail = "批量标记完成"
		case pb.BatchUpdateTodosRequest_MARK_AS_INCOMPLETE:
			updates = map[string]interface{}{"completed": false}
			operationDetail = "批量标记未完成"
		default:
			return status.Errorf(codes.InvalidArgument, "无效的批量操作类型: %s", action.String())
		}

		// 执行批量更新
		// GORM 的 Where 子句接受 []uint32，不需要手动转换
		result := tx.Model(&model.Todo{}).Where("id IN ? AND user_id = ?", todoIDs, userID).Updates(updates)
		if result.Error != nil {
			log.Printf("事务内：批量更新 Todos 失败 for user %d: %v", userID, result.Error)
			operationLog.Details = fmt.Sprintf("数据库更新失败: %v", result.Error)
			return result.Error // 这会回滚事务
		}
		actualAffectedRows = result.RowsAffected

		log.Printf("事务内：用户 %d 的批量操作 '%s' 影响了 %d 行 (请求 %d 个 IDs)", userID, operationDetail, actualAffectedRows, len(todoIDs))

		// 根据实际影响行数和请求数量来确定最终状态和日志详情
		if actualAffectedRows == int64(len(todoIDs)) {
			operationLog.Status = "SUCCESS"
			operationLog.Details = fmt.Sprintf("%s成功，影响 %d 条记录。", operationDetail, actualAffectedRows)
		} else if actualAffectedRows > 0 && actualAffectedRows < int64(len(todoIDs)) {
			operationLog.Status = "PARTIAL_FAILURE"
			operationLog.Details = fmt.Sprintf("%s部分成功，请求 %d 条，实际影响 %d 条。可能部分待办事项不属于该用户或不存在。", operationDetail, len(todoIDs), actualAffectedRows)
			// 根据业务需求，部分成功是否算作整体事务失败并回滚
		} else if actualAffectedRows == 0 && len(todoIDs) > 0 {
			operationLog.Status = "FAILURE" // 或者 "NO_MATCHING_RECORDS"
			operationLog.Details = fmt.Sprintf("%s失败，请求 %d 条，没有记录被影响。可能所有待办事项均不属于该用户或不存在。", operationDetail, len(todoIDs))
			// 如果要求至少有一条被更新，这里也可以 return 一个 error
		}

		// 创建批量操作日志条目
		if err := tx.Create(&operationLog).Error; err != nil {
			log.Printf("事务内：创建批量操作日志失败: %v", err)
			// 这个错误也应该回滚整个事务
			return err
		}

		log.Printf("事务内：批量操作日志已记录: UserID=%d, Type=%s, Status=%s", operationLog.UserID, operationLog.OperationType, operationLog.Status)
		return nil // 事务成功提交
	})

	if dbErr != nil {
		log.Printf("批量操作事务处理失败 for user %d, action %s: %v", userID, action.String(), dbErr)
		// 如果事务是因为我们自己返回的 status.Error 而失败，尝试将其传递出去
		if s, ok := status.FromError(dbErr); ok {
			return nil, s.Err()
		}
		// 对于其他 GORM 错误，统一返回 Internal 错误
		return nil, status.Errorf(codes.Internal, "批量操作处理失败: %v", dbErr)
	}

	// 事务成功，现在清理 Redis 缓存
	log.Printf("批量操作事务成功 for user %d. 清理相关缓存...", userID)
	userCacheKey := fmt.Sprintf("user_todos:%d", userID)
	if rdbErr := s.rdb.Del(ctx, userCacheKey).Err(); rdbErr != nil {
		log.Printf("警告: 批量操作后清除用户 %d 列表缓存 (%s) 失败: %v", userID, userCacheKey, rdbErr)
	} else {
		log.Printf("批量操作后 Redis 用户 %d 列表缓存已清除: %s", userID, userCacheKey)
	}

	for _, singleTodoID := range todoIDs {
		todoCacheKey := fmt.Sprintf("todo:%d", singleTodoID)
		if rdbErr := s.rdb.Del(ctx, todoCacheKey).Err(); rdbErr != nil {
			log.Printf("警告: 批量操作后清除 Todo %d 缓存 (%s) 失败: %v", singleTodoID, todoCacheKey, rdbErr)
		} else {
			log.Printf("批量操作后 Redis Todo %d 缓存已清除: %s", singleTodoID, todoCacheKey)
		}
	}

	log.Printf("用户 %d 的批量操作 (%s) 成功完成并已清理缓存。日志状态: %s", userID, action.String(), operationLog.Status)
	return &emptypb.Empty{}, nil
}
