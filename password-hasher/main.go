package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 1. 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("用法: go run main.go <你的密码>")
		fmt.Println("或者编译后运行: ./password-hasher <你的密码>")
		os.Exit(1) // 参数不足，退出
	}

	// 2. 获取密码
	password := os.Args[1]

	// 3. 生成 bcrypt 哈希值
	// bcrypt.DefaultCost 是推荐的计算成本（当前为 10）
	// 成本越高，哈希越慢，但也越难破解
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("生成密码哈希失败: %v", err)
	}

	// 4. 打印哈希值
	fmt.Printf("密码: %s\n", password)
	fmt.Printf("Bcrypt 哈希: %s\n", string(hashedPassword))
}
