package errors

import "fmt"

type CustomError struct {
	ErrorCode string
	Err       error
}

func (e *CustomError) Error() string {

	if e.Err != nil {
		return e.ErrorCode + ": " + e.Err.Error()
	}
	return e.Err.Error()
}

func (e *CustomError) ErrObj() error {
	return e.Err
}

// NewEncryptionError creates a new EncryptionError
func NewCustomError(message string, err error) *CustomError {
	return &CustomError{
		ErrorCode: message,
		Err:       err,
	}
}

// EncryptionError represents an error that occurred during encryption operations
type EncryptionError struct {
	Message string
	Err     error
}

func (e *EncryptionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// KeyManagementError represents an error that occurred during key management operations
type KeyManagementError struct {
	Message string
	Err     error
}

func (e *KeyManagementError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// ValidationError represents an error that occurred during request validation
type ValidationError struct {
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// BadRequestError represents an error that occurred due to invalid request parameters
type BadRequestError struct {
	Message string
	Err     error
}

func (e *BadRequestError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// UserServiceError represents an error that occurred during user service operations
type UserServiceError struct {
	Message string
	Err     error
}

func (e *UserServiceError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// NewEncryptionError creates a new EncryptionError
func NewEncryptionError(message string, err error) error {
	return &EncryptionError{
		Message: message,
		Err:     err,
	}
}

// NewKeyManagementError creates a new KeyManagementError
func NewKeyManagementError(message string, err error) error {
	return &KeyManagementError{
		Message: message,
		Err:     err,
	}
}

// NewValidationError creates a new ValidationError
func NewValidationError(message string, err error) error {
	return &ValidationError{
		Message: message,
		Err:     err,
	}
}

// NewBadRequestError creates a new BadRequestError
func NewBadRequestError(message string) error {
	return &BadRequestError{
		Message: message,
	}
}

// NewUserServiceError creates a new UserServiceError
func NewUserServiceError(message string, err error) error {
	return &UserServiceError{
		Message: message,
		Err:     err,
	}
}
