package config

import "os"

// Config 包含应用程序配置
type Config struct {
	Port            string
	JWTSecretKey    []byte
	UserServiceAddr string
	TodoServiceAddr string
}

// LoadConfig 从环境变量加载配置
func LoadConfig() *Config {
	return &Config{
		Port:            getEnvOrDefault("PORT", "8080"),
		JWTSecretKey:    []byte(getEnvOrDefault("JWT_SECRET_KEY", "")),
		UserServiceAddr: getEnvOrDefault("USER_SERVICE_ADDR", "user-service:50051"),
		TodoServiceAddr: getEnvOrDefault("TODO_SERVICE_ADDR", "todo-service:50052"),
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
