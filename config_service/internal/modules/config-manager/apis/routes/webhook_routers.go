package routes

import (
	"nps-config-service/internal/common/middleware"
	"nps-config-service/internal/common/roles"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	handler "nps-config-service/internal/modules/config-manager/apis/handlers"
	"nps-config-service/internal/modules/config-manager/apis/middlewares"

	"github.com/gofiber/fiber/v2"
)

func RegisterWebHookRoutes(router fiber.Router, h *handler.WebhookHandler) {
	router.Post("/webhook", middleware.SetContextDataMiddleware[dtos.RegisterWebhookRequest], middlewares.ValidateBody(new(dtos.RegisterWebhookRequest)), middleware.RequireRole(roles.RoleAdmin), h.RegisterWebhook)

	// Get all webhooks for a specific environment/service
	router.Get("/:environment/:service/webhooks", h.GetWebhooks)

	// todo: Update a specific webhook (PATCH for partial updates)
	// router.Patch("/webhook/:environment/:service", h.UpdateWebhook)

	router.Delete("/:environment/:service/webhook", middleware.SetContextDataMiddleware[dtos.RegisterWebhookRequest], middleware.RequireRole(roles.RoleViewer), h.DeleteWebhook)
}
