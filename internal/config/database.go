package config

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// ConnectDatabase membuka koneksi GORM ke MySQL memakai kredensial dari
// environment variable (DB_HOST, DB_PORT, DB_USER, DB_PASS, DB_NAME).
// Panggil config.LoadEnv() dulu sebelum ini kalau kredensialnya dari file .env.
func ConnectDatabase() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal terkoneksi ke database:", err)
	}

	DB = database
	fmt.Println("Koneksi database berhasil!")
}
