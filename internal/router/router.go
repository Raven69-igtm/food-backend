package router

import (
	"strings"
	"time"

	"food-backend/internal/handler"
	"food-backend/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Setup menginisialisasi router Gin beserta semua middleware dan route.
func Setup() *gin.Engine {
	r := gin.Default()

	// CORS: mengizinkan semua port dari localhost dan Railway
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return strings.HasPrefix(origin, "http://localhost") ||
				strings.HasPrefix(origin, "http://127.0.0.1") ||
				strings.Contains(origin, "railway.app") ||
				strings.Contains(origin, "up.railway.app")
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Static("/static", "cmd/api/uploads")

	// --- PUBLIC ROUTES ---
	r.POST("/api/register", handler.Register)
	r.POST("/api/login", handler.Login)
	r.GET("/api/foods", handler.GetProducts)
	r.POST("/api/orders", middleware.OptionalAuth(), handler.CreateOrder)

	// --- PROTECTED ROUTES (Butuh Login) ---
	api := r.Group("/api")
	api.Use(middleware.Auth())
	{
		// Profil
		api.GET("/profile", handler.GetProfile)
		api.PUT("/profile", handler.UpdateProfile)

		// Pesanan user
		api.GET("/orders", handler.GetUserOrders)
		api.GET("/user/orders", handler.GetUserOrders)
		api.PUT("/orders/:id/cancel", handler.CancelOrder)   // Batalkan pesanan (user)
		api.DELETE("/orders/:id", handler.DeleteUserOrder)   // Hapus riwayat pesanan (user)

		// Notifikasi
		api.GET("/notifications", handler.GetUserNotifications)
		api.PUT("/notifications/:id/read", handler.MarkNotificationRead)
		api.DELETE("/notifications/:id", handler.DeleteNotification) // Hapus notifikasi (user)

		// Rating produk
		api.POST("/foods/:id/rating", handler.AddRating)

		// AI
		api.POST("/ask-ai", handler.AskAI)

		// --- KHUSUS ADMIN ROUTES ---
		admin := api.Group("/admin")
		admin.Use(middleware.AdminOnly())
		{
			admin.GET("/stats", handler.GetDashboardStats)
			admin.GET("/orders", handler.GetAllOrders)
			admin.PUT("/orders/:id", handler.UpdateOrderStatus)
			admin.GET("/users", handler.GetAllUsers)

			admin.POST("/foods", handler.CreateProduct)
			admin.PUT("/foods/:id", handler.UpdateProduct)
			admin.DELETE("/foods/:id", handler.DeleteProduct)
		}
	}

	return r
}
