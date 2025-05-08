package utils

import (
	"log"

	"github.com/go-playground/validator/v10"
)

func ValidateStruct[T any](dto T) error {

	v := validator.New()
	err := v.Struct(dto)

	log.Println(err)

	if err != nil {
		return err
	}

	return nil
}
