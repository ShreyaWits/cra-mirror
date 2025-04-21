package routes

import (
	commomiddleware "encryption_microservice/internal/common/middleware"
	"encryption_microservice/internal/modules/encryption/api/dtos"
	"encryption_microservice/internal/modules/encryption/api/handlers"

	"github.com/gofiber/fiber/v2"
)

// SetupAPIRoutes sets up all the API routes for the application
func SetupAPIRoutes(app *fiber.App, handler handlers.EncryptionHandler) {
	// Health check
	app.Get("/health", handler.HandleHealth)

	// Encryption routes
	api := app.Group("/encryption")
	api.Post("/encrypt", commomiddleware.TokenValidator(), commomiddleware.ValidateParams[dtos.EncryptRequest](), handler.HandleEncrypt)
	api.Post("/decrypt", commomiddleware.TokenValidator(), commomiddleware.ValidateParams[dtos.DecryptRequest](), handler.HandleDecrypt)
	api.Get("/generate-edek", commomiddleware.TokenValidator(), handler.HandleGenerateEDEK)
}
