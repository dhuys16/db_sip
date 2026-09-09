package routes

import (
	"db_sip/internal/handlers"

	"github.com/gin-gonic/gin"
)

// SetupRoutes mendaftarkan seluruh route API, dikelompokkan per versi lalu
// per akses (public / admin / webhook).
func SetupRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})

		public := v1.Group("/public")
		{
			campaigns := public.Group("/campaigns")
			{
				campaigns.GET("", handlers.GetCampaigns)
			}
		}

		// admin := v1.Group("/admin")
		// TODO: pasang middleware auth JWT di sini setelah dibuat, baru
		// daftarkan route admin (CRUD blog/campaign/dst, full-column beneficiaries).

		// webhook := v1.Group("/webhook")
		// TODO: route webhook Midtrans setelah service pembayaran dibuat -
		// wajib verifikasi signature sebelum ubah status transaction.
	}
}
