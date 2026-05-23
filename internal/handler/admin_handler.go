package handler

import (
	"fmt"
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

	// Total revenue: hanya dari pesanan yang completed/done (bukan cancelled/pending)
	config.DB.Raw(
		`SELECT COALESCE(SUM(total), 0) FROM "order" WHERE LOWER(status) IN ('completed', 'done')`,
	).Scan(&totalRevenue)

	// Total pesanan yang sudah selesai (completed/done)
	config.DB.Raw(`SELECT COUNT(id) FROM "order" WHERE LOWER(status) IN ('completed', 'done')`).Scan(&totalOrders)

	// Total User: hitung dari tabel user utama
	config.DB.Model(&models.User{}).Count(&totalUsers)

	// Pelanggan baru 30 hari terakhir
	var totalNewUsers int64
	config.DB.Model(&models.Pelanggan{}).Where("tgl_daftar >= ?", time.Now().AddDate(0, 0, -30)).Count(&totalNewUsers)

	// --- Growth Mingguan ---
	// Bandingkan revenue minggu ini vs minggu lalu untuk menghitung pertumbuhan
	now := time.Now()
	startThisWeek := now.AddDate(0, 0, -7)
	startLastWeek := now.AddDate(0, 0, -14)

	type WeekRevenue struct {
		Revenue float64
		Orders  int64
	}
	var thisWeek, lastWeek WeekRevenue

	config.DB.Raw(
		`SELECT COALESCE(SUM(total),0) as revenue, COUNT(id) as orders FROM "order" WHERE LOWER(status) IN ('completed','done') AND created_at >= ?`,
		startThisWeek,
	).Scan(&thisWeek)

	config.DB.Raw(
		`SELECT COALESCE(SUM(total),0) as revenue, COUNT(id) as orders FROM "order" WHERE LOWER(status) IN ('completed','done') AND created_at >= ? AND created_at < ?`,
		startLastWeek, startThisWeek,
	).Scan(&lastWeek)

	// Hitung growth persentase
	salesGrowth := calcGrowth(thisWeek.Revenue, lastWeek.Revenue)
	ordersGrowth := calcGrowth(float64(thisWeek.Orders), float64(lastWeek.Orders))

	// Pelanggan baru minggu ini vs minggu lalu
	var usersThisWeek, usersLastWeek int64
	config.DB.Raw("SELECT COUNT(id) FROM pelanggan WHERE tgl_daftar >= ?", startThisWeek).Scan(&usersThisWeek)
	config.DB.Raw("SELECT COUNT(id) FROM pelanggan WHERE tgl_daftar >= ? AND tgl_daftar < ?", startLastWeek, startThisWeek).Scan(&usersLastWeek)
	usersGrowth := calcGrowth(float64(usersThisWeek), float64(usersLastWeek))

	// Data penjualan harian 14 hari terakhir
	type DailyStat struct {
		Date    string  `json:"date"`
		Revenue float64 `json:"revenue"`
	}
	
	// 1. Ambil data asli dari DB
	var dbStats []DailyStat
	config.DB.Table(`"order"`).
		Where("LOWER(status) IN ('completed','done','processing','pending') AND created_at >= ?", now.AddDate(0, 0, -14)).
		Select("DATE(created_at) as date, SUM(total) as revenue").
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&dbStats)

	// 2. Map data asli ke map untuk pencarian cepat
	statsMap := make(map[string]float64)
	for _, s := range dbStats {
		dateKey := s.Date
		if len(dateKey) > 10 {
			dateKey = dateKey[:10]
		}
		statsMap[dateKey] = s.Revenue
	}

	// 3. Isi semua tanggal dalam 14 hari terakhir
	var dailyStats []DailyStat
	for i := 13; i >= 0; i-- {
		d := now.AddDate(0, 0, -i).Format("2006-01-02")
		rev := 0.0
		if val, ok := statsMap[d]; ok {
			rev = val
		}
		dailyStats = append(dailyStats, DailyStat{
			Date:    d,
			Revenue: rev,
		})
	}

	// --- AKTIVITAS TERKINI ---
	type Activity struct {
		Title     string    `json:"title"`
		Subtitle  string    `json:"subtitle"`
		Type      string    `json:"type"`
		CreatedAt time.Time `json:"created_at"`
	}
	var activities []Activity

	// Ambil 5 pesanan terbaru
	var recentOrders []models.Order
	config.DB.Preload("User").Order("created_at DESC").Limit(5).Find(&recentOrders)
	for _, o := range recentOrders {
		name := o.User.Nama
		if name == "" {
			name = "User #" + fmt.Sprint(o.PelangganID)
		}
		
		activities = append(activities, Activity{
			Title:     "Pesanan Baru #" + fmt.Sprint(o.ID),
			Subtitle:  name + " • Rp. " + fmt.Sprint(int(o.Total)),
			Type:      "order",
			CreatedAt: o.CreatedAt,
		})
	}

	// Ambil 5 pendaftaran terbaru (Ambil manual untuk menghindari error relasi)
	var recentPelanggan []models.Pelanggan
	config.DB.Order("tgl_daftar DESC").Limit(5).Find(&recentPelanggan)
	for _, p := range recentPelanggan {
		var u models.User
		config.DB.Select("nama").First(&u, p.ID)
		
		name := u.Nama
		if name == "" {
			name = "Customer #" + fmt.Sprint(p.ID)
		}
		
		activities = append(activities, Activity{
			Title:     "Pelanggan Baru Terdaftar",
			Subtitle:  name,
			Type:      "user",
			CreatedAt: p.TglDaftar,
		})
	}

	c.JSON(200, gin.H{
		"revenue":       totalRevenue,
		"total_order":   totalOrders,
		"new_users":     totalNewUsers,
		"sales_growth":  salesGrowth,
		"orders_growth": ordersGrowth,
		"users_growth":  usersGrowth,
		"daily_stats":   dailyStats,
		"activities":    activities,
	})
}

// calcGrowth menghitung persentase pertumbuhan dari nilai lama ke nilai baru.
func calcGrowth(current, previous float64) string {
	if previous == 0 {
		if current > 0 {
			return "+100%"
		}
		return "+0%"
	}
	pct := ((current - previous) / previous) * 100
	if pct >= 0 {
		return "+" + formatFloat(pct) + "%"
	}
	return formatFloat(pct) + "%"
}

func formatFloat(f float64) string {
	// Bulatkan ke integer untuk tampilan lebih bersih
	i := int(f)
	if i < 0 {
		return "-" + itoa(-i)
	}
	return itoa(i)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	result := ""
	for i > 0 {
		result = string(rune('0'+i%10)) + result
		i /= 10
	}
	return result
}

