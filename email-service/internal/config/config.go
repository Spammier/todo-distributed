package config

import "os"

// Config 包含应用程序配置
type Config struct {
	AppEnv       string
	RabbitMQURL  string
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPSender   string
}

// LoadConfig 从环境变量加载配置
func LoadConfig() *Config {
	return &Config{
		AppEnv:       getEnvOrDefault("APP_ENV", "development"),
		RabbitMQURL:  getEnvOrDefault("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
		SMTPHost:     getEnvOrDefault("SMTP_HOST", ""),
		SMTPPort:     getEnvOrDefault("SMTP_PORT", "587"),
		SMTPUser:     getEnvOrDefault("SMTP_USER", ""),
		SMTPPassword: getEnvOrDefault("SMTP_PASSWORD", ""),
		SMTPSender:   getEnvOrDefault("SMTP_SENDER", ""),
	}
}

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
