package httpclient_test

import (
	"errors"
	httpclient "messaging_service/pkg/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPError_Error(t *testing.T) {
	httpErr := &httpclient.HTTPError{
		StatusCode: 404,
		Status:     "Not Found",
		Body:       "The requested resource was not found",
	}

	expected := "http error: 404 Not Found"
	assert.Equal(t, expected, httpErr.Error())
}

func TestIsHTTPError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
		wantOk   bool
	}{
		{
			name: "is HTTP error",
			err: &httpclient.HTTPError{
				StatusCode: 500,
				Status:     "Internal Server Error",
			},
			wantCode: 500,
			wantOk:   true,
		},
		{
			name:     "not HTTP error",
			err:      errors.New("regular error"),
			wantCode: 0,
			wantOk:   false,
		},
		{
			name:     "nil error",
			err:      nil,
			wantCode: 0,
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpErr, ok := httpclient.IsHTTPError(tt.err)

			assert.Equal(t, tt.wantOk, ok)

			if tt.wantOk {
				assert.NotNil(t, httpErr)
				assert.Equal(t, tt.wantCode, httpErr.StatusCode)
			} else if tt.err != nil {
				assert.Nil(t, httpErr)
			}
		})
	}
}

func TestPackageErrorConstants(t *testing.T) {
	// Verify that the package-level error constants are defined
	assert.NotNil(t, httpclient.ErrEncodingFailed)
	assert.Equal(t, "request body encoding failed", httpclient.ErrEncodingFailed.Error())

	assert.NotNil(t, httpclient.ErrRequestFailed)
	assert.Equal(t, "http request failed", httpclient.ErrRequestFailed.Error())

	assert.NotNil(t, httpclient.ErrDecodingFailed)
	assert.Equal(t, "response decoding failed", httpclient.ErrDecodingFailed.Error())

	assert.NotNil(t, httpclient.ErrInvalidMethod)
	assert.Equal(t, "invalid HTTP method", httpclient.ErrInvalidMethod.Error())
}
