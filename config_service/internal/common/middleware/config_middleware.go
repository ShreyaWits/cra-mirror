package middleware

import (
	common "nps-config-service/internal/common/errors"
	"nps-config-service/internal/common/roles"
	"nps-config-service/internal/configs"
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

// JWTMiddleware validates JWT token and extracts role
func JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(common.ThrowError(fiber.StatusUnauthorized, "AUTH001"))
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(common.ThrowError(fiber.StatusUnauthorized, "AUTH002"))
		}

		tokenStr := parts[1]
		jwtSecret := configs.AppConfig.JWTSecret

		// Parse and validate token
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(common.ThrowError(fiber.StatusUnauthorized, "AUTH003"))
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(common.ThrowError(fiber.StatusUnauthorized, "AUTH003"))
		}

		// Extract and validate role
		role, ok := claims["role"].(string)
		if !ok || (role != "ADMIN" && role != "VIEWER") {
			return c.Status(fiber.StatusForbidden).JSON(common.ThrowError(fiber.StatusForbidden, "AUTH004"))
		}

		// Store role in context for RBAC middleware
		c.Locals("userRole", role)
		c.Locals("username", claims["username"])

		return c.Next()
	}
}

// RequireRole middleware checks if the user has the required role
func RequireRole(requiredRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals("userRole").(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(common.ThrowError(fiber.StatusUnauthorized, "AUTH003"))
		}

		userLevel, exists := roles.RoleHierarchy[userRole]
		if !exists {
			return c.Status(fiber.StatusForbidden).JSON(common.ThrowError(fiber.StatusForbidden, "AUTH004"))
		}

		for _, role := range requiredRoles {
			requiredLevel, ok := roles.RoleHierarchy[role]
			if ok && userLevel <= requiredLevel {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(common.ThrowError(fiber.StatusForbidden, "AUTH004"))
	}
}
