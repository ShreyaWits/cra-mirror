package utils

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type ResponseStruc struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Error   interface{} `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Success Response
func SendSuccess(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return c.Status(statusCode).JSON(ResponseStruc{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// Error Response
// func SendError(c *fiber.Ctx, statusCode int, message string, err interface{}) error {
// 	return c.Status(statusCode).JSON(ResponseStruc{
// 		Status:  "error",
// 		Message: message,
// 		Error:   err,
// 	})
// }

type SendError struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	ErrContent   interface{} `json:"error"`
}

func (se SendError) Error() string {
	return fmt.Sprintf("%s: %v", se.Message, se.ErrContent)
}