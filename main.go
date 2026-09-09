package main

import (
    "os"
    "db_sip/config"
    "db_sip/database/migrations"
    "db_sip/routes"

    "github.com/gin-gonic/gin"
)

func main() {
    // 1. Connect Database
    config.ConnectDatabase()

    // 2. Jalankan Migrasi
    migrations.RunMigrate()

    // 3. Inisialisasi Gin Router
    r := gin.Default()

    // 4. Setup Routes
    routes.SetupRoutes(r)

    // 5. Jalankan Server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    r.Run(":" + port)
}