package middlewares

import (
	commonDtos "protected_link/internal/common/api/dtos"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

// ValidateBodyDTO validates the body of the incoming request and returns custom error responses
func ValidateBodyDTO(dto interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse the request body into the DTO
		if err := c.BodyParser(dto); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(commonDtos.ApiResponseDto{
				Success: false,
				Message: "Invalid request format",
				Error: []commonDtos.ValidationErrorDetail{
					{Field: "body", Message: err.Error()},
				},
			})
		}

		// Validate the parsed DTO
		if err := validate.Struct(dto); err != nil {
			var details []commonDtos.ValidationErrorDetail
			for _, e := range err.(validator.ValidationErrors) {
				var message string
				switch e.Tag() {
				case "required":
					message = e.Field() + " must be required"
				case "oneof":
					// Example: for a field that expects a set of values like "create" or "update"
					message = e.Field() + " must be one of " + e.Param()
				default:
					message = e.Field() + " must be " + e.Tag()
				}

				details = append(details, commonDtos.ValidationErrorDetail{
					Field:   e.Field(),
					Message: message,
				})
			}

			return c.Status(fiber.StatusBadRequest).JSON(commonDtos.ApiResponseDto{
				Success: false,
				Message: "Validation failed",
				Error:   details,
			})
		}

		// Save validated DTO for later use in the handler
		c.Locals("validatedBody", dto)
		return c.Next()
	}
}
