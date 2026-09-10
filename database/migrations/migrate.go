package migrations

import (
	"log"

	"db_sip/internal/config"
	"db_sip/internal/models"
	"db_sip/internal/utils"
)

func RunMigrate() {
	log.Println("Memulai AutoMigrate...")

	err := config.DB.AutoMigrate(
		&models.Admin{},
		&models.Program{},
		&models.Blog{},
		&models.Activity{},
		&models.Campaign{},
		&models.Donor{},
		&models.PaymentFee{},
		&models.Transaction{},
		&models.Beneficiary{},
	)

	if err != nil {
		log.Fatalf("Gagal migrasi database: %v", err)
	}

	// Seeder Admin Pertama
	var count int64
	config.DB.Model(&models.Admin{}).Count(&count)
	if count == 0 {
		hash, _ := utils.HashPassword("password123")
		defaultAdmin := models.Admin{
			Name:         "Super Admin",
			Email:        "admin@sip.org",
			PasswordHash: hash,
			Role:         "superadmin",
		}
		
		if err := config.DB.Create(&defaultAdmin).Error; err != nil {
			log.Printf("Gagal membuat admin default: %v", err)
		} else {
			log.Println("Seeder: Akun admin default berhasil dibuat! (Email: admin@sip.org | Pass: password123)")
		}
	}

	log.Println("Migrasi dan Seeder selesai.")
}