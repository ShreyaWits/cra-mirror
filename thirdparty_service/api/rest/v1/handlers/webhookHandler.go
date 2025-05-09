package handlers

import (
	"strings"
	"thirdparty_service/internal/models"
	"thirdparty_service/internal/services"

	"github.com/gofiber/fiber/v2"
)

type WebhookHandler struct {
	service services.WebhookService
}

func NewWebhookHandler(service services.WebhookService) *WebhookHandler {
	return &WebhookHandler{service: service}
}

func (h *WebhookHandler) HandleWebhook(c *fiber.Ctx) error {
	path := strings.TrimPrefix(c.Path(), "/dev/withdrawal")
	key := strings.TrimPrefix(path, "/")

	var payload models.Webhook
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid JSON payload")
	}

	if err := h.service.ProcessWebhook(key, &payload); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to process webhook")
	}

	return c.SendString("Webhook received successfully")
}
