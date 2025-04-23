package utils

import (
	"nps-config-service/internal/modules/config-manager/apis/dtos"

	"github.com/gofiber/fiber/v2"
)

func SendSuccess(c *fiber.Ctx, statusCode int, message string, data interface{}) {
	response := dtos.SuccessResponse{
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	}
	c.Status(statusCode).JSON(response)
}

// SendError sends a JSON error response
func SendError(c *fiber.Ctx, statusCode int, message string, err interface{}) {
	response := dtos.ErrorResponse{
		StatusCode: statusCode,
		Message:    message,
		Error:      err,
	}
	c.Status(statusCode).JSON(response)
}
