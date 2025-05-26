package app

import (
	"nps-config-service/internal/common/middleware"
	"nps-config-service/internal/modules/config-manager/apis/routes"
	"nps-config-service/internal/modules/config-manager/di"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// SetupRouter configures the application routes and middleware
func SetupRouter(app *fiber.App, container *di.Container) *fiber.App {
	if app == nil {
		app = fiber.New()
	}

	// Add recovery middleware
	var ConfigDefault = recover.Config{
		Next:              nil,
		EnableStackTrace:  true,
		StackTraceHandler: recover.ConfigDefault.StackTraceHandler,
	}
	app.Use(recover.New(ConfigDefault))

	// Register modules with the container
	registerModules(app, container)
	return app
}

// registerModules registers all application modules with their routes
func registerModules(app *fiber.App, container *di.Container) {
	// Get handlers from container
	adminHandler, configHandler, webhookHandler, err := container.GetHandlers()
	if err != nil {
		panic("Failed to get handlers from container: " + err.Error())
	}

	api := app.Group("/api/v1/")
	protected := api.Group("/config", middleware.JWTMiddleware())

	routes.RegisterWebHookRoutes(protected, webhookHandler)
	routes.RegisterAdminRoutes(api, adminHandler)
	routes.RegisterConfigRoutes(protected, configHandler)
}
