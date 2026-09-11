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

// ProgramRequest dipakai untuk body Create & Update (multipart/form-data,
// karena gambar dikirim langsung sebagai file dalam request yang sama)
type ProgramRequest struct {
	Name        string                `form:"name" binding:"required"`
	Description string                `form:"description"`
	Division    string                `form:"division" binding:"required,oneof=lazsip sarsip"`
	IsActive    *bool                 `form:"is_active"` // pointer supaya bisa bedain "tidak dikirim" vs "false"
	Image       *multipart.FileHeader `form:"image"`     // opsional
}

// generateProgramCode bikin kode unik berformat PG001, PG002, dst.
func generateProgramCode() (string, error) {
	var lastProgram models.Program
	err := config.DB.Order("id desc").First(&lastProgram).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "PG001", nil
		}
		return "", err
	}

	re := regexp.MustCompile(`(\d+)$`)
	match := re.FindStringSubmatch(lastProgram.Slug)

	nextNumber := 1
	if len(match) == 2 {
		if n, convErr := strconv.Atoi(match[1]); convErr == nil {
			nextNumber = n + 1
		}
	}

	return fmt.Sprintf("PG%03d", nextNumber), nil
}

// GetProgramsPublic - list program untuk publik. Cuma yang is_active = true yang ditampilkan.
// Bisa difilter ?division=lazsip atau ?division=sarsip
func GetProgramsPublic(c *gin.Context) {
	var programs []models.Program
	division := c.Query("division")

	query := config.DB.Where("is_active = ?", true).Order("created_at desc")
	if division != "" {
		query = query.Where("division = ?", division)
	}

	if err := query.Find(&programs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data program"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": programs})
}

// GetProgramBySlugPublic - detail satu program untuk publik berdasarkan slug
// (ditampilkan meski is_active = false, supaya link yang sudah dibagikan tetap bisa dibuka)
func GetProgramBySlugPublic(c *gin.Context) {
	slug := c.Param("slug")

	var program models.Program
	if err := config.DB.Where("slug = ?", slug).First(&program).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Program tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": program})
}

// GetProgramsAdmin - list SEMUA program (termasuk yang tidak aktif) buat kebutuhan admin,
// misal buat dropdown pilihan program pas bikin Campaign baru.
func GetProgramsAdmin(c *gin.Context) {
	var programs []models.Program
	if err := config.DB.Order("created_at desc").Find(&programs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data program"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": programs})
}

// CreateProgram - buat program baru. Gambar (kalau ada) dikirim langsung
// sebagai file di field "image".
func CreateProgram(c *gin.Context) {
	var req ProgramRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid: " + err.Error()})
		return
	}

	slug, err := generateProgramCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat kode program"})
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	var imageURL string
	if req.Image != nil {
		imageURL, err = saveImageFile(c, req.Image)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	program := models.Program{
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		Image:       imageURL,
		Division:    req.Division,
		IsActive:    isActive,
	}

	if err := config.DB.Create(&program).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan program"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Program berhasil dibuat",
		"data":    program,
	})
}

// UpdateProgram - update program yang sudah ada berdasarkan slug (full update).
// Kode (slug) tidak pernah berubah lagi setelah dibuat. Kalau field "image" tidak
// dikirim, gambar lama tetap dipakai.
func UpdateProgram(c *gin.Context) {
	slug := c.Param("slug")

	var program models.Program
	if err := config.DB.Where("slug = ?", slug).First(&program).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Program tidak ditemukan"})
		return
	}

	var req ProgramRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid: " + err.Error()})
		return
	}

	isActive := program.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	program.Name = req.Name
	program.Description = req.Description
	program.Division = req.Division
	program.IsActive = isActive

	if req.Image != nil {
		imageURL, err := saveImageFile(c, req.Image)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		program.Image = imageURL
	}

	if err := config.DB.Save(&program).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update program"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Program berhasil diupdate",
		"data":    program,
	})
}

// DeleteProgram - hapus program berdasarkan slug.
// Ditolak (409) kalau masih ada Campaign atau Beneficiary yang nempel ke program ini,
// supaya data relasinya gak jadi yatim (orphan) di database.
func DeleteProgram(c *gin.Context) {
	slug := c.Param("slug")

	var program models.Program
	if err := config.DB.Where("slug = ?", slug).First(&program).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Program tidak ditemukan"})
		return
	}

	var campaignCount int64
	config.DB.Model(&models.Campaign{}).Where("program_id = ?", program.ID).Count(&campaignCount)
	if campaignCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Program tidak bisa dihapus karena masih dipakai oleh campaign yang ada"})
		return
	}

	var beneficiaryCount int64
	config.DB.Model(&models.Beneficiary{}).Where("program_id = ?", program.ID).Count(&beneficiaryCount)
	if beneficiaryCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Program tidak bisa dihapus karena masih punya data beneficiary"})
		return
	}

	if err := config.DB.Delete(&program).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus program"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Program berhasil dihapus"})
}