package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"food-backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB adalah instance database global.
var DB *gorm.DB

// ConnectDatabase menginisialisasi koneksi MySQL dan menjalankan AutoMigrate.
func ConnectDatabase() {
	dsn := buildDSN()

	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	}

	// Jika di Render, kita akan memakai DATABASE_URL langsung dari environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		dsn = dbURL
	}

	database, err := gorm.Open(postgres.Open(dsn), gormCfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Perbaikan Data Lama:
	// Karena sebelumnya UserID adalah `uint` (tidak nullable, default 0),
	///kita perlu mengubah `user_id` yang 0 menjadi NULL agar tidak bentrok dengan Foreign Key pengguna
	///database.Exec("ALTER TABLE orders MODIFY user_id bigint unsigned NULL;")
	// database.Exec("UPDATE orders SET user_id = NULL WHERE user_id = 0;")
	// Hapus constraint jika sudah terlanjur bermasalah
	// database.Exec("ALTER TABLE orders DROP FOREIGN KEY fk_orders_user;") 
	
	// Backfill created_at untuk data lama agar muncul di grafik
	// Menggunakan '2000-01-01' sebagai batas bawah
	// Postgres menggunakan tanda kutip ganda untuk identitas yang bentrok (seperti nama tabel order -> orders)
	database.Exec(`UPDATE "orders" SET created_at = NOW() WHERE created_at IS NULL OR created_at < '2000-01-01'`)


	if err := database.AutoMigrate(
		&models.User{},
		&models.Admin{},
		&models.Pelanggan{},
		&models.Product{},
		&models.JadwalAmbil{},
		&models.Keranjang{},
		&models.KeranjangDetail{},
		&models.Order{},
		&models.OrderDetail{},
		&models.Notification{},
		&models.Rating{},
	); err != nil {
		log.Fatalf("auto-migration failed: %v", err)
	}

	// Sinkronisasi data lama: Set role 'admin' jika user ada di tabel admin
	database.Exec(`UPDATE "users" SET role = 'admin' WHERE id IN (SELECT id FROM admins)`)

	fmt.Println("database connection established successfully")
	DB = database
}

// buildDSN membangun string koneksi MySQL dari environment variable.
func buildDSN() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432") // default port postgres
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DB_NAME", "roti_515_db")

	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		host, user, password, dbname, port,
	)
}

// getEnv mengembalikan nilai env variable, atau fallback jika tidak ada.
func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
