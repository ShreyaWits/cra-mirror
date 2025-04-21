package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// HandleHealth handles the health check request
func (h *EncryptionHandlerImpl) HandleHealth(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "healthy",
		"version": "1.0.0",
	})
}
