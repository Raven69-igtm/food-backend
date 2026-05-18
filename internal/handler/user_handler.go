package handler

import (
	"food-backend/internal/config"
	"food-backend/internal/models"

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
	
	response := gin.H{
		"id":    user.ID,
		"name":  user.Nama,
		"email": user.Email,
		"role":  user.Role,
	}

	if user.Role == "pelanggan" {
		var pelanggan models.Pelanggan
		if err := config.DB.Where("id = ?", user.ID).First(&pelanggan).Error; err == nil {
			response["phone"] = pelanggan.NoHP
			response["created_at"] = pelanggan.TglDaftar
		}
	} else if user.Role == "admin" {
		response["is_admin"] = true
	}
	
	c.JSON(200, gin.H{"user": response})
}

// UpdateProfile mengubah data profil user yang sedang login.
func UpdateProfile(c *gin.Context) {
	id, _ := c.Get("userID")
	roleContext, _ := c.Get("userRole")
	
	var input struct {
		Name     string `json:"name"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Input tidak valid"})
		return
	}

	tx := config.DB.Begin()

	var user models.User
	if err := tx.First(&user, id).Error; err != nil {
		tx.Rollback()
		c.JSON(404, gin.H{"error": "User tidak ditemukan"})
		return
	}

	if input.Name != "" {
		user.Nama = input.Name
	}
	
	if input.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": "Gagal memproses password"})
			return
		}
		user.Password = string(hashed)
	}

	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Gagal memperbarui user"})
		return
	}

	response := gin.H{
		"id":    user.ID,
		"name":  user.Nama,
		"email": user.Email,
		"role":  user.Role,
	}

	// Update data spesifik pelanggan jika rolenya pelanggan
	if roleContext == "pelanggan" && input.Phone != "" {
		var pelanggan models.Pelanggan
		if err := tx.First(&pelanggan, id).Error; err == nil {
			pelanggan.NoHP = input.Phone
			if err := tx.Save(&pelanggan).Error; err != nil {
				tx.Rollback()
				c.JSON(500, gin.H{"error": "Gagal memperbarui data tambahan pelanggan"})
				return
			}
			response["phone"] = pelanggan.NoHP
			response["created_at"] = pelanggan.TglDaftar
		}
	}

	tx.Commit()

	c.JSON(200, gin.H{"user": response})
}

// UploadProfilePhoto tidak didukung karena kolom gambar tidak ada di ERD untuk tabel User/Pelanggan.
func UploadProfilePhoto(c *gin.Context) {
	c.JSON(400, gin.H{"error": "Fitur ini telah dinonaktifkan sesuai ERD"})
}

// GetAllUsers mengambil semua data user beserta perannya (admin only).
func GetAllUsers(c *gin.Context) {
	var users []models.User
	config.DB.Find(&users)
	c.JSON(200, gin.H{"data": users})
}

