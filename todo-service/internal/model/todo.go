package model

import "time"

type Todo struct {
	ID          uint   `gorm:"primaryKey"`
	UserID      uint   `gorm:"not null;index"`
	Title       string `gorm:"not null"`
	Description string
	Completed   bool `gorm:"default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type BatchOperationLog struct {
	ID              uint   `gorm:"primaryKey"`
	UserID          uint   `gorm:"not null;index"`
	OperationType   string `gorm:"size:50;not null"`
	AffectedTodoIDs string `gorm:"type:json;not null"`
	Status          string `gorm:"size:20;not null"`
	Details         string `gorm:"type:text"`
	CreatedAt       time.Time
}
