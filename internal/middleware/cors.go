package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CorsMiddleware ngatur header CORS supaya frontend (Next.js dkk) yang jalan
// di domain/port berbeda bisa akses API ini dari browser.
//
// Daftar origin yang diizinkan diambil dari env ALLOWED_ORIGINS (dipisah koma),
// misal: ALLOWED_ORIGINS=https://erdeconsulting.com,https://www.erdeconsulting.com
// Kalau env-nya kosong/tidak diisi, default-nya "*" (semua origin diizinkan) --
// ini cocok buat development, tapi sebaiknya diisi domain spesifik pas production
// biar gak sembarang situs bisa manggil API ini dari browser.
func CorsMiddleware() gin.HandlerFunc {
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if allowedOrigins == "" || allowedOrigins == "*" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			for _, allowed := range strings.Split(allowedOrigins, ",") {
				if strings.TrimSpace(allowed) == origin {
					c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Browser ngirim OPTIONS dulu (preflight) sebelum request beneran
		// buat method non-GET/POST sederhana. Cukup jawab 204, jangan diteruskan
		// ke handler asli.
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}