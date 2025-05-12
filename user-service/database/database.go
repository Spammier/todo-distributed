package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DB 全局数据库连接
var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() error {
	var err error

	// 检查当前环境
	appEnv := getEnvOrDefault("APP_ENV", "development")
	log.Printf("当前运行环境: %s", appEnv)

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
		dbPass = "" // 与 todo-service 相同的密码
		dbHost = "127.0.0.1"
		dbPort = "3306"
		dbName = getEnvOrDefault("DB_NAME", "todo")
	}

	log.Printf("数据库连接配置: User=%s, Host=%s, Port=%s, DB=%s",
		dbUser, dbHost, dbPort, dbName)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	log.Println("成功连接到数据库")
	return nil
}

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
