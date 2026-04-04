package models

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name"`
	Email     string    `gorm:"unique" json:"email"`
	Password  string    `json:"password"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	PhotoURL  string    `json:"photo_url"`
	Role      string    `json:"role"` // "admin" atau "user"
	CreatedAt time.Time `json:"created_at"`
}
