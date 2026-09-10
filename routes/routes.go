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

			public.GET("/campaigns", handlers.GetCampaignsPublic)
			public.GET("/blogs", handlers.GetBlogsPublic)
			public.GET("/blogs/:slug", handlers.GetBlogBySlugPublic)
			public.GET("/beneficiaries", handlers.GetBeneficiariesPublic)
			public.GET("/activities", handlers.GetActivities)
			public.GET("/activities/:slug", handlers.GetActivityBySlug)
		}

		// 2. Endpoint Auth
		auth := v1.Group("/auth")
		{
			auth.POST("/login", handlers.AdminLogin)
			auth.POST("/logout", handlers.AdminLogout)
		}

		// 3. Endpoint Admin (Protected)
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware())
		{
			admin.GET("/me", func(c *gin.Context) {
				adminID, _ := c.Get("admin_id")
				c.JSON(200, gin.H{"admin_id": adminID, "message": "Token valid!"})
			})

			// CRUD Activity
			admin.POST("/activities", handlers.CreateActivity)
			admin.PUT("/activities/:slug", handlers.UpdateActivity)
			admin.DELETE("/activities/:slug", handlers.DeleteActivity)
			admin.POST("/upload", handlers.UploadImage)

			// CRUD Blog (Create, Update, Delete - butuh login admin)
			admin.POST("/blogs", handlers.CreateBlog)
			admin.PUT("/blogs/:slug", handlers.UpdateBlog)
			admin.DELETE("/blogs/:slug", handlers.DeleteBlog)
		}
	}
}