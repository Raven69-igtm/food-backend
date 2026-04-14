package handler

import (
	"food-backend/internal/config"
	"food-backend/internal/models"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateOrder membuat order baru (bisa guest, tidak perlu login).
func CreateOrder(c *gin.Context) {
	// Ambil user_id dari token jika user sudah login (menggunakan middleware)
	// Jika gagal atau tidak ada (guest), userID bernilai nil/0
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

	// Masukkan user_id ke format order jika ada
	order := models.Order{
		UserID:       userIDPtr, // Ini yang akan menautkan ke profil user
		GuestName:    input.GuestName,
		GuestPhone:   input.GuestPhone,
		GuestAddress: input.GuestAddress,
		Total:        input.Total,
		Status:       "Pending",
		CreatedAt:    time.Now(),
	}

	if err := config.DB.Create(&order).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal membuat order"})
		return
	}

	// Loop dan save OrderItems
	for _, item := range input.Items {
		orderItem := models.OrderItem{
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
		config.DB.Create(&orderItem)

		// 📉 KURANGI STOK PRODUK SECARA OTOMATIS
		config.DB.Model(&models.Product{}).
			Where("id = ? AND stock >= ?", item.ProductID, item.Quantity).
			UpdateColumn("stock", config.DB.Raw("stock - ?", item.Quantity))
	}

	c.JSON(200, gin.H{
		"message":  "Order berhasil dibuat",
		"order_id": order.ID,
	})
}

// ==========================================
// FUNGSI BARU DI BAWAH: MENGAMBIL PESANAN MILIK USER TERSEBUT
// ==========================================
func GetUserOrders(c *gin.Context) {
	// Dapatkan user ID yang login dari JWT middleware
	userID, exists := c.Get("userID")
	if !exists {
		userID, exists = c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			return
		}
	}

	var orders []models.Order
	// Cari seluruh order dimana user_id = user yg masuk, load Preload relasinya
	if err := config.DB.Preload("Items.Product").
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&orders).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal memuat pesanan"})
		return
	}

	c.JSON(200, gin.H{"data": orders})
}

// GetAllOrders mengambil semua order (admin only).
func GetAllOrders(c *gin.Context) {
	var orders []models.Order
	config.DB.Preload("User").Preload("Items.Product").Find(&orders)
	c.JSON(200, gin.H{"data": orders})
}

func UpdateOrderStatus(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Status     string  `json:"status"`
		PickupTime *string `json:"pickup_time"` // jam pengambilan
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Input tidak valid"})
		return
	}
	
	// Cari order untuk mengambil UserID
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

	// Buat Notifikasi jika UserID terkait (bukan guest utuh yang tanpa akun)
	if order.UserID != nil {
		if input.PickupTime != nil {
			notif := models.Notification{
				UserID:  *order.UserID,
				Title:   "Pesanan Siap Diambil!",
				Message: "Pesanan #" + id + " Anda sudah bisa diambil hari ini pada pukul " + *input.PickupTime + " WIB.",
			}
			config.DB.Create(&notif)
		} else if input.Status != "" && input.Status != order.Status {
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
