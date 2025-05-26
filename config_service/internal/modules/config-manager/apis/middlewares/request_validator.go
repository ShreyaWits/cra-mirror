// middlewares/validation.go
package middlewares

import (
	"bytes"
	"encoding/json"
	"errors"
	ers "errors"
	"fmt"
	"log"
	"os"

	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	fiberOtel "github.com/psmarcin/fiber-opentelemetry/pkg/fiber-otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	common "nps-config-service/internal/common/errors"
	"nps-config-service/internal/modules/config-manager/apis/dtos"
)

var validate = validator.New()

// Register "matches" validation function
// var _ = validate.RegisterValidation("isValidName", utils.IsValidName)

// getValidationMessage generates user-friendly error messages
func getValidationMessage(jsonTag string, value interface{}, err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", jsonTag)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", jsonTag, err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", jsonTag, err.Param())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", jsonTag)
	case "isValidName":
		return fmt.Sprintf("%s must contain only letters and spaces", jsonTag)
	default:
		return fmt.Sprintf("%s is invalid", jsonTag)
	}
}

// ValidateBody is a middleware to validate the request body against the provided DTO
func ValidateBody(dto interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if os.Getenv("BYPASS_MIDDLEWARE") == "true" {
			return c.Next()
		}

		// Create a new span for tracing
		ctx := c.UserContext()
		parentCtx, parentSpan := fiberOtel.Tracer.Start(ctx, "Middleware")
		defer parentSpan.End()

		_, childSpan := fiberOtel.Tracer.Start(parentCtx, "Validate-Body")
		defer childSpan.End()
		// Create a new instance of the passed DTO
		dtoType := reflect.TypeOf(dto).Elem()
		newDto := reflect.New(dtoType).Interface()

		// Strict JSON decoding to detect unknown fields
		decoder := json.NewDecoder(bytes.NewReader(c.Body()))
		decoder.DisallowUnknownFields() // Prevent unknown fields

		if err := decoder.Decode(newDto); err != nil {
			childSpan.RecordError(err) // Record the error in the span
			childSpan.SetStatus(codes.Error, err.Error())
			childSpan.End()
			return c.Status(fiber.StatusBadRequest).JSON(dtos.ApiResponseDto{
				Success: false,
				Error: &dtos.ErrorResponseDto{

					Code:    common.Errors["CNF004"],
					Message: fmt.Sprintf("Invalid request body: %v", err.Error()),
				},
			})
		}

		log.Println("Body Extracted:", newDto)

		// Create a map to hold validation errors
		var validationErrors string

		// Reflect the value and type of the DTO
		val := reflect.ValueOf(newDto).Elem()

		// Iterate over the fields to validate only fields with JSON tags
		for i := 0; i < dtoType.NumField(); i++ {
			field := dtoType.Field(i)
			fieldValue := val.Field(i)

			// Check if the field has a JSON tag
			jsonTag := field.Tag.Get("json")
			if jsonTag != "" && jsonTag != "-" {
				if field.Tag.Get("validate") != "" {
					if err := validate.Var(fieldValue.Interface(), field.Tag.Get("validate")); err != nil {
						if validationErrs, ok := err.(validator.ValidationErrors); ok {
							for _, e := range validationErrs { // Iterate over validation errors
								validationErrors = getValidationMessage(jsonTag, fieldValue.Interface(), e)
								break // Show only one error
							}
						} else {
							validationErrors = err.Error()
							break
						}
					}
				}
			}
		}

		if len(validationErrors) > 0 {
			childSpan.RecordError(errors.New(fmt.Sprintf("DTO validation: Missing required fields: %v", validationErrors)))
			childSpan.SetStatus(codes.Error, fmt.Sprintf("DTO validation: %v", validationErrors))
			childSpan.End()
			return c.Status(fiber.StatusBadRequest).JSON(dtos.ApiResponseDto{
				Success: false,
				Error: &dtos.ErrorResponseDto{
					Code:    common.Errors["CNF004"],
					Message: validationErrors,
				},
			})
		}

		// Store the validated DTO in the context
		c.Locals("contextData", newDto)

		log.Println("Body Validated:", newDto)

		requestBody, err := json.Marshal(newDto)
		if err != nil {
			childSpan.RecordError(err)
			childSpan.SetStatus(codes.Error, err.Error())
			return err
		}

		// Add the request body as an attribute
		childSpan.SetAttributes(attribute.String("request.body", string(requestBody)))
		// If valid, continue to the next handler
		return c.Next()
	}
}

func ValidateParams(passedDto interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Create a new span for tracing
		ctx := c.UserContext()
		parentCtx, parentSpan := fiberOtel.Tracer.Start(ctx, "Middleware")
		defer parentSpan.End()

		_, childSpan := fiberOtel.Tracer.Start(parentCtx, "Validate-Params")
		defer childSpan.End()

		// Check if contextData is available
		dto := c.Locals("contextData")
		if dto == nil {
			dtoVal := reflect.ValueOf(passedDto)
			if dtoVal.Kind() != reflect.Ptr || dtoVal.IsNil() {
				childSpan.RecordError(ers.New("Invalid DTO type. Must be a pointer."))
				childSpan.SetStatus(codes.Error, "Invalid DTO type. Must be a pointer.")
				childSpan.End()
				return c.Status(fiber.StatusBadRequest).JSON(dtos.ApiResponseDto{
					Success: false,
					Error: &dtos.ErrorResponseDto{
						Code: common.Errors["CNF004"],

						Message: "Invalid DTO type. Must be a pointer.",
					},
				})
			}

			dto = reflect.New(reflect.TypeOf(passedDto).Elem()).Interface()
		}

		// Extract and set query parameters
		if err := c.QueryParser(dto); err != nil {
			childSpan.RecordError(err) // Record the error in the span
			childSpan.SetStatus(codes.Error, err.Error())
			childSpan.End()
			return c.Status(fiber.StatusBadRequest).JSON(dtos.ApiResponseDto{
				Success: false,
				Error: &dtos.ErrorResponseDto{
					Code: common.Errors["CNF004"],

					Message: "Failed to parse query parameters.",
				},
			})
		}

		// Extract and set path parameters
		if err := c.ParamsParser(dto); err != nil {
			childSpan.RecordError(err) // Record the error in the span
			childSpan.SetStatus(codes.Error, err.Error())
			childSpan.End()
			return c.Status(fiber.StatusBadRequest).JSON(dtos.ApiResponseDto{
				Success: false,
				Error: &dtos.ErrorResponseDto{
					Code: common.Errors["CNF004"],

					Message: "Failed to parse path parameters.",
				},
			})
		}

		// Extract and set JSON body parameters if present
		if len(c.Body()) > 0 {
			if err := c.BodyParser(dto); err != nil {
				childSpan.RecordError(err) // Record the error in the span
				childSpan.SetStatus(codes.Error, err.Error())
				childSpan.End()
				return c.Status(fiber.StatusBadRequest).JSON(dtos.ApiResponseDto{
					Success: false,
					Error: &dtos.ErrorResponseDto{
						Code:    common.Errors["CNF004"],
						Message: "Failed to parse JSON body.",
					},
				})
			}
		}

		// Validate the DTO
		if errs := validate.Struct(dto); errs != nil {
			errors := ""
			for _, err := range errs.(validator.ValidationErrors) {
				field, _ := reflect.TypeOf(dto).Elem().FieldByName(err.Field())

				if tag := field.Tag.Get("json"); tag != "" {
					continue
				}
				tag := field.Name
				errors = getValidationMessage(tag, nil, err)
			}

			if len(errors) > 0 {
				childSpan.RecordError(ers.New(fmt.Sprintf("DTO validation: Missing required fields: %v", errors)))
				childSpan.SetStatus(codes.Error, string(fmt.Sprintf("DTO validation: Missing required fields: %v", errors)))
				childSpan.End()
				return c.Status(fiber.StatusBadRequest).JSON(dtos.ApiResponseDto{
					Success: false,
					Error: &dtos.ErrorResponseDto{
						Code:    common.Errors["CNF004"],
						Message: fmt.Sprintf("DTO validation: Missing required fields: %v", errors),
					},
				})
			}
		}

		// Store the validated DTO in the context
		c.Locals("contextData", dto)

		requestParams, err := json.Marshal(dto)
		if err != nil {
			childSpan.RecordError(err)
			return err
		}

		// Add the request body as an attribute
		childSpan.SetAttributes(attribute.String("request.params", string(requestParams)))
		// Continue to the next handler
		return c.Next()
	}
}

// SetContextData initializes the context data with a default value if none is present
func SetContextData(passedDto interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Retrieve the DTO from context
		ctx := c.UserContext()
		parentCtx, parentSpan := fiberOtel.Tracer.Start(ctx, "Middleware")
		defer parentSpan.End()

		_, childSpan := fiberOtel.Tracer.Start(parentCtx, "Request-Context-Data")
		defer childSpan.End()
		dto := c.Locals("contextData")

		if dto == nil {
			// Initialize the DTO with the passed default value
			dto = reflect.New(reflect.TypeOf(passedDto).Elem()).Interface()
		}

		dtoVal := reflect.ValueOf(dto)
		if dtoVal.Kind() != reflect.Ptr || dtoVal.IsNil() {
			return c.Status(fiber.StatusBadRequest).JSON(dtos.ApiResponseDto{
				Success: false,
				Error: &dtos.ErrorResponseDto{
					Code:    common.Errors["CNF004"],
					Message: "Invalid request body",
				},
			})
		}

		dtoVal = dtoVal.Elem() // Dereference to get the actual struct

		// Use reflection to set the TokenID in the DTO
		tokenID := c.Locals("tokenID")
		if tokenID != nil {
			tokenIDField := dtoVal.FieldByName("TokenID")
			if tokenIDField.IsValid() && tokenIDField.CanSet() {
				// Check if the field is a pointer to string
				if tokenIDField.Kind() == reflect.Ptr && tokenIDField.Type().Elem().Kind() == reflect.String {
					ptr := reflect.New(tokenIDField.Type().Elem())
					ptr.Elem().Set(reflect.ValueOf(tokenID))
					tokenIDField.Set(ptr)
				}
			}
		}

		// Use reflection to set the UserID in the DTO
		userID := c.Locals("userID")
		if userID != nil {
			userIDField := dtoVal.FieldByName("UserID")
			if userIDField.IsValid() && userIDField.CanSet() {
				// Check if the field is a pointer to string
				if userIDField.Kind() == reflect.Ptr && userIDField.Type().Elem().Kind() == reflect.String {
					ptr := reflect.New(userIDField.Type().Elem())
					ptr.Elem().Set(reflect.ValueOf(userID))
					userIDField.Set(ptr)
				}
			}
		}

		// Use reflection to set the Role data in the DTO
		c.Locals("contextData", dto)

		contextDto, err := json.Marshal(dto)
		if err != nil {
			childSpan.RecordError(err)
			return err
		}

		// Add the request body as an attribute
		childSpan.SetAttributes(attribute.String("request.dto", string(contextDto)))

		// Continue to the next handler
		return c.Next()
	}
}
