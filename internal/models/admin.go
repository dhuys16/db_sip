package models

import "gorm.io/gorm"

// Admin merepresentasikan akun admin (satu-satunya role saat ini, tanpa
// scoping per divisi - satu admin bisa kelola data lazsip & sarsip).
// Field masih minimal, akan dilengkapi di iterasi skema berikutnya.
type Admin struct {
	gorm.Model
	Name         string `gorm:"size:150;not null"`
	Email        string `gorm:"size:150;uniqueIndex;not null"`
	PasswordHash string `gorm:"size:255;not null"`
}
