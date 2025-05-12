package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"todo-project/user-service/internal/config"
	"todo-project/user-service/internal/models"

	"github.com/rabbitmq/amqp091-go"
)

// 全局变量
var (
	RabbitConn    *amqp091.Connection
	RabbitChannel *amqp091.Channel
)

// 队列名称常量
const (
	UserRegisteredQueue = "user_registered_queue"
	MaxRabbitMQRetries  = 5
	RabbitMQRetryDelay  = 5 * time.Second
)

// ConnectRabbitMQ 连接到RabbitMQ
func ConnectRabbitMQ(cfg *config.Config) error {
	rabbitURL := cfg.RabbitMQURL
	if rabbitURL == "" {
		log.Println("RABBITMQ_URL not set, skipping RabbitMQ connection.")
		return nil // 允许服务在没有RabbitMQ的情况下运行
	}

	var err error
	var attempt int

	for attempt = 1; attempt <= MaxRabbitMQRetries; attempt++ {
		log.Printf("尝试连接RabbitMQ (第%d次)... URL: %s", attempt, rabbitURL)
		RabbitConn, err = amqp091.Dial(rabbitURL)
		if err == nil {
			log.Println("成功连接到RabbitMQ。")
			break
		}

		log.Printf("连接RabbitMQ失败 (第%d次): %v", attempt, err)
		if attempt == MaxRabbitMQRetries {
			log.Printf("达到最大重试次数 (%d)，放弃连接RabbitMQ。", MaxRabbitMQRetries)
			return fmt.Errorf("failed to connect to RabbitMQ after %d attempts: %w", MaxRabbitMQRetries, err)
		}

		log.Printf("等待%v后重试...", RabbitMQRetryDelay)
		time.Sleep(RabbitMQRetryDelay)
	}

	// 如果连接失败
	if RabbitConn == nil {
		return fmt.Errorf("RabbitMQ connection is nil after retry loop")
	}

	RabbitChannel, err = RabbitConn.Channel()
	if err != nil {
		RabbitConn.Close()
		return fmt.Errorf("failed to open a channel: %w", err)
	}

	// 声明队列
	_, err = RabbitChannel.QueueDeclare(
		UserRegisteredQueue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		RabbitChannel.Close()
		RabbitConn.Close()
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	log.Println("成功打开RabbitMQ通道并声明队列")
	return nil
}

// PublishUserRegistered 发布用户注册事件
func PublishUserRegistered(user *models.User) error {
	if RabbitChannel == nil {
		log.Println("RabbitMQ channel未初始化或连接失败，跳过事件发布。")
		return nil // 允许在没有RabbitMQ的情况下继续
	}

	messageBody := map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	}
	body, err := json.Marshal(messageBody)
	if err != nil {
		return fmt.Errorf("无法序列化注册事件消息: %w", err)
	}

	publishCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = RabbitChannel.PublishWithContext(publishCtx,
		"",                  // exchange
		UserRegisteredQueue, // routing key
		false,               // mandatory
		false,               // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp091.Persistent,
		},
	)
	if err != nil {
		return fmt.Errorf("发布UserRegistered事件失败: %w", err)
	}

	log.Printf("UserRegistered事件已发布到队列%s", UserRegisteredQueue)
	return nil
}

// CloseConnections 关闭RabbitMQ连接
func CloseConnections() {
	if RabbitChannel != nil {
		RabbitChannel.Close()
	}
	if RabbitConn != nil {
		RabbitConn.Close()
	}
	log.Println("RabbitMQ连接和通道已关闭")
}
