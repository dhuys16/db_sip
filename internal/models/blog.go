package models

import "gorm.io/gorm"

// Blog menampung konten blog/berita untuk kedua divisi (lazsip & sarsip),
// dibedakan lewat kolom Divisi, bukan tabel terpisah.
// Field masih minimal, akan dilengkapi di iterasi skema berikutnya.
type Blog struct {
	gorm.Model
	Divisi string `gorm:"size:20;not null;index"`
	Judul  string `gorm:"size:255;not null"`
	Slug   string `gorm:"size:255;uniqueIndex;not null"`
	Konten string `gorm:"type:text"`
}
