package controllers;

import (
    "net/http"
    "db_sip/app/models"
    "db_sip/config"

    "github.com/gin-gonic/gin"
)

// GetCampaigns - Mengambil data berdasarkan divisi
func GetCampaigns(c *gin.Context) {
    var campaigns []models.Campaign
    
    // Ambil parameter dari URL, contoh: /api/campaigns?division=lazsip
    division := c.Query("division") 

    query := config.DB
    if division != "" {
        query = query.Where("division = ?", division)
    }

    // Jalankan query SQL
    query.Find(&campaigns)

    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Data campaign berhasil diambil",
        "data":    campaigns,
    })
}