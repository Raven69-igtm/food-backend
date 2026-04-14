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

	// Mengambil data penjualan harian selama 14 hari terakhir
	type DailyStat struct {
		Date    string  `json:"date"`
		Revenue float64 `json:"revenue"`
	}
	var dailyStats []DailyStat

	// Query untuk mengambil total revenue per hari
	config.DB.Model(&models.Order{}).
		Where("status != ? AND created_at >= ?", "Cancelled", time.Now().AddDate(0, 0, -14)).
		Select("DATE(created_at) as date, SUM(total) as revenue").
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&dailyStats)

	c.JSON(200, gin.H{
		"revenue":     totalRevenue,
		"total_order": totalOrders,
		"new_users":   totalUsers,
		"daily_stats": dailyStats,
	})
}
