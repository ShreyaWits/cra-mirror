package routes

import (
	handler "nps-config-service/internal/modules/config-manager/apis/handlers"

	"github.com/gofiber/fiber/v2"
)

func RegisterWebHookRoutes(router fiber.Router, h *handler.WebhookHandler) {
	router.Post("/webhook", h.RegisterWebhook)

	// Get all webhooks for a specific environment/service
	router.Get("/:environment/:service/webhooks", h.GetWebhooks)

	// todo: Update a specific webhook (PATCH for partial updates)
	// router.Patch("/webhook/:environment/:service", h.UpdateWebhook)

	router.Delete("/:environment/:service/webhook", h.DeleteWebhook)
}
