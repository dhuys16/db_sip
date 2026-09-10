package handlers

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"db_sip/internal/config"
	"db_sip/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ActivityRequest dipakai untuk body Create & Update (multipart/form-data,
// karena gambar dikirim langsung sebagai file dalam request yang sama)
type ActivityRequest struct {
	Title        string                `form:"title" binding:"required"`
	Description  string                `form:"description" binding:"required"`
	Division     string                `form:"division" binding:"required,oneof=lazsip sarsip"`
	IsPinned     bool                  `form:"is_pinned"`
	ActivityDate string                `form:"activity_date" binding:"required"` // contoh: "2026-09-10"
	Image        *multipart.FileHeader `form:"image"`                            // opsional
}

// parseActivityDate mencoba beberapa format tanggal yang umum dipakai
func parseActivityDate(value string) (time.Time, error) {
	formats := []string{"2006-01-02", time.RFC3339, "2006-01-02T15:04:05"}

	var lastErr error
	for _, layout := range formats {
		t, err := time.Parse(layout, value)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}

// generateActivityCode bikin kode unik berformat kg001, kg002, dst.
func generateActivityCode() (string, error) {
	var lastActivity models.Activity
	err := config.DB.Order("id desc").First(&lastActivity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "kg001", nil
		}
		return "", err
	}

	re := regexp.MustCompile(`(\d+)$`)
	match := re.FindStringSubmatch(lastActivity.Slug)

	nextNumber := 1
	if len(match) == 2 {
		if n, convErr := strconv.Atoi(match[1]); convErr == nil {
			nextNumber = n + 1
		}
	}

	return fmt.Sprintf("kg%03d", nextNumber), nil
}

// GetActivitiesPublic - list semua activity untuk publik, bisa difilter ?division=lazsip atau ?division=sarsip
func GetActivitiesPublic(c *gin.Context) {
	var activities []models.Activity
	division := c.Query("division")

	query := config.DB.Order("activity_date desc")
	if division != "" {
		query = query.Where("division = ?", division)
	}

	if err := query.Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data activity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": activities})
}

// GetActivityBySlugPublic - detail satu activity untuk publik berdasarkan slug
func GetActivityBySlugPublic(c *gin.Context) {
	slug := c.Param("slug")

	var activity models.Activity
	if err := config.DB.Where("slug = ?", slug).First(&activity).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Activity tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": activity})
}

// CreateActivity - buat activity baru. Gambar (kalau ada) dikirim langsung
// sebagai file di field "image".
func CreateActivity(c *gin.Context) {
	var req ActivityRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid: " + err.Error()})
		return
	}

	activityDate, err := parseActivityDate(req.ActivityDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format activity_date tidak valid, gunakan YYYY-MM-DD"})
		return
	}

	code, err := generateActivityCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat kode activity"})
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

	activity := models.Activity{
		Title:        req.Title,
		Slug:         code,
		Description:  req.Description,
		Image:        imageURL,
		Division:     req.Division,
		IsPinned:     req.IsPinned,
		ActivityDate: activityDate,
	}

	if err := config.DB.Create(&activity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan activity"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Activity berhasil dibuat",
		"data":    activity,
	})
}

// UpdateActivity - update activity yang sudah ada berdasarkan slug (full update).
// Kalau field "image" tidak dikirim, gambar lama tetap dipakai (tidak dihapus).
func UpdateActivity(c *gin.Context) {
	slug := c.Param("slug")

	var activity models.Activity
	if err := config.DB.Where("slug = ?", slug).First(&activity).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Activity tidak ditemukan"})
		return
	}

	var req ActivityRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid: " + err.Error()})
		return
	}

	activityDate, err := parseActivityDate(req.ActivityDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format activity_date tidak valid, gunakan YYYY-MM-DD"})
		return
	}

	activity.Title = req.Title
	activity.Description = req.Description
	activity.Division = req.Division
	activity.IsPinned = req.IsPinned
	activity.ActivityDate = activityDate

	if req.Image != nil {
		imageURL, err := saveImageFile(c, req.Image)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		activity.Image = imageURL
	}

	if err := config.DB.Save(&activity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update activity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Activity berhasil diupdate",
		"data":    activity,
	})
}

// DeleteActivity - hapus activity berdasarkan slug
func DeleteActivity(c *gin.Context) {
	slug := c.Param("slug")

	var activity models.Activity
	if err := config.DB.Where("slug = ?", slug).First(&activity).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Activity tidak ditemukan"})
		return
	}

	if err := config.DB.Delete(&activity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus activity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Activity berhasil dihapus"})
}