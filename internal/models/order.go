package models

import "time"

type Order struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       *uint     `json:"user_id"` // nullable agar tamu masih bisa beli
	GuestName    string    `json:"guest_name"`
	GuestPhone   string    `json:"guest_phone"`
	GuestAddress string    `json:"guest_address"`
	Total        float64   `json:"total"`
	Status       string    `json:"status"`
	PickupTime   *string   `json:"pickup_time" gorm:"column:pickup_time"`
	CreatedAt    time.Time `json:"created_at"`

	Items []OrderItem `json:"items" gorm:"foreignKey:OrderID"`
	User  *User       `json:"user" gorm:"foreignKey:UserID"`
}

type OrderItem struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	OrderID   uint    `json:"order_id"`
	ProductID uint    `json:"product_id"`
	Product   Product `json:"food" gorm:"foreignKey:ProductID"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}
