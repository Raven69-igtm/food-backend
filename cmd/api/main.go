package main

import (
	"food-backend/internal/config"
	"food-backend/internal/router"
	"log"
	"os"
)

func main() {
	// Inisialisasi koneksi ke Database (Aiven)
	config.ConnectDatabase()

	// Setup Router (Gin)
	r := router.Setup()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Jalankan Server
	log.Printf("server running on :%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
