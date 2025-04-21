package routes

import (
	"encryption_microservice/internal/modules/encryption/api/handlers"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes sets up all the routes for the encryption service
func SetupRoutes(app *fiber.App, handler handlers.EncryptionHandler) {
	// Create API group
	api := app.Group("/api/v1")

	// Encryption routes
	encryption := api.Group("/encryption")
	encryption.Post("/encrypt", handler.HandleEncrypt)
	encryption.Post("/decrypt", handler.HandleDecrypt)
	encryption.Post("/generate-edek", handler.HandleGenerateEDEK)
}
