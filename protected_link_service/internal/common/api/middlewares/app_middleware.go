package middlewares

import (
	"log"
	commonDtos "protected_link/internal/common/api/dtos"
	"protected_link/internal/common/constants"
	"protected_link/internal/common/utils"

	"github.com/gofiber/fiber/v2"
)

func CommonRequestValidator() fiber.Handler {
	return func(c *fiber.Ctx) error {
		log.Printf("Request: %s %s", c.Method(), c.OriginalURL())

		// Log all query parameters
		for key, val := range c.Queries() {
			log.Printf("Query Param: %s = %s", key, val)
		}

		// Check JSON body for POST and PUT
		if c.Method() == fiber.MethodPost || c.Method() == fiber.MethodPut {
			// Check Content-Type
			if c.Get("Content-Type") != "application/json" {
				return c.Status(fiber.StatusUnsupportedMediaType).JSON(

					commonDtos.ApiResponseDto{
						Success: false,
						Error:   utils.GetMessage(string(constants.RequestBodyInvalidContentType)),
					})

			}

			// Validate JSON Body
			var body map[string]interface{}
			if err := c.BodyParser(&body); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(commonDtos.ApiResponseDto{
					Success: false,
					Error:   utils.GetMessage(string(constants.RequestBodyInvalidJSON)),
				})
			}
		}

		// Continue to actual handler
		return c.Next()
	}
}
