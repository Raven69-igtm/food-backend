package handler

import (
	"food-backend/internal/config"
	"food-backend/internal/models"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// GetProfile mengambil data profil user yang sedang login.
func GetProfile(c *gin.Context) {
	id, _ := c.Get("userID")
	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "User tidak ditemukan"})
		return
	}
	c.JSON(200, gin.H{"user": user})
}

// UpdateProfile mengubah data profil user yang sedang login.
// Field yang bisa diubah: name, phone, address, password (opsional).
func UpdateProfile(c *gin.Context) {
	id, _ := c.Get("userID")
	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "User tidak ditemukan"})
		return
	}

	// Gunakan struct input terpisah agar password tidak overwrite langsung
	var input struct {
		Name     string `json:"name"`
		Phone    string `json:"phone"`
		Address  string `json:"address"`
		Password string `json:"password"` // opsional, kosong = tidak ubah password
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Input tidak valid"})
		return
	}

	// Update field profil
	updates := map[string]interface{}{
		"name":    input.Name,
		"phone":   input.Phone,
		"address": input.Address,
	}

	// Hash password baru jika dikirim dan tidak kosong
	if input.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(500, gin.H{"error": "Gagal memproses password"})
			return
		}
		updates["password"] = string(hashed)
	}

	if err := config.DB.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal memperbarui profil"})
		return
	}

	// Reload user terbaru sebelum dikirim sebagai response
	config.DB.First(&user, id)
	c.JSON(200, gin.H{"user": user})
}

// UploadProfilePhoto menangani pengunggahan foto profil user.
func UploadProfilePhoto(c *gin.Context) {
	id, _ := c.Get("userID")

	// 1. Ambil file dari request multipart
	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(400, gin.H{"error": "File tidak ditemukan atau tidak valid"})
		return
	}

	// 2. Generate nama file unik
	filename := time.Now().Format("20060102150405") + "_" + filepath.Base(file.Filename)
	savePath := filepath.Join("cmd/api/uploads", filename)

	// 3. Simpan file ke storage
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(500, gin.H{"error": "Gagal menyimpan berkas ke server"})
		return
	}

	// 4. Update data user di database
	photoURL := "/static/" + filename
	if err := config.DB.Model(&models.User{}).Where("id = ?", id).Update("photo_url", photoURL).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal memperbarui data profil di database"})
		return
	}

	c.JSON(200, gin.H{
		"message":   "Foto profil berhasil diperbarui",
		"photo_url": photoURL,
	})
}

// GetAllUsers mengambil semua data user (admin only).
func GetAllUsers(c *gin.Context) {
	var users []models.User
	config.DB.Find(&users)
	c.JSON(200, gin.H{"data": users})
}
