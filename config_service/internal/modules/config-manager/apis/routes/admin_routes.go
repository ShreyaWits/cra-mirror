package routes

import (
	"nps-config-service/internal/common/middleware"
	handler "nps-config-service/internal/modules/config-manager/apis/handlers"

	"github.com/gofiber/fiber/v2"
)


func RegisterAdminRoutes(router fiber.Router, h *handler.ConfigHandler) {
	router.Post("/admin/login", middleware.SetContextDataAdminMiddleware,h.AdminHandler)
}
