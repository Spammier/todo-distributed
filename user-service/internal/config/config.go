package config

import (
	"os"
)

// Config 包含应用程序配置
type Config struct {
	AppEnv       string
	DBUser       string
	DBPass       string
	DBHost       string
	DBPort       string
	DBName       string
	JWTSecretKey []byte
	RabbitMQURL  string
}

// LoadConfig 从环境变量加载配置
func LoadConfig() *Config {
	appEnv := getEnvOrDefault("APP_ENV", "development")

	// 根据环境选择不同的配置
	var dbUser, dbPass, dbHost, dbPort, dbName string

	if appEnv == "production" || appEnv == "container" {
		// 生产环境或容器环境 - 使用环境变量
		dbUser = getEnvOrDefault("DB_USER", "root")
		dbPass = os.Getenv("DB_PASSWORD")
		dbHost = getEnvOrDefault("DB_HOST", "mysql") // 容器默认使用服务名
		dbPort = getEnvOrDefault("DB_PORT", "3306")
		dbName = getEnvOrDefault("DB_NAME", "todo")
	} else {
		// 开发/测试环境 - 使用硬编码值或本地环境变量
		dbUser = "root"
		dbPass = "" // 开发环境默认空密码
		dbHost = "127.0.0.1"
		dbPort = "3306"
		dbName = getEnvOrDefault("DB_NAME", "todo")
	}

	return &Config{
		AppEnv:       appEnv,
		DBUser:       dbUser,
		DBPass:       dbPass,
		DBHost:       dbHost,
		DBPort:       dbPort,
		DBName:       dbName,
		JWTSecretKey: []byte(getEnvOrDefault("JWT_SECRET_KEY", "")),
		RabbitMQURL:  getEnvOrDefault("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
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
