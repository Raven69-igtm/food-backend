package handler

import (
	"fmt"
	"math/rand"
	"food-backend/internal/config"
	"food-backend/internal/models"
	"strings"

	"github.com/gin-gonic/gin"
)

// CreateOrder membuat order baru. Guest tidak lagi didukung karena skema mewajibkan pelanggan_id.
func CreateOrder(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		userID, exists = c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{"error": "Anda harus login untuk memesan"})
			return
		}
	}

	var input struct {
		JadwalAmbilID uint    `json:"jadwal_ambil_id"`
		Total         float64 `json:"total"`
		MetodeBayar   string  `json:"metode_bayar"`
		Items         []struct {
			ProductID uint    `json:"product_id"`
			Quantity  int     `json:"quantity"`
			Price     float64 `json:"price"`
		} `json:"items"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Input tidak valid"})
		return
	}

	orderRef := fmt.Sprintf("515-%05d", rand.Intn(100000))

	order := models.Order{
		PelangganID: userID.(uint),
		Status:      "pending",
		Total:       input.Total,
		MetodeBayar: input.MetodeBayar,
		OrderRef:    orderRef,
	}

	// Jika jadwal 0 (dari Flutter), simpan sebagai NULL di DB
	if input.JadwalAmbilID != 0 {
		order.JadwalAmbilID = &input.JadwalAmbilID
	} else {
		order.JadwalAmbilID = nil
	}

	tx := config.DB.Begin()

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Gagal membuat order"})
		return
	}

	for _, item := range input.Items {
		orderDetail := models.OrderDetail{
			OrderID:     order.ID,
			ProductID:   item.ProductID,
			Jumlah:      item.Quantity,
			HargaSatuan: item.Price,
		}
		
		if err := tx.Create(&orderDetail).Error; err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": "Gagal membuat detail order"})
			return
		}

		if err := tx.Model(&models.Product{}).
			Where("id = ? AND stok >= ?", item.ProductID, item.Quantity).
			UpdateColumn("stok", config.DB.Raw("stok - ?", item.Quantity)).Error; err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": "Gagal update stok"})
			return
		}
	}

	tx.Commit()

	c.JSON(200, gin.H{
		"message":   "Order berhasil dibuat",
		"order_id":  order.ID,
		"order_ref": order.OrderRef,
	})
}

// GetUserOrders mengambil semua pesanan milik pelanggan yang login.
func GetUserOrders(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		userID, exists = c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			return
		}
	}

	var orders []models.Order
	if err := config.DB.Preload("OrderDetails.Product").Preload("JadwalAmbil").Preload("User").
		Where("pelanggan_id = ?", userID).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal memuat pesanan"})
		return
	}

	c.JSON(200, gin.H{"data": orders})
}

// CancelOrder membatalkan pesanan milik pelanggan (hanya jika status masih pending).
func CancelOrder(c *gin.Context) {
	id := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var order models.Order
	if err := config.DB.First(&order, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Pesanan tidak ditemukan"})
		return
	}

	if order.PelangganID != userID.(uint) {
		c.JSON(403, gin.H{"error": "Bukan pesanan Anda"})
		return
	}

	if !strings.EqualFold(order.Status, "pending") {
		c.JSON(400, gin.H{"error": "Pesanan tidak bisa dibatalkan, status: " + order.Status})
		return
	}

	config.DB.Model(&order).Update("status", "cancelled")

	c.JSON(200, gin.H{"message": "Pesanan berhasil dibatalkan"})
}

// DeleteUserOrder menghapus riwayat pesanan (hanya status completed/cancelled).
func DeleteUserOrder(c *gin.Context) {
	id := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var order models.Order
	if err := config.DB.First(&order, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Pesanan tidak ditemukan"})
		return
	}

	if order.PelangganID != userID.(uint) {
		c.JSON(403, gin.H{"error": "Bukan pesanan Anda"})
		return
	}

	if !strings.EqualFold(order.Status, "completed") &&
		!strings.EqualFold(order.Status, "done") &&
		!strings.EqualFold(order.Status, "cancelled") {
		c.JSON(400, gin.H{"error": "Pesanan aktif tidak bisa dihapus"})
		return
	}

	config.DB.Where("order_id = ?", order.ID).Delete(&models.OrderDetail{})
	config.DB.Delete(&order)

	c.JSON(200, gin.H{"message": "Riwayat pesanan berhasil dihapus"})
}

// DeleteAllUserOrders menghapus seluruh riwayat pesanan (completed/cancelled) milik user.
func DeleteAllUserOrders(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var orders []models.Order
	if err := config.DB.Where("pelanggan_id = ? AND status IN ?", userID, []string{"completed", "done", "cancelled"}).Find(&orders).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal mencari daftar pesanan yang bisa dihapus"})
		return
	}

	if len(orders) == 0 {
		c.JSON(200, gin.H{"message": "Tidak ada riwayat untuk dihapus"})
		return
	}

	var orderIDs []uint
	for _, o := range orders {
		orderIDs = append(orderIDs, o.ID)
	}

	config.DB.Where("order_id IN ?", orderIDs).Delete(&models.OrderDetail{})
	config.DB.Where("id IN ?", orderIDs).Delete(&models.Order{})

	c.JSON(200, gin.H{"message": "Semua riwayat pesanan berhasil dihapus"})
}

// GetAllOrders mengambil semua order (admin only).
func GetAllOrders(c *gin.Context) {
	var orders []models.Order
	config.DB.Preload("Pelanggan").Preload("User").Preload("OrderDetails.Product").Preload("JadwalAmbil").
		Order("created_at DESC").
		Find(&orders)
	c.JSON(200, gin.H{"data": orders})
}

// UpdateOrderStatus mengubah status order (admin only).
func UpdateOrderStatus(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Status        string `json:"status"`
		JadwalAmbilID uint   `json:"jadwal_ambil_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Input tidak valid"})
		return
	}

	var order models.Order
	if err := config.DB.Preload("Pelanggan").Preload("User").First(&order, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Order tidak ditemukan"})
		return
	}

	updates := map[string]interface{}{}
	notifTitle := ""
	notifMsg := ""

	if input.Status != "" {
		updates["status"] = input.Status
		
		// Tentukan pesan notifikasi berdasarkan status baru
		switch strings.ToLower(input.Status) {
		case "processing":
			notifTitle = "Pesanan Diproses 👨‍🍳"
			notifMsg = "Pesanan Anda #" + fmt.Sprint(order.ID) + " sedang kami siapkan. Mohon tunggu ya!"
		case "completed", "done":
			notifTitle = "Pesanan Selesai! 🎁"
			notifMsg = "Pesanan Anda #" + fmt.Sprint(order.ID) + " telah selesai. Terima kasih sudah belanja di roti515!"
		case "cancelled":
			notifTitle = "Pesanan Dibatalkan ❌"
			notifMsg = "Maaf, pesanan Anda #" + fmt.Sprint(order.ID) + " terpaksa kami batalkan."
		}
	}

	if input.JadwalAmbilID != 0 {
		updates["jadwal_ambil_id"] = input.JadwalAmbilID
		notifTitle = "Jadwal Ambil Diatur ⏰"
		notifMsg = "Admin telah mengatur jadwal pengambilan untuk pesanan #" + fmt.Sprint(order.ID) + ". Silakan cek detail pesanan Anda."
	}

	if err := config.DB.Model(&order).Updates(updates).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal mengupdate order"})
		return
	}

	// Kirim Notifikasi ke Pelanggan (jika ada pesan yang didefinisikan)
	if notifTitle != "" && order.PelangganID != 0 {
		notification := models.Notification{
			UserID:  order.PelangganID, // Kirim ke user yang memesan
			Title:   notifTitle,
			Message: notifMsg,
		}
		config.DB.Create(&notification)
	}

	c.JSON(200, gin.H{"message": "Order Updated & Notification Sent"})
}

// AdminDeleteOrder menghapus satu pesanan (admin only).
func AdminDeleteOrder(c *gin.Context) {
	id := c.Param("id")

	var order models.Order
	if err := config.DB.First(&order, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Order tidak ditemukan"})
		return
	}

	config.DB.Where("order_id = ?", order.ID).Delete(&models.OrderDetail{})
	config.DB.Delete(&order)

	c.JSON(200, gin.H{"message": "Order berhasil dihapus oleh admin"})
}

// AdminDeleteFinishedOrders menghapus semua riwayat pesanan yang sudah selesai/dibatalkan (admin only).
func AdminDeleteFinishedOrders(c *gin.Context) {
	var orders []models.Order
	if err := config.DB.Where("status IN ?", []string{"completed", "done", "cancelled"}).Find(&orders).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal mencari daftar pesanan yang bisa dihapus"})
		return
	}

	if len(orders) == 0 {
		c.JSON(200, gin.H{"message": "Tidak ada riwayat pesanan yang bisa dihapus"})
		return
	}

	var orderIDs []uint
	for _, o := range orders {
		orderIDs = append(orderIDs, o.ID)
	}

	config.DB.Where("order_id IN ?", orderIDs).Delete(&models.OrderDetail{})
	config.DB.Where("id IN ?", orderIDs).Delete(&models.Order{})

	c.JSON(200, gin.H{"message": "Semua riwayat pesanan selesai berhasil dihapus"})
}

