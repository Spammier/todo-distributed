package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/rabbitmq/amqp091-go"
)

// --- 常量定义 ---
const (
	userRegisteredQueue = "user_registered_queue" // 应与 user-service 中的队列名称匹配
	// 环境变量键名
	envRabbitURL  = "RABBITMQ_URL"
	envSmtpHost   = "SMTP_HOST"
	envSmtpPort   = "SMTP_PORT"
	envSmtpUser   = "SMTP_USER"     // 用于登录的邮箱地址
	envSmtpPass   = "SMTP_PASSWORD" // 邮箱密码或应用专用密码
	envSmtpSender = "SMTP_SENDER"   // 显示在"发件人"字段中的邮箱地址
)

// --- 结构体定义 ---
// UserRegisteredMessage 定义了预期的消息体结构
type UserRegisteredMessage struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// --- 全局变量 ---
var (
	rabbitConn    *amqp091.Connection // RabbitMQ 连接
	rabbitChannel *amqp091.Channel    // RabbitMQ 通道
	smtpAuth      smtp.Auth           // SMTP 认证信息
	smtpServer    string              // SMTP 服务器地址和端口，例如 "smtp.example.com:587"
	smtpSender    string              // SMTP 发件人地址
)

// --- 辅助函数 ---
// failOnError 检查错误并在出错时终止程序
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

// --- 主要逻辑 ---
func main() {
	log.Println("邮件服务启动中...")

	// --- 加载配置 --- //
	rabbitURL := os.Getenv(envRabbitURL)
	smtpHost := os.Getenv(envSmtpHost)
	smtpPort := os.Getenv(envSmtpPort)
	smtpUser := os.Getenv(envSmtpUser)
	smtpPass := os.Getenv(envSmtpPass)
	smtpSender = os.Getenv(envSmtpSender) // 赋值给全局变量

	// 检查环境变量是否都已设置
	if rabbitURL == "" || smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" || smtpSender == "" {
		log.Fatalf("错误：一个或多个环境变量未设置 (%s, %s, %s, %s, %s, %s)",
			envRabbitURL, envSmtpHost, envSmtpPort, envSmtpUser, envSmtpPass, envSmtpSender)
	}

	smtpServer = fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	// 设置 SMTP 认证
	smtpAuth = smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	// --- 连接到 RabbitMQ --- //
	var err error
	rabbitConn, err = amqp091.Dial(rabbitURL)
	failOnError(err, "连接 RabbitMQ 失败")
	defer rabbitConn.Close() // 确保程序退出时关闭连接

	rabbitChannel, err = rabbitConn.Channel()
	failOnError(err, "打开通道失败")
	defer rabbitChannel.Close() // 确保程序退出时关闭通道

	// 声明队列 (确保队列存在且与 user-service 中的设置匹配)
	q, err := rabbitChannel.QueueDeclare(
		userRegisteredQueue,
		true,  // durable: 队列持久化，重启后依然存在
		false, // delete when unused: 队列在不被使用时不会自动删除
		false, // exclusive: 非独占队列，允许其他连接访问
		false, // no-wait: 服务器无需等待确认
		nil,   // arguments: 其他参数
	)
	failOnError(err, "声明队列失败")

	// 设置服务质量 (QoS): 一次只处理一条消息
	err = rabbitChannel.Qos(
		1,     // prefetch count: 每次从队列获取的消息数量
		0,     // prefetch size: 每次获取消息的总大小限制 (0表示不限制)
		false, // global: false 表示 QoS 设置仅应用于当前消费者
	)
	failOnError(err, "设置 QoS 失败")

	// --- 开始消费消息 --- //
	msgs, err := rabbitChannel.Consume(
		q.Name, // queue: 要消费的队列名称
		"",     // consumer: 消费者标签 (空字符串表示让 RabbitMQ 生成唯一标签)
		false,  // auto-ack: 关闭自动确认，改为手动确认
		false,  // exclusive: 非独占消费者
		false,  // no-local: 不接收自己发送的消息 (如果适用)
		false,  // no-wait: 服务器无需等待确认
		nil,    // args: 其他参数
	)
	failOnError(err, "注册消费者失败")

	log.Printf(" [*] 等待队列 '%s' 上的消息。按 CTRL+C 退出", q.Name)

	// --- 处理消息 --- //
	forever := make(chan bool) // 用于阻塞主 goroutine

	// 启动一个 goroutine 来处理接收到的消息
	go func() {
		for d := range msgs { // 循环处理消息通道中的数据
			log.Printf("收到消息: %s", d.Body)
			var msg UserRegisteredMessage
			// 解析 JSON 消息体
			err := json.Unmarshal(d.Body, &msg)
			if err != nil {
				log.Printf("解析消息错误: %s。可能发送到死信队列或 Nack 处理。", err)
				// 处理错误消息: 发送 Nack 且不重新入队
				d.Nack(false, false) // multiple=false, requeue=false
				continue             // 继续处理下一条消息
			}

			// --- 发送邮件 --- //
			subject := "欢迎加入我们的 Todo 应用！"
			body := fmt.Sprintf("你好 %s,\n\n欢迎加入！我们很高兴有你。\n\n你的用户 ID 是: %d\n\n谢谢,\nTodo 团队", msg.Username, msg.UserID)
			to := []string{msg.Email} // 收件人列表

			// 根据 RFC 822 标准构造邮件消息
			emailMsg := []byte(fmt.Sprintf("To: %s\r\n"+ // 收件人
				"From: %s\r\n"+ // 发件人 (使用配置的 sender 地址)
				"Subject: %s\r\n"+ // 主题
				"\r\n"+ // 空行分隔头部和正文
				"%s\r\n", strings.Join(to, ","), smtpSender, subject, body)) // 正文

			log.Printf("尝试发送欢迎邮件到 %s", msg.Email)
			// 发送邮件
			err = smtp.SendMail(smtpServer, smtpAuth, smtpSender, to, emailMsg)

			// --- 改进的错误处理逻辑 ---
			if err != nil {
				// 检查是否是特定的 "short response" 错误
				if strings.Contains(err.Error(), "short response") {
					// 对于 "short response" 错误，打印警告并 Ack
					errorString := err.Error() // 获取错误字符串
					log.Printf("警告：发送邮件到 %s 时 SMTP 服务器返回 'short response'，邮件可能已发送。", msg.Email)
					log.Printf("  - 错误字符串: %s", errorString)                 // 打印原始字符串
					log.Printf("  - 错误 Go 语法表示: %#v", err)                   // 打印 Go 语法表示
					log.Printf("  - 错误字符串字节 (Hex): %x", []byte(errorString)) // 打印字符串的十六进制字节
					d.Ack(false)                                             // 确认消息，防止重试
				} else {
					// 对于其他错误，打印失败日志并 Nack
					log.Printf("发送邮件到 %s 失败: %v。将消息 Nack (不重新入队)。", msg.Email, err)
					d.Nack(false, false) // multiple=false, requeue=false
				}
			} else {
				// 没有错误，发送成功
				log.Printf("成功发送欢迎邮件到 %s", msg.Email)
				// 确认消息已成功处理
				d.Ack(false) // multiple=false
			}
			// --- 结束改进的错误处理逻辑 ---
		}
	}()

	// --- 优雅关闭处理 --- //
	sigChan := make(chan os.Signal, 1) // 创建一个通道来接收信号
	// 监听 SIGINT (Ctrl+C) 和 SIGTERM 信号
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan // 阻塞，直到接收到上述信号之一

	log.Println("关闭邮件服务...")
	// 此处无需显式清理，因为 defer 语句会处理通道和连接的关闭

	log.Println("邮件服务已停止。")
	close(forever) // 向处理消息的 goroutine 发送停止信号 (尽管从关闭的 msgs 通道读取也会使其停止)
}
