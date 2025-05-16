package utils

import (
	"errors"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- Testable wrapper for validator.New().Struct ---

// ValidatorInterface allows mocking validator behavior in tests.
type ValidatorInterface interface {
	Struct(interface{}) error
}

// validateStructWithValidator allows injection of a mock validator.
func validateStructWithValidator[T any](dto T, v ValidatorInterface) error {
	err := v.Struct(dto)
	log.Println(err)
	if err != nil {
		return err
	}
	return nil
}

// --- Mock implementation ---

type mockValidator struct {
	errToReturn error
}

func (m *mockValidator) Struct(_ interface{}) error {
	return m.errToReturn
}

// --- Test DTO ---

type testDTO struct {
	Name string `validate:"required"`
	Age  int    `validate:"gte=0"`
}

// --- Unit tests ---

func TestValidateStructWithValidator_Success(t *testing.T) {
	// Suppress log output
	log.SetOutput(os.Stdout)

	dto := testDTO{Name: "Alice", Age: 30}
	mock := &mockValidator{errToReturn: nil}

	err := validateStructWithValidator(dto, mock)
	assert.NoError(t, err)
}

func TestValidateStructWithValidator_Failure(t *testing.T) {
	log.SetOutput(os.Stdout)

	dto := testDTO{Name: "", Age: -1}
	mock := &mockValidator{errToReturn: errors.New("validation failed")}

	err := validateStructWithValidator(dto, mock)
	assert.Error(t, err)
	assert.Equal(t, "validation failed", err.Error())
}

// --- Optional: Integration test for the real ValidateStruct ---

func TestValidateStruct_Integration(t *testing.T) {
	log.SetOutput(os.Stdout)

	dto := testDTO{Name: "", Age: -1}
	err := ValidateStruct(dto)
	assert.Error(t, err)
}
