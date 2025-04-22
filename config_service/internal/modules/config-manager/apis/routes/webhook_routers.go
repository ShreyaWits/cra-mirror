package routes

import (
	handler "nps-config-service/internal/modules/config-manager/apis/handlers"

	"github.com/gofiber/fiber/v2"
)

func RegisterWebHookRoutes(router fiber.Router, h *handler.Handler) {
	router.Post("/register-webhook", h.RegisterWebhook)
}
