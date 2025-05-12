package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"todo-project/email-service/internal/config"
	"todo-project/email-service/internal/mail"
	"todo-project/email-service/internal/mq"
)

func main() {
	log.Println("邮件服务启动中...")

	// 加载配置
	cfg := config.LoadConfig()

	// 检查SMTP配置
	if cfg.SMTPHost == "" || cfg.SMTPUser == "" || cfg.SMTPPassword == "" || cfg.SMTPSender == "" {
		log.Fatalf("错误：一个或多个SMTP环境变量未设置 (SMTP_HOST, SMTP_USER, SMTP_PASSWORD, SMTP_SENDER)")
	}

	// 创建邮件发送器
	mailSender := mail.NewSender(cfg)

	// 创建消息消费者
	consumer, err := mq.NewConsumer(cfg, mailSender)
	if err != nil {
		log.Fatalf("创建消息消费者失败: %v", err)
	}
	defer consumer.Close()

	// 开始消费消息
	done, err := consumer.Start()
	if err != nil {
		log.Fatalf("启动消息消费失败: %v", err)
	}

	log.Printf(" [*] 等待队列 '%s' 上的消息。按 CTRL+C 退出", mq.UserRegisteredQueue)

	// 优雅关闭处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigChan:
		log.Println("关闭邮件服务...")
	case <-done:
		log.Println("消息处理已完成。")
	}

	log.Println("邮件服务已停止。")
}
