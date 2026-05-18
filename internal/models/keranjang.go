package models

// Keranjang merepresentasikan entitas KERANJANG di ERD
type Keranjang struct {
	ID          uint   `gorm:"primaryKey;column:id" json:"id"`
	PelangganID uint   `gorm:"column:pelanggan_id" json:"pelanggan_id"`
	
	Pelanggan        Pelanggan         `gorm:"foreignKey:PelangganID;references:ID" json:"pelanggan"`
	KeranjangDetails []KeranjangDetail `gorm:"foreignKey:KeranjangID;references:ID" json:"items"`
}

func (Keranjang) TableName() string {
	return "keranjang"
}

// KeranjangDetail merepresentasikan entitas KERANJANG_DETAIL di ERD
type KeranjangDetail struct {
	ID          uint `gorm:"primaryKey;column:id" json:"id"`
	KeranjangID uint `gorm:"column:keranjang_id" json:"keranjang_id"`
	ProductID   uint `gorm:"column:produk_id" json:"product_id"`
	Jumlah      int  `gorm:"column:jumlah" json:"qty"`
	
	Product     Product `gorm:"foreignKey:ProductID;references:ID" json:"product"`
}

func (KeranjangDetail) TableName() string {
	return "keranjang_detail"
}
