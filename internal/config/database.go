package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"food-backend/internal/models"

	"gorm.io/driver/mysql"
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

	database, err := gorm.Open(mysql.Open(dsn), gormCfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Perbaikan Data Lama:
	// Karena sebelumnya UserID adalah `uint` (tidak nullable, default 0),
	// kita perlu mengubah `user_id` yang 0 menjadi NULL agar tidak bentrok dengan Foreign Key pengguna
	//database.Exec("ALTER TABLE orders MODIFY user_id bigint unsigned NULL;")
	//database.Exec("UPDATE orders SET user_id = NULL WHERE user_id = 0;")
	// Hapus constraint jika sudah terlanjur bermasalah
	//database.Exec("ALTER TABLE orders DROP FOREIGN KEY fk_orders_user;") // Akan di-ignore MySQL jika tidak eksis

	if err := database.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
		&models.Notification{},
	); err != nil {
		log.Fatalf("auto-migration failed: %v", err)
	}

	fmt.Println("database connection established successfully")
	DB = database
}

// buildDSN membangun string koneksi MySQL dari environment variable.
func buildDSN() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "3306")
	user := getEnv("DB_USER", "root")
	password := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DB_NAME", "roti_515_db")

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=skip-verify",
		user, password, host, port, dbname,
	)
}

// getEnv mengembalikan nilai env variable, atau fallback jika tidak ada.
func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
