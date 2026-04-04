package main

import (
	"food-backend/internal/config"
	"food-backend/internal/router"
	"log"
)

func main() {
	// Inisialisasi koneksi ke Database (Aiven)
	config.ConnectDatabase()

	// Setup Router (Gin)
	r := router.Setup()

	// Jalankan Server
	log.Println("server running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
