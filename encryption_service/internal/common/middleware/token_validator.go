package middleware

import (
	"encryption_microservice/internal/modules/encryption/api/mapper"
	"encryption_microservice/pkg/logger"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func TokenValidator() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		logger.Error("🔍 Checking Authorization header", fmt.Errorf("ff"))

		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			if token == "" {
				logger.Error("❌ Bearer token is empty", fmt.Errorf("ff"))
				return mapper.NewErrorResponse(c, "token is required", int(fiber.StatusBadRequest), "")
			}

			logger.Error("✅ Bearer token extracted successfully", fmt.Errorf("fff %s fff", token))
			// Store the validated token in the context
			c.Locals("token", token)
		} else {
			logger.Error("❌ Authorization header missing or malformed", fmt.Errorf("ff"))
			return mapper.NewErrorResponse(c, "token is required", int(fiber.StatusBadRequest), "")
		}

		// Continue to the next handler
		return c.Next()
	}
}
