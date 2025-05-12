package database

import (
	"fmt"
	"log"

	"todo-project/user-service/internal/config"
	"todo-project/user-service/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DB 全局数据库连接
var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB(cfg *config.Config) error {
	var err error

	log.Printf("数据库连接配置: User=%s, Host=%s, Port=%s, DB=%s",
		cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	// 自动迁移数据库表结构
	err = DB.AutoMigrate(&models.User{})
	if err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	log.Println("成功连接到数据库")
	return nil
}
