package routes

import (
	"db_sip/internal/handlers"
	"db_sip/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		// 1. Endpoint Publik
		public := v1.Group("/public")
		{
			public.GET("/ping", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "API SIP Public OK"})
			})
			
			// Routes Read-Only untuk Frontend
			public.GET("/campaigns", handlers.GetCampaignsPublic)
			public.GET("/blogs", handlers.GetBlogsPublic)
			public.GET("/beneficiaries", handlers.GetBeneficiariesPublic)
		}

		// 2. Endpoint Auth
		auth := v1.Group("/auth")
		{
			auth.POST("/login", handlers.AdminLogin)
		}

		// 3. Endpoint Admin (Protected)
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware())
		{
			admin.GET("/me", func(c *gin.Context) {
				adminID, _ := c.Get("admin_id")
				c.JSON(200, gin.H{"admin_id": adminID, "message": "Token valid!"})
			})
			
			// TODO: CRUD Endpoint untuk Admin akan ditambahkan di sini
		}
	}
}