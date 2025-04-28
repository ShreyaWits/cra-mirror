package middlewares

import (
	"protected_link/internal/common/constants"
	"protected_link/internal/common/utils"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": utils.GetMessage(string(constants.UnauthorizedAccess))})
		}

		// Validate token logic here
		return c.Next()
	}
}
