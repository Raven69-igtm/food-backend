package handler

import (
	"food-backend/internal/config"
	"food-backend/internal/middleware"
	"food-backend/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Login memverifikasi email & password lalu mengembalikan JWT token.
func Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Input tidak valid"})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(401, gin.H{"error": "Email tidak terdaftar"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(401, gin.H{"error": "Password salah"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenStr, err := token.SignedString(middleware.JWTKey)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal membuat token"})
		return
	}

	c.JSON(200, gin.H{
		"token": tokenStr,
		"user":  user,
	})
}

// Register membuat akun user baru dengan role default "user".
func Register(c *gin.Context) {
	var input models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Input tidak valid"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal memproses password"})
		return
	}
	input.Password = string(hash)

	if input.Role == "" {
		input.Role = "user"
	}

	if err := config.DB.Create(&input).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal menyimpan user"})
		return
	}

	c.JSON(200, gin.H{"message": "Registrasi Berhasil"})
}
