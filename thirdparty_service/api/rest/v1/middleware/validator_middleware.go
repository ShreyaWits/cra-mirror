package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	v1 "thirdparty_service/api/rest/v1"
	"thirdparty_service/utils"
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
			return v1.Error{
				Code: http.StatusBadRequest,
				Err:  "Invalid Request Payload",
				Data: err,
			}
		}

		return c.Next()
	}
}
