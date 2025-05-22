package httpclient_test

import (
	"bytes"
	"encoding/json"
	"io"
	httpclient "messaging_service/pkg/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONCodec_Encode(t *testing.T) {
	codec := &httpclient.JSONCodec{}

	tests := []struct {
		name        string
		input       interface{}
		wantErr     bool
		expectedCT  httpclient.ContentType
		expectedNil bool
	}{
		{
			name: "encode struct",
			input: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{Name: "John", Age: 30},
			wantErr:     false,
			expectedCT:  httpclient.ContentTypeJSON,
			expectedNil: false,
		},
		{
			name:        "encode nil",
			input:       nil,
			wantErr:     false,
			expectedCT:  "",
			expectedNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, contentType, err := codec.Encode(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCT, contentType)

			if tt.expectedNil {
				assert.Nil(t, reader)
				return
			}

			assert.NotNil(t, reader)

			// For non-nil case, verify the encoded JSON
			if tt.input != nil {
				var buf bytes.Buffer
				enc := json.NewEncoder(&buf)
				err = enc.Encode(tt.input)

				assert.NoError(t, err)

				// Read from reader and compare
				content, err := readAll(reader)
				assert.NoError(t, err)

				expected := buf.String()
				assert.Equal(t, expected, content)
			}
		})
	}
}

func TestJSONCodec_Decode(t *testing.T) {
	codec := &httpclient.JSONCodec{}

	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name     string
		input    string
		expected TestStruct
		wantErr  bool
	}{
		{
			name:     "valid json",
			input:    `{"name":"Jane","age":25}`,
			expected: TestStruct{Name: "Jane", Age: 25},
			wantErr:  false,
		},
		{
			name:    "invalid json",
			input:   `{"name":"Jane","age":}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			var result TestStruct

			err := codec.Decode(reader, &result)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to read all content from a reader
func readAll(r io.Reader) (string, error) {
	if r == nil {
		return "", nil
	}

	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
