package models

import "time"

type Program struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(150);not null" json:"name"`
	Slug        string    `gorm:"type:varchar(150);unique;not null" json:"slug"`
	Description string    `gorm:"type:text" json:"description"`
	Image       string    `gorm:"column:image;type:varchar(255)" json:"image"`
	Division    string    `gorm:"type:enum('lazsip', 'sarsip');not null" json:"division"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}