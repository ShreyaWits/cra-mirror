package middlewares

import (
	"fmt"
	"log"
	commonDtos "protected_link/internal/common/api/dtos"

	"github.com/gofiber/fiber/v2"
)

func RecoveryMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic: %v", r)
				errMsg := fmt.Sprintf("%v", r)

				log.Printf("Recovered from panic: %s", errMsg)

				// Respond with a structured JSON error message
				_ = c.Status(fiber.StatusInternalServerError).JSON(commonDtos.ApiResponseDto{
					Success: false,
					Message: "Something went wrong!",
					Error:   errMsg})
			}
		}()
		return c.Next()
	}
}
