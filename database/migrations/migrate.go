package migrations

import (
	"fmt"
	"log"

	"db_sip/internal/config"
	"db_sip/internal/models"
)

func RunMigrate() {
	err := config.DB.AutoMigrate(
		&models.Admin{},
		&models.Blog{},
		&models.Campaign{},
		&models.Donor{},
		&models.Transaction{},
	)
	if err != nil {
		log.Fatal("Gagal migrasi:", err)
	}
	fmt.Println("Migrasi database berhasil!")
}
