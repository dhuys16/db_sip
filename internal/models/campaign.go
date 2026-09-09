package models

import "gorm.io/gorm"

// Campaign menampung campaign donasi untuk kedua divisi, dibedakan lewat
// kolom Divisi (dipakai handler GetCampaigns via query param ?divisi=).
// Field masih minimal, akan dilengkapi di iterasi skema berikutnya.
type Campaign struct {
	gorm.Model
	Divisi string `gorm:"size:20;not null;index"`
	Judul  string `gorm:"size:255;not null"`
	Slug   string `gorm:"size:255;uniqueIndex;not null"`
}
