package models

import "time"

type PaymentFee struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	PaymentMethod string    `gorm:"type:varchar(100);not null;unique" json:"payment_method"` // qris, bank_transfer, gopay
	FeeFlat       float64   `gorm:"type:decimal(10,2);default:0" json:"fee_flat"`
	FeePercentage float64   `gorm:"type:decimal(5,2);default:0" json:"fee_percentage"`
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}