package mail

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"todo-project/email-service/internal/config"
)

// Sender 邮件发送器
type Sender struct {
	smtpAuth   smtp.Auth
	smtpServer string
	smtpSender string
}

// NewSender 创建新的邮件发送器
func NewSender(cfg *config.Config) *Sender {
	return &Sender{
		smtpAuth:   smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost),
		smtpServer: fmt.Sprintf("%s:%s", cfg.SMTPHost, cfg.SMTPPort),
		smtpSender: cfg.SMTPSender,
	}
}

// SendWelcomeEmail 发送欢迎邮件
func (s *Sender) SendWelcomeEmail(username string, email string, userID uint) error {
	subject := "欢迎加入我们的Todo应用！"
	body := fmt.Sprintf("你好 %s,\n\n欢迎加入！我们很高兴有你。\n\n你的用户ID是: %d\n\n谢谢,\nTodo团队", username, userID)
	to := []string{email} // 收件人列表

	// 根据RFC 822标准构造邮件消息
	emailMsg := []byte(fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", strings.Join(to, ","), s.smtpSender, subject, body))

	log.Printf("尝试发送欢迎邮件到 %s", email)
	// 发送邮件
	err := smtp.SendMail(s.smtpServer, s.smtpAuth, s.smtpSender, to, emailMsg)
	if err != nil {
		// 检查是否是特定的"short response"错误
		if strings.Contains(err.Error(), "short response") {
			// 对于"short response"错误，打印警告
			errorString := err.Error()
			log.Printf("警告：发送邮件到 %s 时SMTP服务器返回'short response'，邮件可能已发送。", email)
			log.Printf("  - 错误字符串: %s", errorString)
			log.Printf("  - 错误Go语法表示: %#v", err)
			log.Printf("  - 错误字符串字节 (Hex): %x", []byte(errorString))
			return nil // 视为成功
		}
		// 对于其他错误，返回错误
		return fmt.Errorf("发送邮件到 %s 失败: %w", email, err)
	}

	log.Printf("成功发送欢迎邮件到 %s", email)
	return nil
}
