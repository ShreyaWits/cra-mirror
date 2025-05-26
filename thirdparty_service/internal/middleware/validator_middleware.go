package middleware

import (
	"net/http"

	"thirdparty_service/internal/dtos"
	"thirdparty_service/internal/utils"

	"github.com/gofiber/fiber/v2"
)

func ValidatorMiddleware[T any]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var data T

		if err := c.BodyParser(&data); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request payload",
			})
		}

		if err := utils.Validate(&data); err != nil {
			return dtos.Error{
				Code: http.StatusBadRequest,
				Err:  "Invalid Request Payload",
				Data: err,
			}
		}

		return c.Next()
	}
}
