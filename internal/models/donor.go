package models

import "gorm.io/gorm"

// Donor menampung data donatur. Data individual TIDAK PERNAH diekspos
// sebagai list/feed publik - publik hanya boleh lihat counter agregat.
// Field masih minimal, akan dilengkapi di iterasi skema berikutnya.
type Donor struct {
	gorm.Model
	Nama  string `gorm:"size:150;not null"`
	Email string `gorm:"size:150"`
	NoHp  string `gorm:"size:30"`
}
