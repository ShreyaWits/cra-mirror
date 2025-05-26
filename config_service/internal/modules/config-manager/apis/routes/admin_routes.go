package routes

import (
	"nps-config-service/internal/common/middleware"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
	handler "nps-config-service/internal/modules/config-manager/apis/handlers"
	"nps-config-service/internal/modules/config-manager/apis/middlewares"

	"github.com/gofiber/fiber/v2"
)

func RegisterAdminRoutes(router fiber.Router, h *handler.AdminHandler) {
	// Public routes (no auth required)
	router.Post("/admin/signup",
		middleware.SetContextDataMiddleware[dtos.AdminSignupDto],
		middlewares.ValidateBody(new(dtos.AdminSignupDto)),
		h.CreateAdminHandler)

	router.Post("/admin/login",
		middleware.SetContextDataMiddleware[dtos.AdminLoginDto],
		h.FetchAdminHandler)

	// Protected routes (require JWT)
	adminGroup := router.Group("/admin", middleware.JWTMiddleware())

	// Admin only routes
	adminGroup.Get("/admins",
		middleware.RequireRole("ADMIN"),
		h.ListAdminsHandler)

	adminGroup.Delete("/admins/:username",
		middleware.RequireRole("ADMIN"),
		h.DeleteAdminHandler)
}
