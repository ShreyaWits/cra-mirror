package app

import (
	"nps-config-service/internal/modules/config-manager/apis/routes"
	"nps-config-service/internal/modules/config-manager/di"

	"github.com/gofiber/fiber/v2"
)

func RegisterModules(app *fiber.App) {

	//dependencies injection
	configHandler, err := di.InitHandler()
	if err != nil {
		panic("Failed to initialize Config module: " + err.Error())
	}

	api := app.Group("/api/v1/configs")
	//Note: The order of registration matters. The first registered route will be of higher priority. that is why we register the webhook routes first
	routes.RegisterWebHookRoutes(api, configHandler)
	routes.RegisterConfigRoutes(api, configHandler)

}
