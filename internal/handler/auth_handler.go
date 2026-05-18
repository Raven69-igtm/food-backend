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

	role := user.Role

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"role": role,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenStr, err := token.SignedString(middleware.JWTKey)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal membuat token"})
		return
	}

	c.JSON(200, gin.H{
		"token": tokenStr,
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Nama,
			"email": user.Email,
			"role":  role,
		},
	})
}

// Register membuat akun pelanggan baru
func Register(c *gin.Context) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Phone    string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Input tidak valid"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal memproses password"})
		return
	}

	tx := config.DB.Begin()

	user := models.User{
		Nama:     input.Name,
		Email:    input.Email,
		Password: string(hash),
		Role:     "pelanggan",
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Gagal menyimpan user"})
		return
	}

	pelanggan := models.Pelanggan{
		ID:        user.ID,
		NoHP:      input.Phone,
		TglDaftar: time.Now(),
	}

	if err := tx.Create(&pelanggan).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Gagal menyimpan data pelanggan"})
		return
	}

	tx.Commit()
	c.JSON(200, gin.H{"message": "Registrasi Berhasil"})
}

// ForgotPassword memverifikasi identitas user berdasarkan Nama, Email, dan No HP.
func ForgotPassword(c *gin.Context) {
	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Input tidak valid"})
		return
	}

	var user models.User
	err := config.DB.Table("user").
		Joins("JOIN pelanggan ON pelanggan.id = user.id").
		Where("user.nama = ? AND user.email = ? AND pelanggan.no_hp = ?", input.Name, input.Email, input.Phone).
		Select("user.*").
		First(&user).Error
	if err != nil {
		c.JSON(401, gin.H{"error": "Identitas tidak ditemukan. Pastikan Nama, Email, dan No. HP sesuai."})
		return
	}

	c.JSON(200, gin.H{
		"message": "Identitas terverifikasi",
		"email":   user.Email,
	})
}

// ResetPassword memperbarui password user setelah identitas diverifikasi.
func ResetPassword(c *gin.Context) {
	var input struct {
		Email       string `json:"email"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Input tidak valid"})
		return
	}

	if len(input.NewPassword) < 6 {
		c.JSON(400, gin.H{"error": "Password minimal 6 karakter"})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"error": "User tidak ditemukan"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal memproses password baru"})
		return
	}

	user.Password = string(hash)
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal memperbarui password"})
		return
	}

	c.JSON(200, gin.H{"message": "Password berhasil diperbarui, silakan login kembali"})
}

