package handlers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// allowedImageExtensions daftar ekstensi gambar yang diizinkan
var allowedImageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

const maxUploadSize = 5 << 20 // 5 MB

// saveImageFile memvalidasi & menyimpan file gambar ke folder ./uploads,
// lalu mengembalikan URL absolut yang langsung siap dipakai/disimpan ke database.
// Dipakai bareng oleh endpoint upload mandiri (UploadImage) maupun
// Create/Update Blog & Activity yang menerima gambar langsung dalam satu request.
func saveImageFile(c *gin.Context, file *multipart.FileHeader) (string, error) {
	if file.Size > maxUploadSize {
		return "", fmt.Errorf("ukuran file maksimal 5MB")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExtensions[ext] {
		return "", fmt.Errorf("format file tidak didukung, gunakan jpg/jpeg/png, atau webp")
	}

	if err := os.MkdirAll("uploads", 0755); err != nil {
		return "", fmt.Errorf("gagal menyiapkan folder upload")
	}

	// Nama file unik (timestamp nano detik) supaya tidak saling menimpa antar-upload
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join("uploads", filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		return "", fmt.Errorf("gagal menyimpan file")
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/uploads/%s", scheme, c.Request.Host, filename), nil
}

// UploadImage - endpoint upload mandiri (field form-data: "image").
// Tetap dipertahankan buat kebutuhan ganti gambar tanpa harus submit ulang
// seluruh form Blog/Activity.
func UploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File 'image' tidak ditemukan di form-data"})
		return
	}

	imageURL, err := saveImageFile(c, file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Upload berhasil",
		"image":   imageURL,
	})
}