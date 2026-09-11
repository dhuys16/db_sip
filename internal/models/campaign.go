package models

import "time"
type Campaign struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ProgramID     uint      `gorm:"not null" json:"program_id"`
	Program       Program   `gorm:"foreignKey:ProgramID" json:"program"`
	Title         string    `gorm:"type:varchar(255);not null" json:"title"`
	Slug          string    `gorm:"type:varchar(255);unique;not null" json:"slug"`
	Description   string    `gorm:"type:text;not null" json:"description"`
	TargetAmount  float64   `gorm:"type:decimal(15,2);default:0" json:"target_amount"`
	CurrentAmount float64   `gorm:"type:decimal(15,2);default:0" json:"current_amount"`
	Division      string    `gorm:"type:enum('lazsip', 'sarsip');not null" json:"division"`
	Status        string    `gorm:"type:enum('active', 'completed', 'cancelled');default:'active'" json:"status"`
	Image         string    `gorm:"column:image;type:varchar(255)" json:"image"`
	IsPinned      bool      `gorm:"default:false" json:"is_pinned"`
	KodeUnik      int       `gorm:"type:int;default:0" json:"kode_unik"` // Untuk transfer bank manual jika diperlukan
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}