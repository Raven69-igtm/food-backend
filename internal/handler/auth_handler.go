package handler

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"food-backend/internal/config"
	"food-backend/internal/middleware"
	"food-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
	"gopkg.in/gomail.v2"
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

// GoogleLogin memverifikasi ID Token dari Google dan membuatkan/mengembalikan JWT Roti 515.
func GoogleLogin(c *gin.Context) {
	var input struct {
		IDToken string `json:"id_token"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "ID Token tidak ditemukan"})
		return
	}

	// Ganti dengan Web Client ID dari Google Cloud Console Anda
	clientID := "492758926071-hmumclc4o1vh96fdfup6s66ij8fr6pvr.apps.googleusercontent.com"

	// 1. Verifikasi ID Token menggunakan package resmi Google
	payload, err := idtoken.Validate(context.Background(), input.IDToken, clientID)
	if err != nil {
		c.JSON(401, gin.H{"error": "Token Google tidak valid atau kadaluarsa", "details": err.Error()})
		return
	}

	// Ambil email dan nama dari payload token Google
	email := payload.Claims["email"].(string)
	name := payload.Claims["name"].(string)

	// 2. Cek apakah user sudah terdaftar di database
	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		// Jika belum ada, otomatis buatkan akun baru (Register by Google)
		user = models.User{
			Name:     name,
			Email:    email,
			Password: "", // User yang login dengan Google tidak punya password lokal default
			Role:     "user",
		}
		if err := config.DB.Create(&user).Error; err != nil {
			c.JSON(500, gin.H{"error": "Gagal mendaftarkan pengguna baru via Google"})
			return
		}
	}

	// 3. Buatkan JWT standar aplikasi Roti 515
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenStr, err := token.SignedString(middleware.JWTKey)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal membuat token aplikasi"})
		return
	}

	c.JSON(200, gin.H{
		"token": tokenStr,
		"user":  user,
	})
}

// ForgotPassword mengirimkan kode OTP 6 digit ke email user.
func ForgotPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Input tidak valid"})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"error": "Email tidak terdaftar"})
		return
	}

	// 1. Generate 6-digit OTP
	otp := generateOTP(6)
	expiry := time.Now().Add(10 * time.Minute)

	// 2. Simpan OTP ke database
	user.ResetOTP = otp
	user.ResetOTPExpiry = &expiry
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal menyimpan kode verifikasi"})
		return
	}

	// 3. Kirim Email secara asynchronous agar tidak menghambat response
	go func() {
		err := sendOTPEmail(user.Email, otp)
		if err != nil {
			log.Printf("Gagal mengirim email ke %s: %v", user.Email, err)
		}
	}()

	c.JSON(200, gin.H{
		"message": "Kode verifikasi telah dikirim ke email Anda",
		"email":   user.Email,
	})
}

// ResetPassword memverifikasi OTP dan memperbarui password user.
func ResetPassword(c *gin.Context) {
	var input struct {
		Email       string `json:"email"`
		OTP         string `json:"otp"`
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

	// 1. Verifikasi OTP
	if user.ResetOTP == "" || user.ResetOTP != input.OTP {
		c.JSON(401, gin.H{"error": "Kode verifikasi salah"})
		return
	}

	// 2. Cek Kadaluarsa
	if user.ResetOTPExpiry == nil || time.Now().After(*user.ResetOTPExpiry) {
		c.JSON(401, gin.H{"error": "Kode verifikasi telah kadaluarsa"})
		return
	}

	// 3. Hash Password Baru
	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal memproses password baru"})
		return
	}

	// 4. Update User (Hapus OTP setelah digunakan)
	user.Password = string(hash)
	user.ResetOTP = ""
	user.ResetOTPExpiry = nil
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal memperbarui password"})
		return
	}

	c.JSON(200, gin.H{"message": "Password berhasil diperbarui, silakan login kembali"})
}

// --- HELPER FUNCTIONS ---

func generateOTP(max int) string {
	var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max || err != nil {
		// Fallback jika crypto/rand gagal
		return "123456"
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}

func sendOTPEmail(targetEmail, otp string) error {
	m := gomail.NewMessage()

	// Ambil dari .env
	from := os.Getenv("SMTP_EMAIL")
	pass := os.Getenv("SMTP_PASSWORD")
	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	port, _ := strconv.Atoi(portStr)

	if from == "" || pass == "" {
		return fmt.Errorf("konfigurasi SMTP belum lengkap di .env")
	}

	m.SetHeader("From", from)
	m.SetHeader("To", targetEmail)
	m.SetHeader("Subject", "Kode Verifikasi Roti 515")
	m.SetBody("text/html", fmt.Sprintf(`
		<div style="font-family: sans-serif; max-width: 400px; margin: auto; padding: 20px; border: 1px solid #eee; border-radius: 10px;">
			<h2 style="color: #D47311; text-align: center;">Roti 515</h2>
			<p>Halo,</p>
			<p>Kami menerima permintaan untuk mereset kata sandi Anda. Gunakan kode verifikasi di bawah ini untuk melanjutkan:</p>
			<div style="background: #fdf2e9; padding: 15px; text-align: center; font-size: 24px; font-weight: bold; color: #D47311; letter-spacing: 5px; border-radius: 5px;">
				%s
			</div>
			<p style="font-size: 12px; color: #777; margin-top: 20px;">
				Kode ini akan kadaluarsa dalam 10 menit. Jika Anda tidak merasa melakukan permintaan ini, abaikan saja email ini.
			</p>
		</div>
	`, otp))

	d := gomail.NewDialer(host, port, from, pass)

	return d.DialAndSend(m)
}
