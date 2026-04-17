package handler

import (
	"food-backend/internal/config"
	"food-backend/internal/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateOrder membuat order baru (bisa guest, tidak perlu login).
func CreateOrder(c *gin.Context) {
	var userIDPtr *uint
	if val, exists := c.Get("userID"); exists {
		if uid, ok := val.(uint); ok {
			userIDPtr = &uid
		}
	} else if val, exists := c.Get("user_id"); exists {
		if uid, ok := val.(uint); ok {
			userIDPtr = &uid
		}
	}

	var input struct {
		GuestName    string  `json:"guest_name"`
		GuestPhone   string  `json:"guest_phone"`
		GuestAddress string  `json:"guest_address"`
		Total        float64 `json:"total"`
		Items        []struct {
			ProductID uint    `json:"product_id"`
			Quantity  int     `json:"quantity"`
			Price     float64 `json:"price"`
		} `json:"items"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Input tidak valid"})
		return
	}

	order := models.Order{
		UserID:       userIDPtr,
		GuestName:    input.GuestName,
		GuestPhone:   input.GuestPhone,
		GuestAddress: input.GuestAddress,
		Total:        input.Total,
		Status:       "pending",
		CreatedAt:    time.Now(),
	}

	if err := config.DB.Create(&order).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal membuat order"})
		return
	}

	for _, item := range input.Items {
		orderItem := models.OrderItem{
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
		config.DB.Create(&orderItem)

		config.DB.Model(&models.Product{}).
			Where("id = ? AND stock >= ?", item.ProductID, item.Quantity).
			UpdateColumn("stock", config.DB.Raw("stock - ?", item.Quantity))
	}

	c.JSON(200, gin.H{
		"message":  "Order berhasil dibuat",
		"order_id": order.ID,
	})
}

// GetUserOrders mengambil semua pesanan milik user yang login.
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
	if err := config.DB.Preload("Items.Product").
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&orders).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal memuat pesanan"})
		return
	}

	c.JSON(200, gin.H{"data": orders})
}

// CancelOrder membatalkan pesanan milik user (hanya jika status masih pending).
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

	// Hanya pemilik pesanan yang bisa membatalkan
	if order.UserID == nil || *order.UserID != userID.(uint) {
		c.JSON(403, gin.H{"error": "Bukan pesanan Anda"})
		return
	}

	// Hanya boleh batalkan jika masih pending
	if !strings.EqualFold(order.Status, "pending") {
		c.JSON(400, gin.H{"error": "Pesanan tidak bisa dibatalkan, status: " + order.Status})
		return
	}

	config.DB.Model(&order).Update("status", "cancelled")

	// Kirim notifikasi ke user
	notif := models.Notification{
		UserID:  *order.UserID,
		Title:   "Pesanan Dibatalkan",
		Message: "Pesanan #" + id + " berhasil dibatalkan.",
	}
	config.DB.Create(&notif)

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

	// Hanya pemilik pesanan
	if order.UserID == nil || *order.UserID != userID.(uint) {
		c.JSON(403, gin.H{"error": "Bukan pesanan Anda"})
		return
	}

	// Hanya pesanan selesai / dibatalkan yang boleh dihapus
	if !strings.EqualFold(order.Status, "completed") && 
	   !strings.EqualFold(order.Status, "done") && 
	   !strings.EqualFold(order.Status, "cancelled") {
		c.JSON(400, gin.H{"error": "Pesanan aktif tidak bisa dihapus"})
		return
	}

	// Hapus items dulu (foreign key), lalu hapus order
	config.DB.Where("order_id = ?", order.ID).Delete(&models.OrderItem{})
	config.DB.Delete(&order)

	c.JSON(200, gin.H{"message": "Riwayat pesanan berhasil dihapus"})
}

// GetAllOrders mengambil semua order (admin only).
func GetAllOrders(c *gin.Context) {
	var orders []models.Order
	config.DB.Preload("User").Preload("Items.Product").
		Order("created_at desc").
		Find(&orders)
	c.JSON(200, gin.H{"data": orders})
}

// UpdateOrderStatus mengubah status order (admin only).
func UpdateOrderStatus(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Status     string  `json:"status"`
		PickupTime *string `json:"pickup_time"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Input tidak valid"})
		return
	}

	var order models.Order
	if err := config.DB.First(&order, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Order tidak ditemukan"})
		return
	}

	updates := map[string]interface{}{}
	if input.Status != "" {
		updates["status"] = input.Status
	}
	if input.PickupTime != nil {
		updates["pickup_time"] = *input.PickupTime
	}
	config.DB.Model(&order).Updates(updates)

	// Buat Notifikasi jika user terdaftar (non-guest)
	if order.UserID != nil {
		if input.PickupTime != nil {
			notif := models.Notification{
				UserID:  *order.UserID,
				Title:   "Pesanan Siap Diambil!",
				Message: "Pesanan #" + id + " Anda sudah bisa diambil hari ini pada pukul " + *input.PickupTime + " WIB.",
			}
			config.DB.Create(&notif)
		} else if input.Status != "" && !strings.EqualFold(input.Status, order.Status) {
			notif := models.Notification{
				UserID:  *order.UserID,
				Title:   "Status Pesanan Diperbarui",
				Message: "Status pesanan #" + id + " berubah menjadi: " + input.Status,
			}
			config.DB.Create(&notif)
		}
	}

	c.JSON(200, gin.H{"message": "Order Updated"})
}
