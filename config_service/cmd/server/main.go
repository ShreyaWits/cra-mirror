package main

import (
	"nps-config-service/internal/app"
	"nps-config-service/internal/configs/db"
	"nps-config-service/internal/configs"
	"os"
)

func main() {

	//env file loading
	configs.LoadConfig()
	
	DbPort := os.Getenv("DATABASE_URL")
	db.ConnectDatabase(DbPort)
	
	//setup routes
	appInstance := app.SetupRouter()

	appInstance.Listen(":" + configs.AppConfig.Port)

}
