package middleware

import (
	"errors"
	"thirdparty_service/internal/dtos"

	"github.com/gofiber/fiber/v2"
)

// FiberErrorHandler Fiber error handler for app.Config.ErrorHandler
func FiberErrorHandler(ctx *fiber.Ctx, err error) error {
	var apiRes dtos.Response
	var apiErr dtos.Error

	switch {
	case errors.As(err, &apiRes):
		return ctx.Status(apiRes.Code).JSON(apiRes)

	case errors.As(err, &apiErr):
		return ctx.Status(apiErr.Code).JSON(apiErr)

	default:
		return ctx.Status(fiber.StatusInternalServerError).JSON(dtos.Error{
			Code: fiber.StatusInternalServerError,
			Err:  "Internal Server Error",
		})
	}
}
