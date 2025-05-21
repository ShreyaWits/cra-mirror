

package httpclient

import (
	"errors"
	"fmt"
)

var (
	ErrEncodingFailed = errors.New("request body encoding failed")
	ErrRequestFailed  = errors.New("http request failed")
	ErrDecodingFailed = errors.New("response decoding failed")
	ErrInvalidMethod  = errors.New("invalid HTTP method")
)

// HTTPError wraps an HTTP response with status and optional body
type HTTPError struct {
	StatusCode int
	Status     string
	Body       string // optional: only populate if needed
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http error: %d %s", e.StatusCode, e.Status)
}

func IsHTTPError(err error) (*HTTPError, bool) {
	var httpErr *HTTPError
	ok := errors.As(err, &httpErr)
	return httpErr, ok
}
