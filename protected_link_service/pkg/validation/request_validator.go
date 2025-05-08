package validation

import (
	"fmt"
	"protected_link/internal/modules/authentication/models"
	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

// Initialize custom validators
func InitValidator() {
	Validate = validator.New()

	// Custom email validator
	Validate.RegisterValidation("emailcustom", func(fl validator.FieldLevel) bool {
		email := fl.Field().String()
		return regexp.MustCompile(`.+@.+\..+`).MatchString(email)
	})
}

func ValidateGenerateUrlRequest(req apiDtos.GenerateUrlRequest) (map[string]string, error) {
	err := Validate.Struct(req)
	if err != nil {
		fieldErrors := make(map[string]string)

		for _, fe := range err.(validator.ValidationErrors) {
			field := fe.Field()
			tag := fe.Tag()
			fieldErrors[field] = fmt.Sprintf("Validation failed for field '%s': %s", field, tag)
		}

		return fieldErrors, err
	}
	return nil, nil
}

func ValidateVerifyOtpRequest(req models.VerifyOTPRequest) (map[string]interface{}, error) {
	err := Validate.Struct(req)
	if err != nil {
		fieldErrors := make(map[string]interface{})
		for _, fe := range err.(validator.ValidationErrors) {
			field := fe.Field()
			tag := fe.Tag()
			if _, exists := fieldErrors[field]; !exists {
				fieldErrors[field] = []string{}
			}
			fieldErrors[field] = append(fieldErrors[field].([]string), fmt.Sprintf("Validation failed for field '%s': %s", field, tag))
		}
		return fieldErrors, err
	}
	return nil, nil
}
