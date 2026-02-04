// 重置管理员密码脚本
// 用法: go run cmd/reset-password/main.go [新密码]
// 或在 Docker 中: docker exec chatbox-backend ./reset-password [新密码]
package main

import (
	"fmt"
	"log"
	"os"

	"chatbox-backend/config"
	"chatbox-backend/database"

	"golang.org/x/crypto/bcrypt"
)

const defaultPassword = "admin123"

func main() {
	// 获取新密码
	newPassword := defaultPassword
	if len(os.Args) > 1 {
		newPassword = os.Args[1]
	}

	if len(newPassword) < 6 {
		log.Fatal("Password must be at least 6 characters")
	}

	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	if err := database.Init(database.DBConfig{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
	}); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// 生成密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// 重置密码
	_, err = database.DB.Exec(
		"UPDATE users SET password_hash = ?, password_changed = 0 WHERE username = 'admin'",
		string(hashedPassword),
	)
	if err != nil {
		log.Fatalf("Failed to reset password: %v", err)
	}

	fmt.Println("Admin password has been reset successfully!")
	fmt.Printf("New password: %s\n", newPassword)
	fmt.Println("Please login and change your password immediately.")
}
