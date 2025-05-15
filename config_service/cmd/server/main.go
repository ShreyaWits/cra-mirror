package main

import (
	"log"
	"nps-config-service/internal/app"
	"nps-config-service/internal/configs"
	"nps-config-service/internal/configs/db"
)

func main() {

	//env file loading
	err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// DbPort := os.Getenv("DATABASE_URL")
	db.ConnectDatabase()

	//setup routes
	appInstance := app.SetupRouter()

	appInstance.Listen(":" + configs.AppConfig.Port)

}
