package models

import "time"

type Beneficiary struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ProgramID      uint      `gorm:"not null" json:"program_id"`
	Program        Program   `gorm:"foreignKey:ProgramID" json:"program"`
	FullName       string    `gorm:"type:varchar(255);not null" json:"full_name"`
	Address        string    `gorm:"type:text" json:"address"`
	AmountReceived float64   `gorm:"type:decimal(15,2);default:0" json:"amount_received"`
	DateReceived   time.Time `json:"date_received"`
	PrivacyLevel   string    `gorm:"type:enum('public', 'initials', 'hidden');default:'hidden'" json:"privacy_level"`
	Story          string    `gorm:"type:text" json:"story"`
	ImageURL       string    `gorm:"type:varchar(255)" json:"image_url"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}