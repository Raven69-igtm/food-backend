package models

// Product merepresentasikan entitas PRODUK di ERD
type Product struct {
	ID           uint    `gorm:"primaryKey;column:id" json:"id"`
	Nama         string  `gorm:"column:nama" json:"name"`
	Deskripsi    string  `gorm:"column:deskripsi;type:text" json:"description"`
	Harga        int     `gorm:"column:harga" json:"price"`
	Gambar       string  `gorm:"column:gambar" json:"image_url"`
	Rating       float64 `gorm:"column:rating;default:0" json:"rating"`
	Kategori     string  `gorm:"column:kategori" json:"category"`
	IsBestseller bool    `gorm:"column:is_bestseller;default:false" json:"is_bestseller"`
	Stok         int     `gorm:"column:stok" json:"stock"`
	SoldCount    int     `gorm:"column:sold_count;default:0" json:"sold_count"`
}

func (Product) TableName() string {
	return "produk"
}