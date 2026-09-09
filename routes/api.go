package routes;

import (
    "db_sip/app/controllers"

    "github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
    // Grouping route API
    api := r.Group("/api")
    {
        api.GET("/ping", func(c *gin.Context) {
            c.JSON(200, gin.H{"message": "pong"})
        })

        // Route untuk Campaign
        campaigns := api.Group("/campaigns")
        {
            campaigns.GET("/", controllers.GetCampaigns)
        }
    }
}