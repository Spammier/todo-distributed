package mq

import (
	"encoding/json"
	"log"
	"time"

	"todo-project/email-service/internal/config"
	"todo-project/email-service/internal/mail"

	"github.com/rabbitmq/amqp091-go"
)

// Consumer 消息消费者
type Consumer struct {
	conn       *amqp091.Connection
	channel    *amqp091.Channel
	mailSender *mail.Sender
}

// NewConsumer 创建新的消息消费者
func NewConsumer(cfg *config.Config, mailSender *mail.Sender) (*Consumer, error) {
	var err error
	var conn *amqp091.Connection
	var attempt int
	maxRetries := 10
	retryDelay := 5 * time.Second

	// 尝试连接RabbitMQ，带重试
	for attempt = 1; attempt <= maxRetries; attempt++ {
		log.Printf("尝试连接RabbitMQ (第%d次)...", attempt)
		conn, err = amqp091.Dial(cfg.RabbitMQURL)
		if err == nil {
			break
		}
		log.Printf("连接RabbitMQ失败: %v", err)
		if attempt == maxRetries {
			return nil, err
		}
		time.Sleep(retryDelay)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// 声明队列
	_, err = channel.QueueDeclare(
		UserRegisteredQueue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	// 设置QoS
	err = channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	return &Consumer{
		conn:       conn,
		channel:    channel,
		mailSender: mailSender,
	}, nil
}

// Start 开始消费消息
func (c *Consumer) Start() (<-chan bool, error) {
	msgs, err := c.channel.Consume(
		UserRegisteredQueue,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return nil, err
	}

	done := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("收到消息: %s", d.Body)
			var msg UserRegisteredMessage
			// 解析JSON消息体
			err := json.Unmarshal(d.Body, &msg)
			if err != nil {
				log.Printf("解析消息错误: %s。可能发送到死信队列或Nack处理。", err)
				// 处理错误消息: 发送Nack且不重新入队
				d.Nack(false, false)
				continue
			}

			// 发送欢迎邮件
			err = c.mailSender.SendWelcomeEmail(msg.Username, msg.Email, msg.UserID)
			if err != nil {
				log.Printf("发送邮件失败: %v。将消息Nack (不重新入队)。", err)
				d.Nack(false, false)
			} else {
				// 确认消息已成功处理
				d.Ack(false)
			}
		}
	}()

	return done, nil
}

// Close 关闭连接
func (c *Consumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
