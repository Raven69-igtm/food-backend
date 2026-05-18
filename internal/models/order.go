package models

import "time"

// Order merepresentasikan entitas ORDER di ERD
type Order struct {
	ID            uint         `gorm:"primaryKey;column:id" json:"id"`
	PelangganID   uint         `gorm:"column:pelanggan_id" json:"pelanggan_id"`
	JadwalAmbilID *uint        `gorm:"column:jadwal_ambil_id" json:"jadwal_ambil_id"`
	Status        string       `gorm:"column:status" json:"status"`
	Total         float64      `gorm:"column:total" json:"total"`
	MetodeBayar   string       `gorm:"column:metode_bayar" json:"metode_bayar"`
	OrderRef      string       `gorm:"column:order_ref;type:varchar(50)" json:"order_ref"`
	CreatedAt     time.Time    `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	Pelanggan    Pelanggan    `gorm:"foreignKey:PelangganID;references:ID" json:"pelanggan"`
	User         User         `gorm:"foreignKey:PelangganID;references:ID" json:"user"`
	JadwalAmbil  JadwalAmbil  `gorm:"foreignKey:JadwalAmbilID;references:ID" json:"jadwal_ambil"`
	OrderDetails []OrderDetail `gorm:"foreignKey:OrderID;references:ID" json:"items"`
}

func (Order) TableName() string {
	return "order"
}


// OrderDetail merepresentasikan entitas ORDER_DETAIL di ERD
type OrderDetail struct {
	ID          uint    `gorm:"primaryKey;column:id" json:"id"`
	OrderID     uint    `gorm:"column:order_id" json:"order_id"`
	ProductID   uint    `gorm:"column:produk_id" json:"product_id"`
	Jumlah      int     `gorm:"column:jumlah" json:"quantity"`
	HargaSatuan float64 `gorm:"column:harga_satuan" json:"price"`
	
	Product     Product `gorm:"foreignKey:ProductID;references:ID" json:"product"`
}

func (OrderDetail) TableName() string {
	return "order_detail"
}
