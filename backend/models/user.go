package models

import (
	"chatbox-backend/database"
	"time"
)

type User struct {
	ID              int64     `json:"id"`
	Username        string    `json:"username"`
	PasswordHash    string    `json:"-"` // 不在 JSON 中暴露
	Role            string    `json:"role"`
	PasswordChanged bool      `json:"password_changed"`
	CreatedAt       time.Time `json:"created_at"`
}

type UserResponse struct {
	ID              int64     `json:"id"`
	Username        string    `json:"username"`
	Role            string    `json:"role"`
	PasswordChanged bool      `json:"password_changed"`
	CreatedAt       time.Time `json:"created_at"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:              u.ID,
		Username:        u.Username,
		Role:            u.Role,
		PasswordChanged: u.PasswordChanged,
		CreatedAt:       u.CreatedAt,
	}
}

// GetUserByID 根据 ID 获取用户
func GetUserByID(id int64) (*User, error) {
	user := &User{}
	var passwordChanged int
	err := database.DB.QueryRow(
		"SELECT id, username, password_hash, role, password_changed, created_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &passwordChanged, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	user.PasswordChanged = passwordChanged == 1
	return user, nil
}

// GetUserByUsername 根据用户名获取用户
func GetUserByUsername(username string) (*User, error) {
	user := &User{}
	var passwordChanged int
	err := database.DB.QueryRow(
		"SELECT id, username, password_hash, role, password_changed, created_at FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &passwordChanged, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	user.PasswordChanged = passwordChanged == 1
	return user, nil
}

// GetUserCount 获取用户总数
func GetUserCount() (int, error) {
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

// GetAllUsers 获取所有用户
func GetAllUsers() ([]User, error) {
	rows, err := database.DB.Query(
		"SELECT id, username, password_hash, role, password_changed, created_at FROM users ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		var passwordChanged int
		if err := rows.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &passwordChanged, &user.CreatedAt); err != nil {
			return nil, err
		}
		user.PasswordChanged = passwordChanged == 1
		users = append(users, user)
	}

	return users, rows.Err()
}

// UpdatePassword 更新用户密码
func UpdatePassword(userID int64, newPasswordHash string) error {
	_, err := database.DB.Exec(
		"UPDATE users SET password_hash = ?, password_changed = 1 WHERE id = ?",
		newPasswordHash, userID,
	)
	return err
}

// ResetAdminPassword 重置管理员密码（用于脚本）
func ResetAdminPassword(newPasswordHash string) error {
	_, err := database.DB.Exec(
		"UPDATE users SET password_hash = ?, password_changed = 0 WHERE username = 'admin'",
		newPasswordHash,
	)
	return err
}

// IsAdmin 检查用户是否为管理员
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}
