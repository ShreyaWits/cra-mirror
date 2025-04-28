package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"nps-config-service/internal/modules/config-manager/models"
)


var DB *gorm.DB

func ConnectDatabase(port string) {
	dsn := port
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
	database.AutoMigrate(&models.Admin{})
	DB = database
	log.Println("Database connected successfully!")
}

