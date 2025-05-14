package routes

import (
	"template-services/internal/template/dto"
	"template-services/internal/template/handler"
	"template-services/internal/template/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupTemplateRoutes(app *fiber.App, h *handler.TemplateHandler) {
	templates := app.Group("v1/templates")

	// Create template
	templates.Post("/create", middleware.ValidateBody[dto.CreateTemplateRequest](), h.CreateTemplate)

	// Get template
	templates.Post("/get", middleware.ValidateBody[dto.GetTemplateRequest](), h.GetTemplate)

	// Update template
	templates.Post("/update/", middleware.ValidateBody[dto.UpdateTemplateRequest](), h.UpdateTemplate)

	// Delete template
	templates.Post("/delete/", middleware.ValidateBody[dto.DeleteTemplateRequest](), h.DeleteTemplate)

}
