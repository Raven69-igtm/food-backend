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

// DemoLogin adalah endpoint khusus untuk mode demo (Google Login Simulasi).
// Langsung masuk sebagai akun demo tanpa verifikasi password.
func DemoLogin(c *gin.Context) {
	const demoEmail = "jmk48@gmail.com"

	var user models.User
	if err := config.DB.Where("email = ?", demoEmail).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"error": "Akun demo tidak ditemukan di database"})
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


// ForgotPassword memverifikasi identitas user (nama, email, no HP) sebelum mengizinkan reset password.
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

	if input.Name == "" || input.Email == "" || input.Phone == "" {
		c.JSON(400, gin.H{"error": "Nama, email, dan nomor HP wajib diisi"})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"error": "Email tidak ditemukan"})
		return
	}

	// Verifikasi nama dan nomor HP
	if user.Name != input.Name || user.Phone != input.Phone {
		c.JSON(401, gin.H{"error": "Data tidak cocok. Periksa kembali nama dan nomor HP Anda."})
		return
	}

	c.JSON(200, gin.H{"message": "Identitas terverifikasi", "email": user.Email})
}

// ResetPassword memperbarui password user setelah identitas berhasil diverifikasi.
func ResetPassword(c *gin.Context) {
	var input struct {
		Email       string `json:"email"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Input tidak valid"})
		return
	}

	if input.Email == "" || input.NewPassword == "" {
		c.JSON(400, gin.H{"error": "Email dan password baru wajib diisi"})
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

	// Hash password baru
	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal memproses password"})
		return
	}

	// Update password di database
	if err := config.DB.Model(&user).Update("password", string(hash)).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal memperbarui password"})
		return
	}

	c.JSON(200, gin.H{"message": "Password berhasil diperbarui"})
}

