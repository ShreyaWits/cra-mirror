package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// Generic Middleware to parse and validate request body
func SetContextDataMiddleware(c *fiber.Ctx) error {
	var request map[string]interface{}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	c.Locals("contextData", &request)

	return c.Next()
}
