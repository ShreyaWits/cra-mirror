package routes

import (
	"messaging_service/internal/modules/message_broker/handler"

	"github.com/gofiber/fiber/v2"
)

func RegisterConfigRoutes(app *fiber.App, config *handler.ConfigHandler) {
	app.Post("/config", config.UpdateConfigurations)
	// app.Post("/config/development/audit", configHandler.HandleEnvConfigWebhook) // WILL BE REMOVED ACCORDING TO CHANGES IN CONFIG SERVICE
	app.Get("/health", config.Health)
}
