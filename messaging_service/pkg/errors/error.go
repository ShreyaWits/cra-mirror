package errors

type CustomError struct {
	ErrorCode string
	Err       error
}

func (e *CustomError) Error() string {
	if e.Err != nil {
		return e.ErrorCode + ": " + e.Err.Error()
	}
	return e.ErrorCode + ": <nil>"
}

func (e *CustomError) ErrObj() error {
	return e.Err
}

func NewCustomError(message string, err error) *CustomError {
	return &CustomError{
		ErrorCode: message,
		Err:       err,
	}
}
