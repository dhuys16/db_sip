package models

import (
	"time"
	"gorm.io/gorm"
)

type Admin struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"type:varchar(100);not null" json:"name"`
	Email        string         `gorm:"type:varchar(100);unique;not null" json:"email"`
	PasswordHash string         `gorm:"not null" json:"-"` // Hidden from JSON
	Role         string         `gorm:"type:enum('superadmin', 'admin');default:'admin'" json:"role"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}