package models

import (
	"time"
)

// User 表示一个用户
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"type:varchar(255);uniqueIndex;not null"`
	Email     string    `json:"email" gorm:"type:varchar(255);not null"`
	Password  string    `json:"-" gorm:"type:varchar(255);not null"` // 密码在 JSON 中省略
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// (移除了 LoginRequest 和 LoginResponse，因为它们由 proto 定义)
// (移除了 RegisterRequest，由 proto 定义)
