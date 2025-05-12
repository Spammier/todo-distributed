package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"todo-project/todo-service/internal/config"
	"todo-project/todo-service/internal/model"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	ctx = context.Background()
)

func InitDB(cfg *config.Config) *gorm.DB {
	// 检查当前环境
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}
	log.Printf("当前运行环境: %s", appEnv)

	// 根据环境选择不同的配置
	if appEnv == "production" || appEnv == "container" {
		// 生产环境 - 使用环境变量覆盖配置文件中的设置
		log.Printf("生产/容器环境 - 原始配置: DBUser=%s, DBHost=%s, DBPort=%s, DBName=%s",
			cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)

		// 从环境变量获取配置，如果环境变量不存在则保留原值
		if dbUser := os.Getenv("DB_USER"); dbUser != "" {
			cfg.DBUser = dbUser
		}
		if dbPass := os.Getenv("DB_PASSWORD"); dbPass != "" {
			cfg.DBPass = dbPass
		}
		if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
			cfg.DBHost = dbHost
		} else {
			// 容器环境默认使用服务名作为主机名
			cfg.DBHost = "mysql"
		}
		if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
			cfg.DBPort = dbPort
		}
		if dbName := os.Getenv("DB_NAME"); dbName != "" {
			cfg.DBName = dbName
		}

		log.Printf("生产/容器环境 - 最终配置: DBUser=%s, DBHost=%s, DBPort=%s, DBName=%s",
			cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)
	} else {
		// 开发/测试环境 - 使用本地配置
		log.Printf("开发/测试环境 - 修改为本地配置，原始配置: DBUser=%s, DBHost=%s, DBPort=%s, DBName=%s",
			cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)

		cfg.DBHost = "127.0.0.1" // 本地开发使用 localhost
		cfg.DBUser = "root"      // 本地开发使用 root
		cfg.DBPass = ""          // 本地测试密码

		log.Printf("开发/测试环境 - 修改后配置: DBUser=%s, DBHost=%s, DBPort=%s, DBName=%s",
			cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)

	// 打印DSN（注意不要打印密码）
	log.Printf("数据库连接DSN: %s:***@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("无法初始化数据库: %v", err)
	}
	log.Println("成功连接到数据库")

	if err := db.AutoMigrate(&model.Todo{}, &model.BatchOperationLog{}); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	return db
}

func InitRedis(cfg *config.Config) *redis.Client {
	// 检查当前环境
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	// 根据环境选择不同的配置
	if appEnv == "production" || appEnv == "container" {
		// 生产环境 - 使用环境变量覆盖配置
		log.Printf("生产/容器环境 - 原始 Redis 地址: %s", cfg.RedisAddr)

		// 从环境变量获取 Redis 配置
		if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
			cfg.RedisAddr = redisAddr
		} else {
			// 容器环境默认使用服务名
			cfg.RedisAddr = "redis:6379"
		}

		if redisPass := os.Getenv("REDIS_PASSWORD"); redisPass != "" {
			cfg.RedisPass = redisPass
		}

		if redisDB := os.Getenv("REDIS_DB"); redisDB != "" {
			if db, err := strconv.Atoi(redisDB); err == nil {
				cfg.RedisDB = db
			}
		}

		log.Printf("生产/容器环境 - 最终 Redis 地址: %s", cfg.RedisAddr)
	} else {
		// 开发/测试环境 - 使用本地 Redis
		log.Printf("开发/测试环境 - 修改 Redis 配置，原始地址: %s", cfg.RedisAddr)
		cfg.RedisAddr = "127.0.0.1:6379" // 本地开发使用 localhost
		log.Printf("开发/测试环境 - 修改后 Redis 地址: %s", cfg.RedisAddr)
	}

	redisDBNum := cfg.RedisDB
	log.Printf("Redis配置: Addr=%s, DB=%d", cfg.RedisAddr, redisDBNum)

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       redisDBNum,
	})

	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("警告: Redis PING 命令失败 (Addr: %s, Password Used: %t, DB: %d): %v",
			cfg.RedisAddr, cfg.RedisPass != "", redisDBNum, err)
	} else {
		log.Printf("Redis PING 响应: %s", pong)
		log.Printf("成功连接到 Redis (Addr: %s, DB: %d)", cfg.RedisAddr, redisDBNum)
	}

	return rdb
}
