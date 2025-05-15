package middleware

import (
	"reflect"
	"strings"
	appErrors "template-services/internal/pkg/errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// New custom validation function to check for empty strings
func validateEmpty(fl validator.FieldLevel) bool {
	return strings.TrimSpace(fl.Field().String()) != ""
}

var validate = validator.New()

func init() {
	// Register custom validation for empty string check
	validate.RegisterValidation("nonempty", validateEmpty)
}

// ValidateBody validates incoming JSON request bodies and returns consistent error response
func ValidateBody[T any]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body T

		// Step 1: Parse JSON body
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success":    false,
				"message":    "Invalid or malformed JSON",
				"error_code": appErrors.TmpErrInvalidRequestBody,
			})
		}

		// Step 2: Validate Struct
		if err := validate.Struct(body); err != nil {
			if ve, ok := err.(validator.ValidationErrors); ok {
				errMap := make(fiber.Map)

				// Collect missing or invalid fields
				for _, fe := range ve {
					field := fe.StructField()
					fieldInfo, found := reflect.TypeOf(body).FieldByName(field)
					if !found {
						continue
					}

					// Get the JSON field name if available, otherwise use lowercased field name
					jsonTag := fieldInfo.Tag.Get("json")
					fieldName := strings.ToLower(fe.Field())
					if jsonTag != "" {
						fieldName = strings.Split(jsonTag, ",")[0]
					}

					// Get error message from error_code tag
					code := fieldInfo.Tag.Get("error_code")
					var errMsg string
					if code == "" || code == "TmpErrInvalidRequestBody" {
						errMsg = "Invalid or malformed JSON"
					} else {
						errMsg = appErrors.GetAppErrorMessage(code)
						if errMsg == "" {
							errMsg = "An unknown error occurred"
						}
					}

					errMap[fieldName] = errMsg
				}

				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success":    false,
					"message":    errMap,
					"error_code": appErrors.TmpErrInvalidRequestBody,
				})
			}

			// Unexpected validation error
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success":    false,
				"message":    "An unknown error occurred",
				"error_code": appErrors.TmpErrInvalidRequestBody,
			})
		}

		// Step 3: Store parsed & validated body into context
		c.Locals("body", body)
		return c.Next()
	}
}

func ValidateStruct[T any](input T) (map[string]string, bool) {
	errMap := make(map[string]string)

	if err := validate.Struct(input); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			for _, fe := range ve {
				field := fe.StructField()
				fieldInfo, found := reflect.TypeOf(input).FieldByName(field)
				if !found {
					continue
				}

				// Get the JSON field name if available, otherwise use lowercased field name
				jsonTag := fieldInfo.Tag.Get("json")
				fieldName := strings.ToLower(fe.Field())
				if jsonTag != "" {
					fieldName = strings.Split(jsonTag, ",")[0]
				}

				// Get error message from error_code tag
				code := fieldInfo.Tag.Get("error_code")
				var errMsg string
				if code == "" || code == "TmpErrInvalidRequestBody" {
					errMsg = "Invalid or malformed JSON"
				} else {
					errMsg = appErrors.GetAppErrorMessage(code)
					if errMsg == "" {
						errMsg = "An unknown error occurred"
					}
				}

				errMap[fieldName] = errMsg
			}
			return errMap, false
		}
		return map[string]string{"general": "An unknown error occurred"}, false
	}

	return nil, true
}
