package models

import "time"

type Donor struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(150);not null" json:"name"`
	Email       string    `gorm:"type:varchar(150)" json:"email"` // Opsional jika Hamba Allah via offline
	Phone       string    `gorm:"type:varchar(20)" json:"phone"`
	IsAnonymous bool      `gorm:"default:false" json:"is_anonymous"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}