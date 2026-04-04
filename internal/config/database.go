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

	if err := database.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
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
