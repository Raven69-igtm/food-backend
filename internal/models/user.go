package models

import "time"

type User struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Name           string     `json:"name"`
	Email          string     `gorm:"unique" json:"email"`
	Password       string     `json:"password"`
	Phone          string     `json:"phone"`
	Address        string     `json:"address"`
	PhotoURL       string     `json:"photo_url"`
	Role           string     `json:"role"` // "admin" atau "user"
	ResetOTP       string     `json:"reset_otp"`
	ResetOTPExpiry *time.Time `json:"reset_otp_expiry"`
	CreatedAt      time.Time  `json:"created_at"`
}
