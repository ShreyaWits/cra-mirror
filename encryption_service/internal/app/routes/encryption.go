package routes

import (
	"encryption_microservice/internal/common/middleware"
	"encryption_microservice/internal/modules/encryption/api/dtos"
	"encryption_microservice/internal/modules/encryption/api/handlers"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes sets up all the routes for the encryption service
func SetupEncryptionRoutes(app *fiber.App, handler handlers.EncryptionHandler) {
	// Create API group

	// Encryption routes
	api := app.Group("/encryption")
	api.Post("/encrypt", middleware.TokenValidator(), middleware.ValidateParams[dtos.EncryptRequest](), handler.HandleEncrypt)
	api.Post("/decrypt", middleware.TokenValidator(), middleware.ValidateParams[dtos.DecryptRequest](), handler.HandleDecrypt)
	api.Get("/generate-edek", middleware.TokenValidator(), handler.HandleGenerateEDEK)
}
