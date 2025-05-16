package routes

import (
	
	"template-services/internal/template/dto"
	"template-services/internal/template/handler"
	"template-services/internal/template/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupTemplateRoutes(app *fiber.App, h *handler.TemplateHandler) {

	// Register routes
	v1 := app.Group("/v1/templates")
	v1.Post("/create", middleware.ValidateBody[dto.CreateTemplateRequest](), h.CreateTemplate)
	v1.Post("/get", middleware.ValidateBody[dto.GetTemplateRequest](), h.GetTemplate)
	v1.Post("/update", middleware.ValidateBody[dto.UpdateTemplateRequest](), h.UpdateTemplate)
	v1.Post("/delete", middleware.ValidateBody[dto.DeleteTemplateRequest](), h.DeleteTemplate)
}
