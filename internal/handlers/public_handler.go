package handlers

import (
	"net/http"
	"strings"

	"db_sip/internal/config"
	"db_sip/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetCampaignsPublic mengambil data campaign aktif
func GetCampaignsPublic(c *gin.Context) {
	var campaigns []models.Campaign
	division := c.Query("division") // lazsip atau sarsip
	category := c.Query("category")

	query := config.DB.Where("status = ?", "active")
	if division != "" {
		query = query.Where("division = ?", division)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// Mengambil data beserta relasi Program
	query.Preload("Program").Find(&campaigns)

	// Bungkus tiap campaign dengan donor_count yang dihitung langsung dari Transaction
	responses := make([]CampaignPublicResponse, len(campaigns))
	for i, campaign := range campaigns {
		responses[i] = CampaignPublicResponse{
			Campaign:   campaign,
			DonorCount: countCampaignDonors(campaign.ID),
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": responses})
}
// GetBlogsPublic mengambil artikel/berita
func GetBlogsPublic(c *gin.Context) {
	var blogs []models.Blog
	division := c.Query("division")

	query := config.DB.Order("created_at desc")
	if division != "" {
		query = query.Where("division = ?", division)
	}

	// Preload nama admin penulisnya, abaikan password dsb
	query.Preload("Admin", func(db *gorm.DB) *gorm.DB {
		return db.Select("ID, Name")
	}).Find(&blogs)

	c.JSON(http.StatusOK, gin.H{"data": blogs})
}

// GetBeneficiariesPublic mengambil data penerima manfaat dengan proteksi privasi
func GetBeneficiariesPublic(c *gin.Context) {
	var beneficiaries []models.Beneficiary
	
	// Ambil data dari DB
	if err := config.DB.Preload("Program").Find(&beneficiaries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data"})
		return
	}

	// Data Sanitization / Privacy Masking
	for i := range beneficiaries {
		switch beneficiaries[i].PrivacyLevel {
		case "hidden":
			beneficiaries[i].FullName = "Hamba Allah"
			beneficiaries[i].Address = "Dirahasiakan"
			beneficiaries[i].Image = "" // Sembunyikan foto
		case "initials":
			beneficiaries[i].FullName = maskNameWithInitials(beneficiaries[i].FullName)
			// Alamat bisa di-masking sebagian jika perlu, untuk MVP disembunyikan
			beneficiaries[i].Address = "Telah disensor" 
		case "public":
			// Biarkan data apa adanya
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": beneficiaries})
}

// Helper function untuk menyamarkan nama (Budi Santoso -> B*** S***)
func maskNameWithInitials(name string) string {
	words := strings.Fields(name)
	for i, word := range words {
		if len(word) > 1 {
			words[i] = string(word[0]) + strings.Repeat("*", len(word)-1)
		}
	}
	return strings.Join(words, " ")
}