package kafka

type RetriableError struct {
	Err       error
	Retryable bool
}

func (e RetriableError) Error() string {
	return e.Err.Error()
}