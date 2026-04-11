package handler

import (
	"food-backend/internal/config"
	"food-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// GetUserNotifications mengambil history notifikasi untuk user (dari middleware auth)
func GetUserNotifications(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	notifications := []models.Notification{}
	// Tampilkan yang terbaru di atas
	if err := config.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&notifications).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal mengambil notifikasi"})
		return
	}

	c.JSON(200, notifications)
}

// MarkNotificationRead mengubah status is_read menjadi true
func MarkNotificationRead(c *gin.Context) {
	id := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	if err := config.DB.Model(&models.Notification{}).Where("id = ? AND user_id = ?", id, userID).Update("is_read", true).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal membaca notifikasi"})
		return
	}

	c.JSON(200, gin.H{"message": "Notifikasi ditandai dibaca"})
}
