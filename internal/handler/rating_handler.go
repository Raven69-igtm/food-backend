package handler

import (
	"food-backend/internal/config"
	"food-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// AddRating menambahkan rating dari user untuk produk tertentu.
// Endpoint: POST /api/foods/:id/rating
func AddRating(c *gin.Context) {
	foodID := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var input struct {
		Rating  float64 `json:"rating" binding:"required,min=1,max=5"`
		OrderID uint    `json:"order_id"`
		Comment string  `json:"comment"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Rating harus antara 1–5"})
		return
	}

	// Cari produk
	var product models.Product
	if err := config.DB.First(&product, foodID).Error; err != nil {
		c.JSON(404, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	// Cek apakah user sudah pernah rating produk ini untuk order yang sama
	if input.OrderID > 0 {
		var existingCount int64
		config.DB.Model(&models.Rating{}).
			Where("user_id = ? AND product_id = ? AND order_id = ?", userID, product.ID, input.OrderID).
			Count(&existingCount)
		if existingCount > 0 {
			c.JSON(400, gin.H{"error": "Anda sudah memberi rating untuk pesanan ini"})
			return
		}
	}

	// Simpan rating baru
	rating := models.Rating{
		UserID:    userID.(uint),
		ProductID: product.ID,
		OrderID:   input.OrderID,
		Rating:    input.Rating,
		Comment:   input.Comment,
	}
	if err := config.DB.Create(&rating).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal menyimpan rating"})
		return
	}

	// Update rata-rata rating produk
	var avgRating float64
	config.DB.Model(&models.Rating{}).
		Where("product_id = ?", product.ID).
		Select("COALESCE(AVG(rating), 0)").
		Scan(&avgRating)

	config.DB.Model(&product).Update("rating", avgRating)

	c.JSON(201, gin.H{
		"message": "Rating berhasil ditambahkan",
		"rating":  input.Rating,
		"average": avgRating,
	})
}
