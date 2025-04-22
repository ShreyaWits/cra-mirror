package middleware

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// ValidateParams is a middleware that validates the request body, query parameters, and path parameters against a given struct T.
//
// It first tries to parse the request body, query parameters, and path parameters into the given struct T using the respective parser functions.
// If any of the parsing fails, it returns a 400 Bad Request response with a JSON body containing a field error response with the appropriate error message.
//
// If all parsing succeeds, it then uses the go-playground/validator library to validate the struct T.
// If the validation fails, it returns a 400 Bad Request response with a JSON body containing a field error response with the appropriate error message for each validation error.
// If the validation succeeds, it stores the validated struct T in the context and continues to the next handler in the chain.
//
// The type parameter T must be a struct with fields that have validation tags.
// The struct must also have a no-argument constructor, so that the middleware can create a new instance of the struct to validate.
func ValidateParams[T any]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var dto T

		// Parse request body
		if err := c.BodyParser(&dto); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})

		}

		// Extract query parameters
		if err := c.QueryParser(&dto); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})

		}

		// Extract path parameters
		if err := c.ParamsParser(&dto); err != nil {

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		var validate = validator.New()
		// Validate the DTO
		if err := validate.Struct(dto); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		// Store the validated DTO in the context
		c.Locals("contextData", &dto)

		// Continue to the next handler
		return c.Next()
	}
}
