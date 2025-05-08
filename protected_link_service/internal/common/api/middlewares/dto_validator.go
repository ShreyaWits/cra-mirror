package middlewares

import (
	commonDtos "protected_link/internal/common/api/dtos"
	"protected_link/internal/common/constants"
	"protected_link/internal/common/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

// ValidateBodyDTO validates the body of the incoming request and returns custom error responses
func ValidateBodyDTO(dtoFactory func() interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		dto := dtoFactory() // Create a new DTO instance for this request

		// Parse the request body into the DTO
		if err := c.BodyParser(dto); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(commonDtos.ApiResponseDto{
				Success: false,
				Message: utils.GetMessage(string(constants.RequestInvalidFormat)),
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
					message = e.Field() + " is required"
				case "oneof":
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
				Message: utils.GetMessage(string(constants.RequestValidationFailed)),
				Error:   details,
			})
		}

		c.Locals("validatedBody", dto)
		return c.Next()
	}
}
