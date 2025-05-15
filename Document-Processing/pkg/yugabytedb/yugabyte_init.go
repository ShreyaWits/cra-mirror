package yugabytedb

import (
	"Document-Processing/internal/models"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "yugabyte"),
		getEnv("DB_PASSWORD", "yugabyte"),
		getEnv("DB_NAME", "yugabyte"),
		getEnv("DB_PORT", "5434"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to YugabyteDB: %w", err)
	}

	if err := db.AutoMigrate(&models.DocumentData{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Successfully connected to YugabyteDB and ran migrations.")
	return db, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
