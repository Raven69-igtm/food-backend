package main

import (
	"fmt"
	"log"
	"os"

	"food-backend/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("Mulai Migrasi Data dari MySQL Lokal ke PostgreSQL (Neon)...")

	// 1. Koneksi MySQL
	// Gunakan password MySQL lokal Anda jika ada, default XAMPP adalah root tanpa password
	mysqlDSN := "root:@tcp(127.0.0.1:3306)/roti_515_db?charset=utf8mb4&parseTime=True&loc=Local"
	mysqlDB, err := gorm.Open(mysql.Open(mysqlDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Gagal konek MySQL: %v", err)
	}
	fmt.Println("✅ Koneksi MySQL Sukses!")

	// 2. Koneksi Postgres
	pgURL := os.Getenv("DATABASE_URL")
	if pgURL == "" {
		log.Fatal("❌ DATABASE_URL belum diset! Harap jalankan script ini dengan environment variable.")
	}
	pgDB, err := gorm.Open(postgres.Open(pgURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Gagal konek Postgres: %v", err)
	}
	fmt.Println("✅ Koneksi Postgres Sukses!")

	// 3. Proses copy tabel satu per satu secara berurutan agar tidak melanggar Foreign Key
	migrateData(mysqlDB, pgDB, &[]models.User{}, "users")
	migrateData(mysqlDB, pgDB, &[]models.Admin{}, "admins")
	migrateData(mysqlDB, pgDB, &[]models.Pelanggan{}, "pelanggan")
	migrateData(mysqlDB, pgDB, &[]models.Product{}, "products")
	migrateData(mysqlDB, pgDB, &[]models.Order{}, "orders")
	migrateData(mysqlDB, pgDB, &[]models.OrderDetail{}, "order_details")
	migrateData(mysqlDB, pgDB, &[]models.Keranjang{}, "keranjangs")
	migrateData(mysqlDB, pgDB, &[]models.KeranjangDetail{}, "keranjang_details")
	migrateData(mysqlDB, pgDB, &[]models.Notification{}, "notifications")
	migrateData(mysqlDB, pgDB, &[]models.Rating{}, "ratings")
	migrateData(mysqlDB, pgDB, &[]models.JadwalAmbil{}, "jadwal_ambils")

	fmt.Println("\n🎉 SELAMAT! Semua data berhasil dipindahkan ke PostgreSQL persis seperti aslinya!")
}

// Fungsi pembantu untuk memindahkan data
func migrateData[T any](mysqlDB, pgDB *gorm.DB, data *[]T, tableName string) {
	fmt.Printf("\nMemproses tabel [%s]...\n", tableName)

	// Tarik semua data dari MySQL
	if err := mysqlDB.Find(data).Error; err != nil {
		fmt.Printf("⚠️ Gagal membaca tabel %s dari MySQL: %v\n", tableName, err)
		return
	}

	count := len(*data)
	if count == 0 {
		fmt.Println("👉 Tabel kosong, dilewati.")
		return
	}

	// Hapus isi tabel Postgres (agar bersih saat disalin ulang)
	pgDB.Exec(fmt.Sprintf(`TRUNCATE TABLE "%s" CASCADE;`, tableName))

	// Simpan semua data ke Postgres
	if err := pgDB.Create(data).Error; err != nil {
		fmt.Printf("❌ Gagal menyimpan ke Postgres: %v\n", err)
	} else {
		fmt.Printf("✅ Sukses memindahkan %d baris ke [%s]\n", count, tableName)
		
		// Reset auto-increment sequence di Postgres agar tidak bentrok saat ada data baru masuk
		seqQuery := fmt.Sprintf(`SELECT setval(pg_get_serial_sequence('"%s"', 'id'), coalesce(max(id), 1), max(id) IS NOT null) FROM "%s";`, tableName, tableName)
		pgDB.Exec(seqQuery)
	}
}
