package api

import (
	"net/http"
	"food-backend/internal/config"
	"food-backend/internal/router"
	"github.com/gin-gonic/gin"
)

var app *gin.Engine

func init() {
	// Inisialisasi database sebelum router berjalan
	config.ConnectDatabase()
	
	// Setup router Gin
	app = router.Setup()
}

// Handler adalah entrypoint wajib untuk Vercel Serverless Go
func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
