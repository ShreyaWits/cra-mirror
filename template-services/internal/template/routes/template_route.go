package routes

import (
	appErrors "template-services/internal/pkg/errors"
	"template-services/internal/template/dto"
	"template-services/internal/template/handler"
	"template-services/internal/template/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupTemplateRoutes(app *fiber.App, h *handler.TemplateHandler) {
	// Validation error middleware
	app.Use(func(c *fiber.Ctx) error {
		err := c.Next()
		if err != nil {
			code := fiber.StatusBadRequest
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success":    false,
				"message":    err.Error(),
				"error_code": appErrors.TmpErrInvalidRequestBody,
			})
		}
		return nil
	})

	// Register routes
	v1 := app.Group("/v1/templates")
	v1.Post("/create", middleware.ValidateBody[dto.CreateTemplateRequest](), h.CreateTemplate)
	v1.Post("/get", middleware.ValidateBody[dto.GetTemplateRequest](), h.GetTemplate)
	v1.Post("/update", middleware.ValidateBody[dto.UpdateTemplateRequest](), h.UpdateTemplate)
	v1.Post("/delete", middleware.ValidateBody[dto.DeleteTemplateRequest](), h.DeleteTemplate)

	// 405 handler: route exists but wrong method
	app.Use(func(c *fiber.Ctx) error {
	if c.Route() != nil {
		return c.Next() // Valid route and method
	}

	// Check if a route exists for the same path but different method
	for _, r := range app.GetRoutes() {
		if r.Path == c.Path() {
			// Path exists but wrong method => 405
			return c.Status(fiber.StatusMethodNotAllowed).JSON(fiber.Map{
				"success":    false,
				"message":    "Method Not Allowed",
				"error_code": appErrors.TmpErrInvalidRequestBody,
			})
		}
	}

	// Fall back to next middleware
	return c.Next()
})

	// 404 fallback handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success":    false,
			"message":    "Route not found",
			"error_code": appErrors.TmpErrInvalidRequestBody,
		})
	})
}
