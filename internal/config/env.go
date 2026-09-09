package config

import (
	"log"

	"github.com/joho/godotenv"
)

// LoadEnv membaca file .env ke environment variable proses. Dipanggil sekali
// di awal main(). Kalau .env tidak ada (misal di production yang env-nya
// di-set langsung oleh platform/orchestrator), ini tidak fatal.
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("Info: .env tidak ditemukan, pakai environment variable yang sudah di-set")
	}
}
