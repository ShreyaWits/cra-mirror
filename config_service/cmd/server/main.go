package main

import (
	"nps-config-service/internal/app"
	"nps-config-service/internal/configs"
)

func main() {

	//env file loading
	configs.LoadConfig()

	//setup routes
	appInstance := app.SetupRouter()

	appInstance.Listen(":" + configs.AppConfig.Port)

}
