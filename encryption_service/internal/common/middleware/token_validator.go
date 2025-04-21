package commomiddleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func TokenValidator() fiber.Handler {
	var token string
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "token is required"})
		}
		// Store the validated DTO in the context
		c.Locals("token", token)

		// Continue to the next handler
		return c.Next()
	}
}
