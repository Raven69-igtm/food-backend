package handler

import (
	"food-backend/internal/config"
	"food-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// GetProducts mengambil daftar produk dengan filter opsional search dan category.
func GetProducts(c *gin.Context) {
	search := c.Query("search")
	category := c.Query("category")

	db := config.DB.Model(&models.Product{})

	if category != "" {
		db = db.Where("LOWER(kategori) = LOWER(?)", category)
	}

	if search != "" {
		db = db.Where("LOWER(nama) LIKE LOWER(?)", "%"+search+"%")
	}

	var products []models.Product
	db.Find(&products)
	c.JSON(200, gin.H{"data": products})
}

// CreateProduct menambahkan produk baru (admin only).
func CreateProduct(c *gin.Context) {
	var input models.Product
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Input tidak valid"})
		return
	}
	if err := config.DB.Create(&input).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal menyimpan produk"})
		return
	}
	c.JSON(200, gin.H{"data": input})
}

// UpdateProduct mengubah data produk berdasarkan ID (admin only).
func UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product
	if err := config.DB.First(&product, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Produk tidak ditemukan"})
		return
	}
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(400, gin.H{"error": "Input tidak valid"})
		return
	}
	config.DB.Save(&product)
	c.JSON(200, gin.H{"data": product})
}

// DeleteProduct menghapus produk berdasarkan ID (admin only).
func DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	if err := config.DB.Delete(&models.Product{}, id).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal menghapus produk"})
		return
	}
	c.JSON(200, gin.H{"message": "Produk dihapus"})
}
