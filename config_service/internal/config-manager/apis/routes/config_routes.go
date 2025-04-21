package routes

import (
	handler "nps-config-service/internal/config-manager/apis/handlers"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all the routes for the config service
func SetupRoutes(app *fiber.App) {
	config := app.Group("/api/v1/config")

	config.Put("/:environment/:service", handler.StoreConfigHandler)
	config.Get("/:environment/:service", handler.GetfullConfig)
	// // Get specific config value
	config.Get("/:environment/:service/:key", handler.GetByValue)
	// // Get config metadata
	config.Get("/:environment/:service/metadata", handler.GetByMetadata)
} 