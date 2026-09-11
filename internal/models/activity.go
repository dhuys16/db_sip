package models

import "time"

type Activity struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	Title            string    `gorm:"type:varchar(255);not null" json:"title"`
	Slug             string    `gorm:"type:varchar(255);unique;not null" json:"slug"`
	Description      string    `gorm:"type:text;not null" json:"description"`
	Image            string    `gorm:"column:image;type:varchar(255)" json:"image"`
	Division         string    `gorm:"type:enum('lazsip', 'sarsip');not null" json:"division"`
	Category         string    `gorm:"type:varchar(50)" json:"category"` // opsional, bebas isi
	Location         string    `gorm:"type:varchar(255)" json:"location"`
	Team             string    `gorm:"type:varchar(255)" json:"team"`      // opsional, misal "18 relawan & 6 personel SAR"
	ItemsDistributed string    `gorm:"type:text" json:"items_distributed"` // opsional, ringkasan barang yang disalurkan
	BeneficiaryCount int       `gorm:"default:0" json:"beneficiary_count"` // opsional, jumlah agregat penerima manfaat
	CampaignID       *uint     `json:"campaign_id"`                        // opsional, nullable
	Campaign         Campaign  `gorm:"foreignKey:CampaignID" json:"campaign"`
	IsPinned         bool      `gorm:"default:false" json:"is_pinned"`
	ActivityDate     time.Time `json:"activity_date"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}