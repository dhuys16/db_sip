package models

import "time"

type Activity struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"type:varchar(255);not null" json:"title"`
	Slug        string    `gorm:"type:varchar(255);unique;not null" json:"slug"`
	Description string    `gorm:"type:text;not null" json:"description"`
	ImageURL    string    `gorm:"type:varchar(255)" json:"image_url"`
	Division    string    `gorm:"type:enum('lazsip', 'sarsip');not null" json:"division"`
	IsPinned    bool      `gorm:"default:false" json:"is_pinned"`
	ActivityDate time.Time `json:"activity_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}