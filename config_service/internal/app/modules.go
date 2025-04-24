package app

import (
	"nps-config-service/internal/common/middleware"
	"nps-config-service/internal/modules/config-manager/apis/routes"

	"nps-config-service/internal/modules/config-manager/di"

	"github.com/gofiber/fiber/v2"
)

func RegisterModules(app *fiber.App) {
	container, err := di.NewContainer()
	if err != nil {
		panic("Failed to initialize container: " + err.Error())
	}

	configHandler, webhookHandler, err := di.InitHandlers(container)
	if err != nil {
		panic("Failed to initialize handlers: " + err.Error())
	}

	api := app.Group("/api/v1/")
	protected := api.Group("/config", middleware.JWTMiddleware())

	routes.RegisterWebHookRoutes(protected, webhookHandler)
	routes.RegisterAdminRoutes(api, configHandler)
	routes.RegisterConfigRoutes(protected, configHandler)
}
