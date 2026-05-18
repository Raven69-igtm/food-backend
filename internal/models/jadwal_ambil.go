package models

// JadwalAmbil merepresentasikan entitas JADWAL_AMBIL di ERD
type JadwalAmbil struct {
	ID         uint   `gorm:"primaryKey;column:id" json:"id"`
	JamMulai   string `gorm:"type:time;column:jam_mulai" json:"jam_mulai"`
	JamSelesai string `gorm:"type:time;column:jam_selesai" json:"jam_selesai"`
	IsAktif    bool   `gorm:"column:is_aktif" json:"is_aktif"`
}

func (JadwalAmbil) TableName() string {
	return "jadwal_ambil"
}
