package models

import "time"

// User merepresentasikan entitas dasar USER di ERD
type User struct {
	ID        uint       `gorm:"primaryKey;column:id" json:"id"`
	Nama      string     `gorm:"column:nama" json:"name"`
	Email     string     `gorm:"unique;column:email" json:"email"`
	Password  string     `gorm:"column:password" json:"password"`
	Role      string     `gorm:"column:role;default:pelanggan" json:"role"`
	Admin     *Admin     `gorm:"-" json:"admin,omitempty"`
	Pelanggan *Pelanggan `gorm:"-" json:"pelanggan,omitempty"`
}

func (User) TableName() string {
	return "user"
}

// Admin merepresentasikan entitas ADMIN di ERD
type Admin struct {
	ID   uint  `gorm:"primaryKey;column:id" json:"id"` // PK dan FK ke USER
	User *User `gorm:"-" json:"user,omitempty"`
}

func (Admin) TableName() string {
	return "admin"
}

// Pelanggan merepresentasikan entitas PELANGGAN di ERD
type Pelanggan struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"` // PK dan FK ke USER
	NoHP      string    `gorm:"column:no_hp" json:"phone"`
	TglDaftar time.Time `gorm:"column:tgl_daftar" json:"created_at"`
	User      *User     `gorm:"-" json:"user,omitempty"`
}

func (Pelanggan) TableName() string {
	return "pelanggan"
}
