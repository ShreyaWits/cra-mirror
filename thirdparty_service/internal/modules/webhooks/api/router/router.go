package router

import (
	"thirdparty_service/internal/app"
	handlers "thirdparty_service/internal/modules/webhooks/api/handler"

	"github.com/gofiber/fiber/v2"
)

func SetupWebhookRoutes(router fiber.Router) {
	webhookPrefix := router.Group("/webhook")

	webhookService := app.GetWebhookService()

	webhookHandlers := handlers.NewWebhookHandler(webhookService)

	webhookPrefix.Post("/", webhookHandlers.HandleWebhook)
}
