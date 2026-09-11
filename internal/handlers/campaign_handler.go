package handlers

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"

	"db_sip/internal/config"
	"db_sip/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CampaignRequest dipakai untuk body Create & Update (multipart/form-data,
// karena gambar dikirim langsung sebagai file dalam request yang sama)
type CampaignRequest struct {
	ProgramID    uint    `form:"program_id" binding:"required"`
	Title        string  `form:"title" binding:"required"`
	Description  string  `form:"description" binding:"required"`
	TargetAmount float64 `form:"target_amount" binding:"required,gt=0"`
	Division     string  `form:"division" binding:"required,oneof=lazsip sarsip"`
	Status       string  `form:"status" binding:"omitempty,oneof=active completed cancelled"`
	Category     string  `form:"category"` // opsional, bebas isi (misal: banjir, gempa_bumi untuk SARSIP)
	Location     string  `form:"location"` // opsional
	IsPinned     bool    `form:"is_pinned"`
	StartDate    string  `form:"start_date" binding:"required"` // contoh: "2026-09-10"
	EndDate      string  `form:"end_date" binding:"required"`
	// CurrentAmount cuma dipakai waktu Update, buat rekonsiliasi transfer manual
	// selama sistem payment gateway belum ada. Dibiarkan kosong (nil) waktu Create,
	// campaign baru selalu mulai dari 0.
	CurrentAmount *float64              `form:"current_amount"`
	Image         *multipart.FileHeader `form:"image"` // opsional
}

// generateCampaignCode bikin kode unik berformat CP001, CP002, dst.
func generateCampaignCode() (string, error) {
	var lastCampaign models.Campaign
	err := config.DB.Order("id desc").First(&lastCampaign).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "CP001", nil
		}
		return "", err
	}

	re := regexp.MustCompile(`(\d+)$`)
	match := re.FindStringSubmatch(lastCampaign.Slug)

	nextNumber := 1
	if len(match) == 2 {
		if n, convErr := strconv.Atoi(match[1]); convErr == nil {
			nextNumber = n + 1
		}
	}

	return fmt.Sprintf("CP%03d", nextNumber), nil
}

// CampaignPublicResponse membungkus data Campaign ditambah statistik yang
// dihitung langsung dari tabel Transaction (bukan kolom tersimpan), supaya
// selalu akurat tanpa perlu disinkronkan manual.
type CampaignPublicResponse struct {
	models.Campaign
	DonorCount int64 `json:"donor_count"`
}

// countCampaignDonors ngitung jumlah donatur UNIK yang transaksinya sudah
// "paid" buat sebuah campaign.
func countCampaignDonors(campaignID uint) int64 {
	var count int64
	config.DB.Model(&models.Transaction{}).
		Where("campaign_id = ? AND status = ?", campaignID, "paid").
		Distinct("donor_id").
		Count(&count)
	return count
}

// GetCampaignBySlugPublic - detail satu campaign untuk publik berdasarkan slug
func GetCampaignBySlugPublic(c *gin.Context) {
	slug := c.Param("slug")

	var campaign models.Campaign
	if err := config.DB.Preload("Program").Where("slug = ?", slug).First(&campaign).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Campaign tidak ditemukan"})
		return
	}

	response := CampaignPublicResponse{
		Campaign:   campaign,
		DonorCount: countCampaignDonors(campaign.ID),
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// GetCampaignsAdmin - list SEMUA campaign buat kebutuhan admin, termasuk yang
// statusnya 'completed'/'cancelled' (endpoint publik cuma nampilin yang 'active').
// Bisa difilter ?division= dan ?status=
func GetCampaignsAdmin(c *gin.Context) {
	var campaigns []models.Campaign
	division := c.Query("division")
	status := c.Query("status")
	category := c.Query("category")

	query := config.DB.Preload("Program").Order("created_at desc")
	if division != "" {
		query = query.Where("division = ?", division)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}

	if err := query.Find(&campaigns).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data campaign"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": campaigns})
}

// CreateCampaign - buat campaign baru. Gambar (kalau ada) dikirim langsung
// sebagai file di field "image".
func CreateCampaign(c *gin.Context) {
	var req CampaignRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid: " + err.Error()})
		return
	}

	// Pastikan Program yang direferensikan beneran ada
	var program models.Program
	if err := config.DB.First(&program, req.ProgramID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Program dengan ID tersebut tidak ditemukan"})
		return
	}

	startDate, err := parseActivityDate(req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format start_date tidak valid, gunakan YYYY-MM-DD"})
		return
	}
	endDate, err := parseActivityDate(req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format end_date tidak valid, gunakan YYYY-MM-DD"})
		return
	}
	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_date tidak boleh sebelum start_date"})
		return
	}

	if req.Status == "" {
		req.Status = "active"
	}

	code, err := generateCampaignCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat kode campaign"})
		return
	}

	var imageURL string
	if req.Image != nil {
		imageURL, err = saveImageFile(c, req.Image)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	campaign := models.Campaign{
		ProgramID:    req.ProgramID,
		Title:        req.Title,
		Slug:         code,
		Description:  req.Description,
		TargetAmount: req.TargetAmount,
		Division:     req.Division,
		Status:       req.Status,
		Image:        imageURL,
		Category:     req.Category,
		Location:     req.Location,
		IsPinned:     req.IsPinned,
		StartDate:    startDate,
		EndDate:      endDate,
	}

	if err := config.DB.Create(&campaign).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan campaign"})
		return
	}

	// Tempelin balik data Program yang udah kita ambil buat validasi tadi,
	// supaya response-nya nampilin detail Program, bukan struct kosong.
	campaign.Program = program

	c.JSON(http.StatusCreated, gin.H{
		"message": "Campaign berhasil dibuat",
		"data":    campaign,
	})
}

// UpdateCampaign - update campaign yang sudah ada berdasarkan slug (full update).
// Kalau field "image" tidak dikirim, gambar lama tetap dipakai. current_amount
// cuma berubah kalau field itu eksplisit dikirim.
func UpdateCampaign(c *gin.Context) {
	slug := c.Param("slug")

	var campaign models.Campaign
	if err := config.DB.Where("slug = ?", slug).First(&campaign).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Campaign tidak ditemukan"})
		return
	}

	var req CampaignRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid: " + err.Error()})
		return
	}

	var program models.Program
	if err := config.DB.First(&program, req.ProgramID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Program dengan ID tersebut tidak ditemukan"})
		return
	}

	startDate, err := parseActivityDate(req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format start_date tidak valid, gunakan YYYY-MM-DD"})
		return
	}
	endDate, err := parseActivityDate(req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format end_date tidak valid, gunakan YYYY-MM-DD"})
		return
	}
	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_date tidak boleh sebelum start_date"})
		return
	}

	if req.Status == "" {
		req.Status = "active"
	}

	campaign.ProgramID = req.ProgramID
	campaign.Title = req.Title
	campaign.Description = req.Description
	campaign.TargetAmount = req.TargetAmount
	campaign.Division = req.Division
	campaign.Status = req.Status
	campaign.Category = req.Category
	campaign.Location = req.Location
	campaign.IsPinned = req.IsPinned
	campaign.StartDate = startDate
	campaign.EndDate = endDate

	if req.CurrentAmount != nil {
		campaign.CurrentAmount = *req.CurrentAmount
	}

	if req.Image != nil {
		imageURL, err := saveImageFile(c, req.Image)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		campaign.Image = imageURL
	}

	if err := config.DB.Save(&campaign).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update campaign"})
		return
	}

	// Tempelin balik data Program yang udah kita ambil buat validasi tadi,
	// supaya response-nya nampilin detail Program, bukan struct kosong.
	campaign.Program = program

	c.JSON(http.StatusOK, gin.H{
		"message": "Campaign berhasil diupdate",
		"data":    campaign,
	})
}

// DeleteCampaign - hapus campaign berdasarkan slug
func DeleteCampaign(c *gin.Context) {
	slug := c.Param("slug")

	var campaign models.Campaign
	if err := config.DB.Where("slug = ?", slug).First(&campaign).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Campaign tidak ditemukan"})
		return
	}

	if err := config.DB.Delete(&campaign).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus campaign"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Campaign berhasil dihapus"})
}