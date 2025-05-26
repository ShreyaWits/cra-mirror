package routes

import (
	"nps-config-service/internal/common/middleware"
	"nps-config-service/internal/common/roles"
	handler "nps-config-service/internal/modules/config-manager/apis/handlers"

	"github.com/gofiber/fiber/v2"
)

func RegisterConfigRoutes(router fiber.Router, h *handler.ConfigHandler) {
	router.Put("/:environment/:service", middleware.RequireRole(roles.RoleAdmin), middleware.SetContextDataMiddleware[map[string]any], h.StoreConfigHandler)
	router.Get("/:environment/:service", middleware.RequireRole(roles.RoleViewer), h.GetfullConfig)
	router.Get("/:environment/:service/:key", middleware.RequireRole(roles.RoleViewer), h.GetByValue)
}
