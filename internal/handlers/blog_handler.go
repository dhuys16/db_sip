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

// BlogRequest dipakai untuk body Create & Update (multipart/form-data,
// karena gambar dikirim langsung sebagai file dalam request yang sama)
type BlogRequest struct {
	Title      string                `form:"title" binding:"required"`
	Content    string                `form:"content" binding:"required"`
	Division   string                `form:"division" binding:"required,oneof=lazsip sarsip"`
	Category   string                `form:"category"`    // opsional, bebas isi (misal kategori laporan lapangan SARSIP)
	CampaignID *uint                 `form:"campaign_id"` // opsional, nullable
	IsPinned   bool                  `form:"is_pinned"`
	Image      *multipart.FileHeader `form:"image"` // opsional
}

// generateBlogCode bikin kode unik berformat BL001, BL002, dst.
// Diambil dari angka terbesar pada blog yang terakhir dibuat, supaya tetap
// aman walau ada blog yang sudah dihapus di tengah jalan.
func generateBlogCode() (string, error) {
	var lastBlog models.Blog
	err := config.DB.Order("id desc").First(&lastBlog).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "BL001", nil
		}
		return "", err
	}

	re := regexp.MustCompile(`(\d+)$`)
	match := re.FindStringSubmatch(lastBlog.Slug)

	nextNumber := 1
	if len(match) == 2 {
		if n, convErr := strconv.Atoi(match[1]); convErr == nil {
			nextNumber = n + 1
		}
	}

	return fmt.Sprintf("BL%03d", nextNumber), nil
}

// GetBlogBySlugPublic - detail satu blog untuk publik berdasarkan slug
func GetBlogBySlugPublic(c *gin.Context) {
	slug := c.Param("slug")

	var blog models.Blog
	err := config.DB.Preload("Admin", func(db *gorm.DB) *gorm.DB {
		return db.Select("ID, Name")
	}).Preload("Campaign").Where("slug = ?", slug).First(&blog).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Blog tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": blog})
}

// CreateBlog - buat artikel blog baru. Penulisnya diambil dari admin yang sedang
// login (JWT). Gambar (kalau ada) dikirim langsung sebagai file di field "image".
// Bisa dikaitkan opsional ke sebuah Campaign lewat field "campaign_id".
func CreateBlog(c *gin.Context) {
	var req BlogRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid: " + err.Error()})
		return
	}

	adminIDVal, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin tidak terautentikasi"})
		return
	}
	adminID, ok := adminIDVal.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca data admin"})
		return
	}

	var campaign models.Campaign
	if req.CampaignID != nil {
		if err := config.DB.First(&campaign, *req.CampaignID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Campaign dengan ID tersebut tidak ditemukan"})
			return
		}
	}

	code, err := generateBlogCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat kode blog"})
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

	blog := models.Blog{
		Title:      req.Title,
		Slug:       code,
		Content:    req.Content,
		Image:      imageURL,
		Division:   req.Division,
		Category:   req.Category,
		CampaignID: req.CampaignID,
		IsPinned:   req.IsPinned,
		AdminID:    adminID,
	}

	if err := config.DB.Create(&blog).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan blog"})
		return
	}

	if req.CampaignID != nil {
		blog.Campaign = campaign
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Blog berhasil dibuat",
		"data":    blog,
	})
}

// UpdateBlog - update blog yang sudah ada berdasarkan slug (full update).
// Penulis asli (AdminID) tidak berubah walau yang mengedit adalah admin lain.
// Kalau field "image" tidak dikirim, gambar lama tetap dipakai (tidak dihapus).
func UpdateBlog(c *gin.Context) {
	slug := c.Param("slug")

	var blog models.Blog
	if err := config.DB.Where("slug = ?", slug).First(&blog).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Blog tidak ditemukan"})
		return
	}

	var req BlogRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid: " + err.Error()})
		return
	}

	var campaign models.Campaign
	if req.CampaignID != nil {
		if err := config.DB.First(&campaign, *req.CampaignID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Campaign dengan ID tersebut tidak ditemukan"})
			return
		}
	}

	blog.Title = req.Title
	blog.Content = req.Content
	blog.Division = req.Division
	blog.Category = req.Category
	blog.CampaignID = req.CampaignID
	blog.IsPinned = req.IsPinned

	if req.Image != nil {
		imageURL, err := saveImageFile(c, req.Image)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		blog.Image = imageURL
	}

	if err := config.DB.Save(&blog).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update blog"})
		return
	}

	if req.CampaignID != nil {
		blog.Campaign = campaign
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Blog berhasil diupdate",
		"data":    blog,
	})
}

// DeleteBlog - hapus blog berdasarkan slug
func DeleteBlog(c *gin.Context) {
	slug := c.Param("slug")

	var blog models.Blog
	if err := config.DB.Where("slug = ?", slug).First(&blog).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Blog tidak ditemukan"})
		return
	}

	if err := config.DB.Delete(&blog).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus blog"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Blog berhasil dihapus"})
}