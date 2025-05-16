package utils

import (
	"github.com/go-playground/validator/v10"
)

func ValidateStruct[T any](dto T) error {
	v := validator.New()
	return v.Struct(dto)
}
