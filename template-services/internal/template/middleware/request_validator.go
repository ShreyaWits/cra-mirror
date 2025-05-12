package middleware

import (
	"reflect"
	"strings"
	"template-services/internal/pkg/errors"

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
				"message": fiber.Map{
					"body": "Invalid or malformed JSON",
				},
				"error_code": appErrors.TmpErrInvalidRequestBody,
				"data":       fiber.Map{},
			})
		}

		// Step 2: Validate Struct
		if err := validate.Struct(body); err != nil {
			if ve, ok := err.(validator.ValidationErrors); ok {
				errMap := make(fiber.Map)

				// Collect missing or invalid fields
				for _, fe := range ve {
					field := strings.ToLower(fe.Field())

					// Default error message
					msg := ""
					switch fe.Tag() {
					case "required":
						msg = field + " is required"
					case "len":
						msg = field + " must be " + fe.Param() + " characters long"
					case "oneof":
						msg = field + " must be one of [" + fe.Param() + "]"
					case "nonempty":
						msg = field + " cannot be empty"
					default:
						msg = "Invalid value for " + field
					}

					// Attempt to override with error_code tag
					t := reflect.TypeOf(body)
					if t.Kind() == reflect.Ptr {
						t = t.Elem()
					}
					if f, ok := t.FieldByName(fe.Field()); ok {
						if customCode := f.Tag.Get("error_code"); customCode != "" {
							msg = appErrors.GetAppErrorMessage(customCode)
						}
					}

					// Store the error message for the field
					errMap[field] = msg
				}

				// Return the error map with specific missing values
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success":    false,
					"message":    errMap,
					"error_code": appErrors.TmpErrInvalidRequestBody,
					"data":       fiber.Map{},
				})
			}

			// Unexpected validation error
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success":    false,
				"message": fiber.Map{
					"validation": err.Error(),
				},
				"error_code": appErrors.TmpErrInvalidRequestBody,
				"data":       fiber.Map{},
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
				tag := fieldInfo.Tag
				code := tag.Get("error_code")
				if code == "" {
					code = "TmpErrInvalidRequestBody"
				}
				errMap[strings.ToLower(fe.Field())] = appErrors.GetAppErrorMessage(code)
			}
			return errMap, false
		}
		return map[string]string{"general": "validation error"}, false
	}

	return nil, true
}

