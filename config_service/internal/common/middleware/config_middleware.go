package middleware

import (
	common "nps-config-service/internal/common/errors"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
)

// Generic Middleware to parse and validate request body
func SetContextDataMiddleware[T any](c *fiber.Ctx) error {
	var request T
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	c.Locals("contextData", &request)

	return c.Next()
}

func JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(common.ThrowError(fiber.StatusBadRequest, "AUTH001"))
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(common.ThrowError(fiber.StatusBadRequest, "AUTH002"))
		}

		tokenStr := parts[1]
		jwtSecret := os.Getenv("JWT_SECRET")

		// Parse and validate token
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			// Make sure signing method is HMAC
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(common.ThrowError(fiber.StatusBadRequest, "AUTH003"))
		}

		return c.Next()
	}
}
