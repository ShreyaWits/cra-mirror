package utils

import (
	"thirdparty_service/internal/dtos"

	"github.com/go-playground/validator/v10"
)

var v = validator.New(validator.WithRequiredStructEnabled())

func Validate(data any) []error {
	err := v.Struct(data)
	if err == nil {
		return nil
	}

	var errs []error
	for _, err := range err.(validator.ValidationErrors) {
		fieldErr := dtos.FieldError{
			Code:    "VALIDATION_ERROR",
			Field:   err.Field(),
			Message: generateErrorMessage(err),
			Data:    err.Tag(),
		}
		errs = append(errs, fieldErr)
	}

	return errs
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
