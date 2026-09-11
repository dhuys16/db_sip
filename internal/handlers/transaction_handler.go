package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"db_sip/internal/config"
	"db_sip/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DonationRequest dipakai buat bikin donasi baru (belum ada payment gateway,
// jadi ini murni bikin record transaksi dengan status "pending")
type DonationRequest struct {
	DonorName       string  `json:"donor_name" binding:"required"`
	DonorEmail      string  `json:"donor_email"`
	DonorPhone      string  `json:"donor_phone"`
	IsAnonymous     bool    `json:"is_anonymous"`
	CampaignID      *uint   `json:"campaign_id"` // opsional, nullable buat donasi/zakat umum
	TransactionType string  `json:"transaction_type" binding:"omitempty,oneof=donasi zakat infaq"`
	GrossAmount     float64 `json:"gross_amount" binding:"required,gt=0"`
	Notes           string  `json:"notes"`
}

// findOrCreateDonor cari Donor berdasarkan email (kalau diisi), atau bikin baru
// kalau belum pernah donasi sebelumnya / emailnya kosong (donasi anonim/offline).
func findOrCreateDonor(req DonationRequest) (models.Donor, error) {
	var donor models.Donor

	if req.DonorEmail != "" {
		err := config.DB.Where("email = ?", req.DonorEmail).First(&donor).Error
		if err == nil {
			return donor, nil // donatur lama, dipakai ulang datanya
		}
	}

	donor = models.Donor{
		Name:        req.DonorName,
		Email:       req.DonorEmail,
		Phone:       req.DonorPhone,
		IsAnonymous: req.IsAnonymous,
	}
	if err := config.DB.Create(&donor).Error; err != nil {
		return donor, err
	}
	return donor, nil
}

// CreateDonation - bikin transaksi donasi baru (status awal selalu "pending").
// Endpoint publik, dipanggil pas donatur pertama kali submit form donasi.
// Karena belum ada payment gateway, tidak ada biaya_admin/potongan yang dihitung
// (net_amount = gross_amount), dan payment_type diisi "simulasi".
func CreateDonation(c *gin.Context) {
	var req DonationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid: " + err.Error()})
		return
	}

	if req.TransactionType == "" {
		req.TransactionType = "donasi"
	}

	var campaign models.Campaign
	if req.CampaignID != nil {
		if err := config.DB.Preload("Program").First(&campaign, *req.CampaignID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Campaign dengan ID tersebut tidak ditemukan"})
			return
		}
	}

	donor, err := findOrCreateDonor(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan data donatur"})
		return
	}

	// Placeholder order ID, biasanya ini yang dikirim ke Midtrans. Sementara
	// dibikin sendiri format "SIP-<timestamp>" biar tetap unik.
	orderID := fmt.Sprintf("SIP-%d", time.Now().UnixNano())

	transaction := models.Transaction{
		MidtransOrderID: orderID,
		CampaignID:      req.CampaignID,
		DonorID:         donor.ID,
		TransactionType: req.TransactionType,
		GrossAmount:     req.GrossAmount,
		BiayaAdmin:      0,
		NetAmount:       req.GrossAmount,
		PaymentType:     "simulasi",
		Status:          "pending",
		Notes:           req.Notes,
	}

	if err := config.DB.Create(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat transaksi"})
		return
	}

	transaction.Donor = donor
	if req.CampaignID != nil {
		transaction.Campaign = campaign
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Donasi berhasil dibuat, menunggu pembayaran (simulasi)",
		"data":    transaction,
	})
}

// GetDonationStatusPublic - cek status transaksi berdasarkan order ID.
// Dipakai donatur buat ngecek "donasi saya udah kebayar belum".
func GetDonationStatusPublic(c *gin.Context) {
	orderID := c.Param("order_id")

	var transaction models.Transaction
	if err := config.DB.Preload("Donor").Preload("Campaign.Program").
		Where("midtrans_order_id = ?", orderID).First(&transaction).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaksi tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": transaction})
}

// GetTransactionsAdmin - list semua transaksi buat kebutuhan admin.
// Bisa difilter ?status= dan ?transaction_type=
func GetTransactionsAdmin(c *gin.Context) {
	var transactions []models.Transaction
	status := c.Query("status")
	transactionType := c.Query("transaction_type")

	query := config.DB.Preload("Donor").Preload("Campaign.Program").Order("created_at desc")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if transactionType != "" {
		query = query.Where("transaction_type = ?", transactionType)
	}

	if err := query.Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data transaksi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": transactions})
}

// SimulatePaymentSuccess - TEMPORARY: dipakai buat simulasi transaksi jadi "paid"
// selama payment gateway asli belum diintegrasikan. Nanti kalau Midtrans udah
// jalan, endpoint ini digantikan webhook resmi yang diverifikasi signature-nya,
// bukan dipanggil manual kayak sekarang. Otomatis nambah current_amount Campaign
// terkait kalau transaksinya nempel ke sebuah Campaign.
func SimulatePaymentSuccess(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID transaksi tidak valid"})
		return
	}

	var transaction models.Transaction
	if err := config.DB.Preload("Donor").Preload("Campaign.Program").First(&transaction, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaksi tidak ditemukan"})
		return
	}

	if transaction.Status == "paid" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaksi ini sudah berstatus paid sebelumnya"})
		return
	}

	now := time.Now()
	transaction.Status = "paid"
	transaction.PaidAt = &now

	if err := config.DB.Save(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update status transaksi"})
		return
	}

	// Tambah current_amount Campaign secara atomik (aman dari race condition)
	if transaction.CampaignID != nil {
		config.DB.Model(&models.Campaign{}).
			Where("id = ?", *transaction.CampaignID).
			Update("current_amount", gorm.Expr("current_amount + ?", transaction.NetAmount))
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pembayaran disimulasikan berhasil (paid)",
		"data":    transaction,
	})
}

// SimulatePaymentFailed - TEMPORARY: simulasi transaksi gagal/dibatalkan.
// Sama kayak SimulatePaymentSuccess, ini bakal digantikan webhook asli nanti.
func SimulatePaymentFailed(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID transaksi tidak valid"})
		return
	}

	var transaction models.Transaction
	if err := config.DB.Preload("Donor").Preload("Campaign.Program").First(&transaction, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaksi tidak ditemukan"})
		return
	}

	if transaction.Status == "paid" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaksi yang sudah paid tidak bisa diubah jadi failed"})
		return
	}

	transaction.Status = "failed"

	if err := config.DB.Save(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update status transaksi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pembayaran disimulasikan gagal (failed)",
		"data":    transaction,
	})
}