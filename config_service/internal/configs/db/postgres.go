package db

import (
	"fmt"
	"log"
	"nps-config-service/internal/configs"
	"nps-config-service/internal/modules/config-manager/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase(config *configs.Config) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s",
		configs.AppConfig.DatabaseHost, config.DatabasePort, config.DatabaseUser, config.DatabasePassword, config.DatabaseName)
	// dsn := port
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
	database.AutoMigrate(&models.Admin{})
	DB = database
	log.Println("Database connected successfully!")
}
