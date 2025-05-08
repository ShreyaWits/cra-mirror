package errors

import (
	"google.golang.org/grpc/codes"
)

type CustomError struct {
	code      codes.Code
	ErrorCode string
	Err       error
}

func (e *CustomError) Error() string {

	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Err.Error()
}

func (e *CustomError) ErrObj() error {
	return e.Err
}

// NewEncryptionError creates a new EncryptionError
func NewCustomError(code codes.Code, ErrorCode string, err error) *CustomError {
	return &CustomError{
		code:      code,
		ErrorCode: ErrorCode,
		Err:       err,
	}
}
