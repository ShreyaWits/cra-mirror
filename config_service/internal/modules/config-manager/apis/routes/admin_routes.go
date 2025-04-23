package routes

import (
	"nps-config-service/internal/common/middleware"
	handler "nps-config-service/internal/modules/config-manager/apis/handlers"

	"github.com/gofiber/fiber/v2"
)

func RegisterAdminRoutes(router fiber.Router, h *handler.Handler) {
	router.Post("/admin/login", middleware.SetContextDataMiddleware,h.AdminHandler)
}
