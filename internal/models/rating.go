package models

// Rating menyimpan penilaian bintang dari user untuk produk tertentu.
type Rating struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	UserID    uint    `json:"user_id"`
	ProductID uint    `json:"product_id"`
	OrderID   uint    `json:"order_id"`
	Rating    float64 `json:"rating"`
	Comment   string  `json:"comment"`

	User    User    `json:"user" gorm:"foreignKey:UserID"`
	Product Product `json:"product" gorm:"foreignKey:ProductID"`
}
