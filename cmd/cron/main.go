package main

import (
	"log"
	"time"

	"db_sip/internal/config"
	"db_sip/internal/models"
)

func main() {
	log.Println("Memulai Job Cek Transaksi Pending...")

	// 1. Inisialisasi Database
	config.ConnectDatabase()

	// 2. Tentukan batas waktu transaksi dianggap kadaluarsa
	// Default Midtrans Snap expiry biasanya 24 jam
	expiryTime := time.Now().Add(-24 * time.Hour)

	var pendingTransactions []models.Transaction

	// 3. Cari transaksi dengan status 'pending' yang melewati batas waktu
	result := config.DB.Where("status = ? AND created_at <= ?", "pending", expiryTime).Find(&pendingTransactions)
	if result.Error != nil {
		log.Fatalf("Gagal melakukan query database: %v", result.Error)
	}

	if len(pendingTransactions) == 0 {
		log.Println("Tidak ada transaksi pending yang kadaluarsa.")
		return
	}

	log.Printf("Ditemukan %d transaksi pending lama. Memulai proses update...", len(pendingTransactions))

	// 4. Proses Update Status
	for _, tx := range pendingTransactions {
		log.Printf("Menandai transaksi %s (ID: %d) sebagai expired", tx.MidtransOrderID, tx.ID)
		
		// TODO (Fase Layanan Midtrans): 
		// Idealnya di sini kita melakukan request GET ke API Status Midtrans (https://api.midtrans.com/v2/{order_id}/status)
		// untuk memastikan bahwa user benar-benar tidak membayar, menghindari race condition jika webhook gagal masuk.
		
		tx.Status = "expired"
		if err := config.DB.Save(&tx).Error; err != nil {
			log.Printf("Gagal update transaksi %d: %v", tx.ID, err)
		}
	}

	log.Println("Job Cek Transaksi selesai dieksekusi.")
}