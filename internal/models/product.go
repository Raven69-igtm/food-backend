package models // <-- Pastikan baris ini ada di paling atas!

type Product struct {
	ID           uint    `gorm:"primaryKey" json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Price        int     `json:"price"`
	ImageURL     string  `json:"image_url"`
	Rating       float64 `json:"rating"`
	Category     string  `json:"category"`
	IsBestseller bool    `json:"is_bestseller"`
	Stock        int     `json:"stock"`
}

func (Product) TableName() string {
	return "product"
}