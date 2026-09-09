package main

import (
	"os"

	"db_sip/database/migrations"
	"db_sip/internal/config"
	"db_sip/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load environment variables dari .env
	config.LoadEnv()

	// 2. Connect Database
	config.ConnectDatabase()

	// 3. Jalankan Migrasi
	migrations.RunMigrate()

	// 4. Inisialisasi Gin Router
	r := gin.Default()

	// 5. Setup Routes
	routes.SetupRoutes(r)

	// 6. Jalankan Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
