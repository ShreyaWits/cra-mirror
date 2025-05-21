package errors_test

import (
	stderrors "errors"
	"messaging_service/pkg/errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomError_Error(t *testing.T) {
	// Test with a wrapped error
	baseErr := stderrors.New("base error")
	customErr := errors.NewCustomError("ERR_CODE", baseErr)

	assert.Equal(t, "ERR_CODE: base error", customErr.Error())
	assert.Equal(t, baseErr, customErr.ErrObj())

	// Test with nil error (edge case)
	nilErr := errors.NewCustomError("ERR_CODE", nil)
	assert.Equal(t, "ERR_CODE: <nil>", nilErr.Error())
	assert.Nil(t, nilErr.ErrObj())
}

func TestNewCustomError(t *testing.T) {
	baseErr := stderrors.New("some error")
	customErr := errors.NewCustomError("CUSTOM_CODE", baseErr)

	assert.IsType(t, &errors.CustomError{}, customErr)
	assert.Equal(t, "CUSTOM_CODE", customErr.ErrorCode)
	assert.Equal(t, baseErr, customErr.Err)
}
