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
			public.GET("/beneficiaries/:id", handlers.GetBeneficiaryDetailPublic)
			public.GET("/activities", handlers.GetActivitiesPublic)
			public.GET("/activities/:slug", handlers.GetActivityBySlugPublic)
			public.GET("/blogs/:slug", handlers.GetBlogBySlugPublic)
			public.GET("/campaigns/:slug", handlers.GetCampaignBySlugPublic)
			public.GET("/programs", handlers.GetProgramsPublic)
			public.GET("/programs/:slug", handlers.GetProgramBySlugPublic)

			// Simulasi Donasi (belum ada payment gateway asli)
			public.POST("/donations", handlers.CreateDonation)
			public.GET("/donations/:order_id", handlers.GetDonationStatusPublic)
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

			admin.POST("/logout", handlers.AdminLogout)

			// Upload gambar (dipakai buat isi field image_url resource lain)
			admin.POST("/upload", handlers.UploadImage)

			// CRUD Activity (Create, Update, Delete - butuh login admin)
			admin.POST("/activities", handlers.CreateActivity)
			admin.PUT("/activities/:slug", handlers.UpdateActivity)
			admin.DELETE("/activities/:slug", handlers.DeleteActivity)

			// CRUD Blog (Create, Update, Delete - butuh login admin)
			admin.GET("/blogs", handlers.GetBlogsAdmin)
			admin.GET("/blogs/:id", handlers.GetBlogByIDAdmin)
			admin.POST("/blogs", handlers.CreateBlog)
			admin.PUT("/blogs/:slug", handlers.UpdateBlog)
			admin.DELETE("/blogs/:slug", handlers.DeleteBlog)

			// CRUD Campaign (Create, Update, Delete - butuh login admin)
			admin.GET("/campaigns", handlers.GetCampaignsAdmin)
			admin.POST("/campaigns", handlers.CreateCampaign)
			admin.PUT("/campaigns/:slug", handlers.UpdateCampaign)
			admin.DELETE("/campaigns/:slug", handlers.DeleteCampaign)

			// CRUD Program (Create, Update, Delete - butuh login admin)
			admin.GET("/programs", handlers.GetProgramsAdmin)
			admin.POST("/programs", handlers.CreateProgram)
			admin.PUT("/programs/:slug", handlers.UpdateProgram)
			admin.DELETE("/programs/:slug", handlers.DeleteProgram)

			// CRUD Beneficiary (Create, Update, Delete - butuh login admin)
			admin.GET("/beneficiaries", handlers.GetBeneficiariesAdmin)
			admin.POST("/beneficiaries", handlers.CreateBeneficiary)
			admin.PUT("/beneficiaries/:id", handlers.UpdateBeneficiary)
			admin.DELETE("/beneficiaries/:id", handlers.DeleteBeneficiary)

			// Transaksi & simulasi pembayaran (sementara, sebelum payment gateway asli)
			admin.GET("/transactions", handlers.GetTransactionsAdmin)
			admin.POST("/transactions/:id/simulate-paid", handlers.SimulatePaymentSuccess)
			admin.POST("/transactions/:id/simulate-failed", handlers.SimulatePaymentFailed)

			// TODO: CRUD Endpoint lain akan ditambahkan di sini
		}
	}
}