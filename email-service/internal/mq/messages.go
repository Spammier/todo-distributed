package mq

// UserRegisteredMessage 定义了预期的消息体结构
type UserRegisteredMessage struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// 队列名称常量
const (
	UserRegisteredQueue = "user_registered_queue" // 应与user-service中的队列名称匹配
)
