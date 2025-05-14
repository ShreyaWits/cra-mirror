package utils

import (
	"reflect"

	"thirdparty_service/internal/dtos"

	"github.com/go-playground/validator/v10"
)

var v = validator.New(validator.WithRequiredStructEnabled())

func Validate(data any) []dtos.FieldError {
	err := v.Struct(data)
	if err == nil {
		return nil
	}

	var errs []dtos.FieldError
	for _, fe := range err.(validator.ValidationErrors) {
		errorCode := getErrorCode(data, fe.StructField())
		fieldErr := dtos.FieldError{
			Code:    errorCode, // use the extracted error_code tag
			Field:   fe.Field(),
			Message: generateErrorMessage(fe),
			Data:    fe.Tag(),
		}
		errs = append(errs, fieldErr)
	}

	return errs
}

// getErrorCode uses reflection to extract the `error_code` tag
func getErrorCode(structVal any, fieldName string) string {
	rt := reflect.TypeOf(structVal)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	if field, ok := rt.FieldByName(fieldName); ok {
		return field.Tag.Get("error_code")
	}
	return "VALIDATION_ERROR" // fallback default
}

func generateErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "max":
		return fe.Field() + " must be at most " + fe.Param() + " characters"
	case "min":
		return fe.Field() + " must be at least " + fe.Param() + " characters"
	case "email":
		return fe.Field() + " must be a valid email"
	default:
		return fe.Field() + " is invalid"
	}
}
