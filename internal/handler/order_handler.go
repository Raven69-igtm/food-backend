package handler

import (
	"food-backend/internal/config"
	"food-backend/internal/models"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateOrder membuat order baru (bisa guest, tidak perlu login).
func CreateOrder(c *gin.Context) {
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

	for _, item := range input.Items {
		config.DB.Create(&models.OrderItem{
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}

	c.JSON(200, gin.H{"order_id": order.ID})
}

// GetUserOrders mengambil daftar order milik user yang sedang login.
func GetUserOrders(c *gin.Context) {
	id, _ := c.Get("userID")
	var orders []models.Order
	config.DB.Preload("Items.Product").Where("user_id = ?", id).Find(&orders)
	c.JSON(200, gin.H{"data": orders})
}

// GetAllOrders mengambil semua order (admin only).
func GetAllOrders(c *gin.Context) {
	var orders []models.Order
	config.DB.Preload("Items.Product").Find(&orders)
	c.JSON(200, gin.H{"data": orders})
}

// UpdateOrderStatus mengubah status order berdasarkan ID (admin only).
func UpdateOrderStatus(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Input tidak valid"})
		return
	}
	config.DB.Model(&models.Order{}).Where("id = ?", id).Update("status", input.Status)
	c.JSON(200, gin.H{"message": "Status Updated"})
}
