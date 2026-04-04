package handler

import (
	"food-backend/internal/config"
	"food-backend/internal/models"
	"time"

	"github.com/gin-gonic/gin"
)

// GetDashboardStats mengembalikan statistik untuk halaman dashboard admin.
func GetDashboardStats(c *gin.Context) {
	var totalRevenue float64
	var totalOrders int64
	var totalUsers int64

	config.DB.Model(&models.Order{}).
		Where("status != ?", "Cancelled").
		Select("COALESCE(SUM(total), 0)").
		Scan(&totalRevenue)

	config.DB.Model(&models.Order{}).Count(&totalOrders)

	config.DB.Model(&models.User{}).
		Where("created_at >= ?", time.Now().AddDate(0, 0, -30)).
		Count(&totalUsers)

	c.JSON(200, gin.H{
		"revenue":     totalRevenue,
		"total_order": totalOrders,
		"new_users":   totalUsers,
	})
}
