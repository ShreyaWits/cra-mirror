package db

import (
	"fmt"
	"log"
	"template-services/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type YugabyteDB struct {
	*gorm.DB
}

func NewYugabyteDB(HOST, PORT, USER, PASSWORD, DB string) *YugabyteDB {

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		HOST,
		PORT,
		USER,
		PASSWORD,
		DB,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to YugabyteDB: %v", err)
	}

	// Auto migrate the schema
	if err := db.AutoMigrate(&models.Template{}); err != nil {
		log.Fatalf("Failed to migrate schema: %v", err)
	}

	return &YugabyteDB{db}
}
