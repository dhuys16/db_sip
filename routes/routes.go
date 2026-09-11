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

			public.GET("/blogs", handlers.GetBlogsPublic)
			public.GET("/blogs/:slug", handlers.GetBlogBySlugPublic)
			public.GET("/beneficiaries", handlers.GetBeneficiariesPublic)
			public.GET("/beneficiaries/:id", handlers.GetBeneficiaryDetailPublic)
			public.GET("/activities", handlers.GetActivitiesPublic)
			public.GET("/activities/:slug", handlers.GetActivityBySlugPublic)
			public.GET("/campaigns", handlers.GetCampaignsPublic)
			public.GET("/campaigns/:slug", handlers.GetCampaignBySlugPublic)
			public.GET("/programs", handlers.GetProgramsPublic)
			public.GET("/programs/:slug", handlers.GetProgramBySlugPublic)
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

			// CRUD Activity
			admin.POST("/activities", handlers.CreateActivity)
			admin.PUT("/activities/:slug", handlers.UpdateActivity)
			admin.DELETE("/activities/:slug", handlers.DeleteActivity)
			admin.POST("/upload", handlers.UploadImage)

			// CRUD Blog (Create, Update, Delete - butuh login admin)
			admin.POST("/blogs", handlers.CreateBlog)
			admin.PUT("/blogs/:slug", handlers.UpdateBlog)
			admin.DELETE("/blogs/:slug", handlers.DeleteBlog)

			// CRUD Campaign (Create, Update, Delete - butuh login admin)
			admin.POST("/campaigns", handlers.CreateCampaign)
			admin.PUT("/campaigns/:slug", handlers.UpdateCampaign)
			admin.DELETE("/campaigns/:slug", handlers.DeleteCampaign)

			// CRUD Program (Create, Update, Delete - butuh login admin)
			admin.GET("/programs", handlers.GetProgramsAdmin)
			admin.POST("/programs", handlers.CreateProgram)
			admin.PUT("/programs/:slug", handlers.UpdateProgram)
			admin.DELETE("/programs/:slug", handlers.DeleteProgram)

			// CRUD Beneficiary 
			admin.GET("/beneficiaries", handlers.GetBeneficiariesAdmin)
			admin.POST("/beneficiaries", handlers.CreateBeneficiary)
			admin.PUT("/beneficiaries/:id", handlers.UpdateBeneficiary)
			admin.DELETE("/beneficiaries/:id", handlers.DeleteBeneficiary)
		}
	}
}