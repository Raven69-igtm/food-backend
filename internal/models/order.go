package models

import "time"

type Order struct {
	ID           uint        `gorm:"primaryKey" json:"id"`
	UserID       uint        `json:"user_id"`
	GuestName    string      `json:"guest_name"`
	GuestPhone   string      `json:"guest_phone"`
	GuestAddress string      `json:"guest_address"`
	Total        float64     `json:"total"`
	Status       string      `json:"status"`
	Items        []OrderItem `json:"items" gorm:"foreignKey:OrderID"`
	CreatedAt    time.Time   `json:"created_at"`
}

type OrderItem struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	OrderID   uint    `json:"order_id"`
	ProductID uint    `json:"product_id"`
	Product   Product `json:"product" gorm:"foreignKey:ProductID"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}
