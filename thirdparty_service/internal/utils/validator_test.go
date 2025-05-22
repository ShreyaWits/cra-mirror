package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"thirdparty_service/internal/utils"
)

// Sample DTO with validation and error_code tags
type TestInput struct {
	Name     string `validate:"required" error_code:"TS1001"`
	Email    string `validate:"required,email" error_code:"TS1002"`
	Age      int    `validate:"min=18,max=60" error_code:"TS1003"`
	Comments string `validate:"max=10" error_code:"TS1004"`
}

func TestValidate_ValidInput(t *testing.T) {
	input := TestInput{
		Name:     "John",
		Email:    "john@example.com",
		Age:      30,
		Comments: "short msg",
	}
	errs := utils.Validate(input)
	assert.Nil(t, errs, "expected no validation errors for valid input")
}

func TestValidate_MissingRequiredFields(t *testing.T) {
	input := TestInput{}
	errs := utils.Validate(input)
	assert.Len(t, errs, 3) // Name, Email, Age

	expectedCodes := []string{"TS1001", "TS1002", "TS1003"}
	fields := []string{"Name", "Email", "Age"}

	for i, err := range errs {
		assert.Equal(t, expectedCodes[i], err.Code)
		assert.Equal(t, fields[i], err.Field)
		assert.NotEmpty(t, err.Message)
	}
}

func TestValidate_InvalidEmail(t *testing.T) {
	input := TestInput{
		Name:  "Alice",
		Email: "invalid-email",
		Age:   25,
	}
	errs := utils.Validate(input)
	assert.Len(t, errs, 1)
	assert.Equal(t, "TS1002", errs[0].Code)
	assert.Equal(t, "Email", errs[0].Field)
	assert.Contains(t, errs[0].Message, "valid email")
}

func TestValidate_AgeTooLow(t *testing.T) {
	input := TestInput{
		Name:  "Bob",
		Email: "bob@example.com",
		Age:   15,
	}
	errs := utils.Validate(input)
	assert.Len(t, errs, 1)
	assert.Equal(t, "TS1003", errs[0].Code)
	assert.Equal(t, "Age", errs[0].Field)
	assert.Contains(t, errs[0].Message, "at least")
}

func TestValidate_CommentTooLong(t *testing.T) {
	input := TestInput{
		Name:     "Jane",
		Email:    "jane@example.com",
		Age:      25,
		Comments: "this comment is too long",
	}
	errs := utils.Validate(input)
	assert.Len(t, errs, 1)
	assert.Equal(t, "TS1004", errs[0].Code)
	assert.Equal(t, "Comments", errs[0].Field)
	assert.Contains(t, errs[0].Message, "at most")
}

func TestValidate_FieldWithoutErrorCodeTag(t *testing.T) {
	type InputWithMissingTag struct {
		FieldX string `validate:"required"` // no error_code tag
	}
	input := InputWithMissingTag{}

	errs := utils.Validate(input)
	assert.Len(t, errs, 1)
	assert.Equal(t, "TS1000", errs[0].Code) // fallback code
	assert.Equal(t, "FieldX", errs[0].Field)
}
