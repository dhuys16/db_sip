package handlers

import (
	"mime/multipart"
	"net/http"
	"strconv"

	"db_sip/internal/config"
	"db_sip/internal/models"

	"github.com/gin-gonic/gin"
)

// BeneficiaryRequest dipakai untuk body Create & Update (multipart/form-data,
// karena gambar dikirim langsung sebagai file dalam request yang sama)
type BeneficiaryRequest struct {
	ProgramID      uint                  `form:"program_id" binding:"required"`
	FullName       string                `form:"full_name" binding:"required"`
	Address        string                `form:"address"`
	AmountReceived float64               `form:"amount_received"`
	DateReceived   string                `form:"date_received" binding:"required"` // contoh: "2026-09-10"
	PrivacyLevel   string                `form:"privacy_level" binding:"omitempty,oneof=public initials hidden"`
	Story          string                `form:"story"`
	Image          *multipart.FileHeader `form:"image"` // opsional
}

// getBeneficiaryIDParam ambil & validasi :id dari URL
func getBeneficiaryIDParam(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// maskBeneficiary nerapin masking privasi yang sama kayak di GetBeneficiariesPublic,
// dipakai bareng buat endpoint detail satuan.
func maskBeneficiary(b *models.Beneficiary) {
	switch b.PrivacyLevel {
	case "hidden":
		b.FullName = "Hamba Allah"
		b.Address = "Dirahasiakan"
		b.Image = ""
	case "initials":
		b.FullName = maskNameWithInitials(b.FullName)
		b.Address = "Telah disensor"
	case "public":
		// Biarkan data apa adanya
	}
}

// GetBeneficiaryDetailPublic - detail satu beneficiary untuk publik berdasarkan ID,
// dengan masking privasi yang sama kayak versi list-nya.
func GetBeneficiaryDetailPublic(c *gin.Context) {
	id, err := getBeneficiaryIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID beneficiary tidak valid"})
		return
	}

	var beneficiary models.Beneficiary
	if err := config.DB.Preload("Program").First(&beneficiary, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Beneficiary tidak ditemukan"})
		return
	}

	maskBeneficiary(&beneficiary)

	c.JSON(http.StatusOK, gin.H{"data": beneficiary})
}

// GetBeneficiariesAdmin - list SEMUA beneficiary buat kebutuhan admin, TANPA masking
// privasi (admin perlu lihat data asli buat keperluan pengelolaan/verifikasi).
// Bisa difilter ?program_id=
func GetBeneficiariesAdmin(c *gin.Context) {
	var beneficiaries []models.Beneficiary
	programID := c.Query("program_id")

	query := config.DB.Preload("Program").Order("date_received desc")
	if programID != "" {
		query = query.Where("program_id = ?", programID)
	}

	if err := query.Find(&beneficiaries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data beneficiary"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": beneficiaries})
}

// CreateBeneficiary - catat penerima manfaat baru di bawah sebuah Program.
// Gambar (kalau ada) dikirim langsung sebagai file di field "image".
func CreateBeneficiary(c *gin.Context) {
	var req BeneficiaryRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid: " + err.Error()})
		return
	}

	var program models.Program
	if err := config.DB.First(&program, req.ProgramID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Program dengan ID tersebut tidak ditemukan"})
		return
	}

	dateReceived, err := parseActivityDate(req.DateReceived)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format date_received tidak valid, gunakan YYYY-MM-DD"})
		return
	}

	if req.PrivacyLevel == "" {
		req.PrivacyLevel = "hidden" // default paling aman, sesuai default di model
	}

	var imageURL string
	if req.Image != nil {
		imageURL, err = saveImageFile(c, req.Image)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	beneficiary := models.Beneficiary{
		ProgramID:      req.ProgramID,
		FullName:       req.FullName,
		Address:        req.Address,
		AmountReceived: req.AmountReceived,
		DateReceived:   dateReceived,
		PrivacyLevel:   req.PrivacyLevel,
		Story:          req.Story,
		Image:          imageURL,
	}

	if err := config.DB.Create(&beneficiary).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan beneficiary"})
		return
	}

	beneficiary.Program = program

	c.JSON(http.StatusCreated, gin.H{
		"message": "Beneficiary berhasil dicatat",
		"data":    beneficiary,
	})
}

// UpdateBeneficiary - update data beneficiary yang sudah ada berdasarkan ID (full update).
// Kalau field "image" tidak dikirim, gambar lama tetap dipakai.
func UpdateBeneficiary(c *gin.Context) {
	id, err := getBeneficiaryIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID beneficiary tidak valid"})
		return
	}

	var beneficiary models.Beneficiary
	if err := config.DB.First(&beneficiary, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Beneficiary tidak ditemukan"})
		return
	}

	var req BeneficiaryRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid: " + err.Error()})
		return
	}

	var program models.Program
	if err := config.DB.First(&program, req.ProgramID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Program dengan ID tersebut tidak ditemukan"})
		return
	}

	dateReceived, err := parseActivityDate(req.DateReceived)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format date_received tidak valid, gunakan YYYY-MM-DD"})
		return
	}

	if req.PrivacyLevel == "" {
		req.PrivacyLevel = "hidden"
	}

	beneficiary.ProgramID = req.ProgramID
	beneficiary.FullName = req.FullName
	beneficiary.Address = req.Address
	beneficiary.AmountReceived = req.AmountReceived
	beneficiary.DateReceived = dateReceived
	beneficiary.PrivacyLevel = req.PrivacyLevel
	beneficiary.Story = req.Story

	if req.Image != nil {
		imageURL, err := saveImageFile(c, req.Image)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		beneficiary.Image = imageURL
	}

	if err := config.DB.Save(&beneficiary).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update beneficiary"})
		return
	}

	beneficiary.Program = program

	c.JSON(http.StatusOK, gin.H{
		"message": "Beneficiary berhasil diupdate",
		"data":    beneficiary,
	})
}

// DeleteBeneficiary - hapus data beneficiary berdasarkan ID
func DeleteBeneficiary(c *gin.Context) {
	id, err := getBeneficiaryIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID beneficiary tidak valid"})
		return
	}

	var beneficiary models.Beneficiary
	if err := config.DB.First(&beneficiary, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Beneficiary tidak ditemukan"})
		return
	}

	if err := config.DB.Delete(&beneficiary).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus beneficiary"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Beneficiary berhasil dihapus"})
}