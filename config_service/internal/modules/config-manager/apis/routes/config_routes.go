package routes

import (
	handler "nps-config-service/internal/modules/config-manager/apis/handlers"

	"github.com/gofiber/fiber/v2"
)

func RegisterConfigRoutes(router fiber.Router, h *handler.ConfigHandler) {
	router.Put("/:environment/:service", h.StoreConfigHandler)
	router.Get("/:environment/:service", h.GetfullConfig)
	router.Get("/:environment/:service/:key", h.GetByValue)
	router.Get("/:environment/:service/metadata", h.GetByMetadata)
}
