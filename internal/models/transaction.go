package models

import "gorm.io/gorm"

// Transaction menampung transaksi donasi maupun zakat dalam satu tabel,
// dibedakan lewat kolom Tipe. Status HANYA boleh diubah lewat handler
// webhook payment gateway setelah verifikasi signature - tidak pernah
// dari request pengguna (redirect sukses, dsb).
// Field masih minimal, akan dilengkapi di iterasi skema berikutnya.
type Transaction struct {
	gorm.Model
	Tipe       string  `gorm:"size:20;not null"` // "donasi" | "zakat"
	CampaignID *uint   `gorm:"index"`            // nullable, kosong kalau tipe zakat
	DonorID    uint    `gorm:"not null;index"`
	Nominal    float64 `gorm:"not null"`
	Status     string  `gorm:"size:20;not null;default:pending"`
}
