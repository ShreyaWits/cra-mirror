package routes

import (
	"template-services/internal/template/dto"
	"template-services/internal/template/handler"
	"template-services/internal/template/middleware" // assuming location

	"github.com/gofiber/fiber/v2"
)

func SetupTemplateRoutes(app *fiber.App, h *handler.TemplateHandler) {
	templates := app.Group("v1/templates")

	// Create template with validation
	templates.Post("/", middleware.ValidateBody[dto.CreateTemplateRequest](), h.CreateTemplate)

	// Get template with validation on query
	templates.Get("/", middleware.ValidateQuery[dto.GetTemplateRequestV1](), h.GetTemplate)

	// Update template with validation
	templates.Put("/:id", middleware.ValidateBody[dto.UpdateTemplateRequest](), h.UpdateTemplate)

	// Delete doesn't need validation middleware unless you use a struct
	templates.Delete("/:id", h.DeleteTemplate)
}
