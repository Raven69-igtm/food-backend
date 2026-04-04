package handler

import (
	"food-backend/internal/config"
	"food-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// GetProfile mengambil data profil user yang sedang login.
func GetProfile(c *gin.Context) {
	id, _ := c.Get("userID")
	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "User tidak ditemukan"})
		return
	}
	c.JSON(200, gin.H{"user": user})
}

// UpdateProfile mengubah data profil user yang sedang login.
func UpdateProfile(c *gin.Context) {
	id, _ := c.Get("userID")
	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "User tidak ditemukan"})
		return
	}
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": "Input tidak valid"})
		return
	}
	config.DB.Save(&user)
	c.JSON(200, gin.H{"user": user})
}

// GetAllUsers mengambil semua data user (admin only).
func GetAllUsers(c *gin.Context) {
	var users []models.User
	config.DB.Find(&users)
	c.JSON(200, gin.H{"data": users})
}
