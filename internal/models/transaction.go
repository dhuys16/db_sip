package models

import "time"

type Transaction struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	MidtransOrderID string     `gorm:"type:varchar(100);unique;not null" json:"midtrans_order_id"`
	CampaignID      *uint      `json:"campaign_id"` // Nullable jika donasi umum/zakat
	Campaign        Campaign   `gorm:"foreignKey:CampaignID" json:"campaign"`
	DonorID         uint       `gorm:"not null" json:"donor_id"`
	Donor           Donor      `gorm:"foreignKey:DonorID" json:"donor"`
	
	TransactionType string     `gorm:"type:enum('donasi', 'zakat', 'infaq');default:'donasi'" json:"transaction_type"`
	GrossAmount     float64    `gorm:"type:decimal(15,2);not null" json:"gross_amount"` // Total bayar donatur
	BiayaAdmin      float64    `gorm:"type:decimal(15,2);default:0" json:"biaya_admin"` // Potongan gateway
	NetAmount       float64    `gorm:"type:decimal(15,2);not null" json:"net_amount"`   // Masuk ke kas SIP
	
	PaymentType     string     `gorm:"type:varchar(50)" json:"payment_type"`
	Status          string     `gorm:"type:enum('pending', 'paid', 'failed', 'expired');default:'pending'" json:"status"`
	SnapToken       string     `gorm:"type:varchar(255)" json:"snap_token"`
	
	Notes           string     `gorm:"type:text" json:"notes"` // Pesan doa dari donatur
	PaidAt          *time.Time `json:"paid_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}