package routes

import (
	"nps-config-service/internal/common/middleware"
	handler "nps-config-service/internal/modules/config-manager/apis/handlers"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	"github.com/gofiber/fiber/v2"
)


func RegisterAdminRoutes(router fiber.Router, h *handler.AdminHandler) {
	router.Post("/admin/signup", middleware.SetContextDataMiddleware[dtos.AdminSignupDto],h.CreateAdminHandler)
	router.Post("/admin/login", middleware.SetContextDataMiddleware[dtos.AdminLoginDto],h.FetchAdminHandler)
}
