package routes

import (
	"template-services/internal/template/handler"

	"github.com/gofiber/fiber/v2"
)

// SetupTemplateRoutes configures all template-related routes
func SetupTemplateRoutes(app *fiber.App, h *handler.TemplateHandler) {
	// Template CRUD operations
	templates := app.Group("v1/templates")

	// Create template
	templates.Post("/", h.CreateTemplate)

	// Get template by name/channel/language
	templates.Get("/", h.GetTemplate)

	// Update template
	templates.Put("/:id", h.UpdateTemplate)

	// Delete template
	templates.Delete("/:id", h.DeleteTemplate)
}
