package middleware

import (
	"fmt"
	commonDtos "protected_link/internal/common/api/dtos"
	"protected_link/internal/common/constants"
	"protected_link/internal/common/utils"

	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"
	enum "protected_link/internal/modules/link_generation/apis/enums"
	"protected_link/internal/modules/link_generation/models"

	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func ValidateDTOKeys() fiber.Handler {
	return func(c *fiber.Ctx) error {

		validate := validator.New()
		requestBody := c.Locals("validatedBody").(*apiDtos.GenerateUrlRequest)

		if requestBody.ExpireIn == "" {
			requestBody.ExpireIn = "24h"
		}
		var errorMessages []models.FieldErrorResponseDTO

		requetTypeOk := ValidateRequestType(requestBody.RequestType)

		if !requetTypeOk {
			errorMessages = append(errorMessages, models.FieldErrorResponseDTO{
				Field:        "request_type",
				ErrorMessage: fmt.Sprintf("%s [%s]", utils.GetMessage(string(constants.InvalidValueShouldBeOneOfList)), strings.Join(enum.GetValidRequestTypes(), ", ")),
				ErrorCode:    "InvalidRequestType",
			})
		}

		channel := ValidateModelType(requestBody.ChannelType)
		if !channel {
			errorMessages = append(errorMessages, models.FieldErrorResponseDTO{
				Field:        "channel_type",
				ErrorMessage: fmt.Sprintf("%s [%s]", utils.GetMessage(string(constants.InvalidValueShouldBeOneOfList)), strings.Join(enum.GetChannelTypes(), ", ")),
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

func ValidateDynamicData[T any](jSONB apiDtos.JSONB, requestData *T, validate *validator.Validate) error {
	// Implement the logic to validate the dynamic data
	// For now, return nil to avoid compilation errors
	return nil
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
				ErrorMessage: fmt.Sprintf("%s [%s]", utils.GetMessage(string(constants.InvalidValueShouldBeOneOfList)), strings.Join(enum.GetValidRequestTypes(), ", ")),
				ErrorCode:    "InvalidRequestType",
			})

		default:
			errorMessages = append(errorMessages, models.FieldErrorResponseDTO{
				Field:        jsonKey,
				ErrorMessage: fmt.Sprintf("%s '%s'", utils.GetMessage(string(constants.ValidationFailedFor)), err.Tag()),
				ErrorCode:    "ValidationFailed",
			})
		}
	}

	return errorMessages
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
	for _, valid := range enum.GetChannelTypes() {
		if value == valid {
			return true
		}
	}
	return false
}
