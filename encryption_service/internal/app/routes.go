package app

import (
	"encryption_microservice/internal/modules/encryption/api/handlers"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// SetupRoutes configures the Fiber application with all routes
func SetupRoutes(app *fiber.App, encryptionHandler handlers.EncryptionHandler, configHandler *handlers.ConfigHandler) {
	// Add middleware
	app.Use(recover.New())
	app.Use(logger.New())

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "OK",
		})
	})

	// Configuration endpoints
	if configHandler != nil {
		app.Post("/config", configHandler.UpdateConfigurations)
	} else {
		fmt.Println("Warning: Config handler is nil, config routes will not be available")
	}

	// Add more routes here if needed
}
