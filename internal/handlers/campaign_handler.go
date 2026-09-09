package handlers

import (
	"net/http"

	"db_sip/internal/config"
	"db_sip/internal/models"

	"github.com/gin-gonic/gin"
)

// GetCampaigns mengambil daftar campaign, opsional difilter per divisi
// lewat query param ?divisi=lazsip atau ?divisi=sarsip.
func GetCampaigns(c *gin.Context) {
	var campaigns []models.Campaign

	divisi := c.Query("divisi")

	query := config.DB
	if divisi != "" {
		query = query.Where("divisi = ?", divisi)
	}

	query.Find(&campaigns)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Data campaign berhasil diambil",
		"data":    campaigns,
	})
}
