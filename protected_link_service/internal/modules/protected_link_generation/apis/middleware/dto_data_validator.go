package middleware

import (
	"encoding/json"
	"fmt"
	commonDtos "protected_link/internal/common/api/dtos"
	"protected_link/internal/module/models"
	apiDtos "protected_link/internal/module/protected_link_generation/apis/dtos"

	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func ValidateDTOKeys() fiber.Handler {
	return func(c *fiber.Ctx) error {

		validate := validator.New()
		requestBody := c.Locals("validatedBody").(*apiDtos.GenerateUrlRequest)
		var errorMessages []models.FieldErrorResponseDTO

		requetTypeOk := ValidateRequestType(requestBody.RequestType)

		if !requetTypeOk {
			errorMessages = append(errorMessages, models.FieldErrorResponseDTO{
				Field:        "request_type",
				ErrorMessage: fmt.Sprintf("Invalid value: should be one of [%s]", strings.Join(enum.GetValidRequestTypes(), ", ")),
				ErrorCode:    "InvalidRequestType",
			})
		}

		ok := ValidateModelType(requestBody.ModelType)
		if !ok {
			errorMessages = append(errorMessages, models.FieldErrorResponseDTO{
				Field:        "model_type",
				ErrorMessage: fmt.Sprintf("Invalid value: should be one of [%s]", strings.Join(enum.GetValidModelTypes(), ", ")),
				ErrorCode:    "InvalidRequestType",
			})
		}

		channel := ValidateModelType(requestBody.ChannelType)
		if !channel {
			errorMessages = append(errorMessages, models.FieldErrorResponseDTO{
				Field:        "channel_type",
				ErrorMessage: fmt.Sprintf("Invalid value: should be one of [%s]", strings.Join(enum.GetChannelTypes(), ", ")),
				ErrorCode:    "InvalidRequestType",
			})
		}

		// Validate the main DTO
		if err := validate.Struct(requestBody); err != nil {
			errors := ExtractValidationErrors(requestBody, err)
			errorMessages = append(errorMessages, errors...)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"fieldErrors": errorMessages,
			})
		}

		if !requetTypeOk {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"fieldErrors": errorMessages,
			})
		}

		var dataErr error

		switch requestBody.RequestType {
		case "auth":
			var authData models.AuthRequestData
			dataErr = ValidateDynamicData(requestBody.Data, &authData, validate)
		case "file":
			var fileData models.FileRequestData
			dataErr = ValidateDynamicData(requestBody.Data, &fileData, validate)
		case "payment":
			var paymentData models.PaymentRequestData
			dataErr = ValidateDynamicData(requestBody.Data, &paymentData, validate)
		default:
			dataErr = fmt.Errorf("unsupported request type")
		}

		if dataErr != nil {

			errorMessages = append(errorMessages, models.FieldErrorResponseDTO{
				Field:        "request_type",
				ErrorMessage: dataErr.Error(),
				ErrorCode:    "InvalidData",
			})
			return c.Status(fiber.StatusBadRequest).JSON(commonDtos.ApiResponseDto{Success: false, Message: "Invalid Data", Error: errorMessages})

		}
		return c.Next()
	}

}

// ✅ Extract validation errors into DTO
func ExtractValidationErrors[T any](dto T, err error) []models.FieldErrorResponseDTO {
	var errorMessages []models.FieldErrorResponseDTO
	fieldMap := GetFieldJsonMap(dto)

	for _, err := range err.(validator.ValidationErrors) {
		jsonKey := fieldMap[err.StructField()]
		switch err.Tag() {
		case "requesttype":
			errorMessages = append(errorMessages, models.FieldErrorResponseDTO{
				Field:        jsonKey,
				ErrorMessage: fmt.Sprintf("Invalid value: should be one of [%s]", strings.Join(enum.GetValidRequestTypes(), ", ")),
				ErrorCode:    "InvalidRequestType",
			})
		case "modeltype":
			errorMessages = append(errorMessages, models.FieldErrorResponseDTO{
				Field:        jsonKey,
				ErrorMessage: fmt.Sprintf("Invalid value: should be one of [%s]", strings.Join(enum.GetValidModelTypes(), ", ")),
				ErrorCode:    "InvalidModelType",
			})
		default:
			errorMessages = append(errorMessages, models.FieldErrorResponseDTO{
				Field:        jsonKey,
				ErrorMessage: fmt.Sprintf("Validation failed for '%s'", err.Tag()),
				ErrorCode:    "ValidationFailed",
			})
		}
	}

	return errorMessages
}

// ✅ Validate dynamic data
func ValidateDynamicData(data apiDtos.JSONB, target interface{}, validate *validator.Validate) error {
	// Marshal JSONB to bytes
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Unmarshal into the target struct
	if err := json.Unmarshal(jsonData, target); err != nil {
		return fmt.Errorf("invalid data structure: %w", err)
	}

	// Validate the unmarshaled struct
	if err := validate.Struct(target); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

// ✅ Extract JSON field names from struct
func GetFieldJsonMap(dto interface{}) map[string]string {
	fieldMap := make(map[string]string)
	val := reflect.TypeOf(dto)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			fieldMap[field.Name] = jsonTag
		}
	}

	return fieldMap
}

// ✅ Custom validator for RequestType
func ValidateRequestType(value string) bool {
	for _, valid := range enum.GetValidRequestTypes() {
		if value == valid {
			return true
		}
	}
	return false
}

// ✅ Custom validator for ModelType
func ValidateModelType(value string) bool {
	for _, valid := range enum.GetValidModelTypes() {
		if value == valid {
			return true
		}
	}
	return false
}
