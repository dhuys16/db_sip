package main

import (
	"log"
	"os"

	"db_sip/database/migrations"
	"db_sip/internal/config"
	"db_sip/internal/middleware"
	"db_sip/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load Environment & Connect Database
	// Pastikan fungsi ini di dalam internal/config/database.go sudah memanggil godotenv.Load()
	config.ConnectDatabase()

	// 2. Jalankan AutoMigrate (Fase MVP)
	migrations.RunMigrate()

	// 3. Inisialisasi Router Gin
	r := gin.Default()

	// Serve file hasil upload (gambar) supaya bisa diakses lewat /uploads/<nama_file>
	r.Static("/uploads", "./uploads")

	// Aktifkan CORS supaya frontend (Next.js dkk) di domain/port berbeda bisa akses API ini
	r.Use(middleware.CorsMiddleware())

	// 4. Setup Routes (Publik, Admin, Webhook)
	routes.SetupRoutes(r)

	// 5. Tentukan Port & Jalankan Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server SIP Backend berjalan di port :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}