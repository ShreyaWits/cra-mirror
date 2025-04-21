package routes

import (
	"encryption_microservice/internal/modules/encryption/api/handlers"

	"github.com/gofiber/fiber/v2"
)

// SetupAPIRoutes sets up all the API routes for the application
func SetupAPIRoutes(app *fiber.App, handler handlers.EncryptionHandler) {
	// Health check
	app.Get("/health", handler.HandleHealth)

	// Encryption routes
	api := app.Group("/encryption")
	api.Post("/encrypt", handler.HandleEncrypt)
	api.Post("/decrypt", handler.HandleDecrypt)
	api.Post("/generate-edek", handler.HandleGenerateEDEK)
}
