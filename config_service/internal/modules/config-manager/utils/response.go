package utils

import (
	"nps-config-service/internal/modules/config-manager/apis/dtos"

	"github.com/gofiber/fiber/v2"
)

func SendSuccess(c *fiber.Ctx, statusCode int, message string, data interface{}) {
	response := dtos.SuccessResponse{
		Success:    true,
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	}
	c.JSON(response)
}

// SendError sends a JSON error response
func SendError(c *fiber.Ctx, statusCode int, message string, err interface{}) {
	response := dtos.ErrorResponse{
		Success:    false,
		StatusCode: statusCode,
		Message:    message,
		Error:      err,
	}
	c.JSON(response)
}
