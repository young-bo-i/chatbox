package database

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

const (
	DefaultAdminUsername = "admin"
	DefaultAdminPassword = "admin123"
)

// SeedDefaultAdmin 创建默认管理员账户（如果不存在），或修复密码哈希
func SeedDefaultAdmin() error {
	// 检查是否已存在管理员
	var count int
	var passwordHash string
	err := DB.QueryRow("SELECT COUNT(*), COALESCE(MAX(password_hash), '') FROM users WHERE username = ?", DefaultAdminUsername).Scan(&count, &passwordHash)
	if err != nil {
		return err
	}

	if count > 0 {
		// 验证密码是否正确
		if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(DefaultAdminPassword)) == nil {
			log.Printf("Admin user '%s' already exists with correct password", DefaultAdminUsername)
			return nil
		}
		// 密码哈希不正确（可能是迁移脚本的硬编码哈希），需要修复
		log.Printf("Admin user '%s' exists but password hash is invalid, resetting password", DefaultAdminUsername)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(DefaultAdminPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		_, err = DB.Exec("UPDATE users SET password_hash = ?, password_changed = 0 WHERE username = ?", string(hashedPassword), DefaultAdminUsername)
		if err != nil {
			return err
		}
		log.Printf("Admin password reset to '%s'", DefaultAdminPassword)
		return nil
	}

	// 生成密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(DefaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 创建管理员账户
	_, err = DB.Exec(`
		INSERT INTO users (username, password_hash, role, password_changed)
		VALUES (?, ?, 'admin', 0)
	`, DefaultAdminUsername, string(hashedPassword))
	if err != nil {
		return err
	}

	log.Printf("Default admin user '%s' created with password '%s'", DefaultAdminUsername, DefaultAdminPassword)
	return nil
}
