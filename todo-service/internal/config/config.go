package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUser    string
	DBPass    string
	DBHost    string
	DBPort    string
	DBName    string
	RedisAddr string
	RedisPass string
	RedisDB   int
	GRPCPort  string
}

func Load() *Config {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Printf("警告: 加载 ../.env 文件失败: %v. 将依赖已设置的环境变量。", err)
	}

	return &Config{
		DBUser:    getEnvOrDefault("DB_USER", "root"),
		DBPass:    os.Getenv("DB_PASSWORD"),
		DBHost:    getEnvOrDefault("DB_HOST", "localhost"),
		DBPort:    getEnvOrDefault("DB_PORT", "3306"),
		DBName:    getEnvOrDefault("DB_NAME", "todo"),
		RedisAddr: getEnvOrDefault("REDIS_ADDR", "localhost:6379"),
		RedisPass: getEnvOrDefault("REDIS_PASSWORD", ""),
		RedisDB:   getEnvOrDefaultInt("REDIS_DB", 0),
		GRPCPort:  ":50052",
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvOrDefaultInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
