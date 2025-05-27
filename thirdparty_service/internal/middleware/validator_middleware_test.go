package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http" // Import net/http for status codes
	"net/http/httptest"
	"testing"
	"thirdparty_service/internal/dtos"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// Sample struct for testing validation
type TestStruct struct {
	Field1 string `json:"field1" validate:"required" error_code:"TS1000"` // Added error_code tag
	Field2 int    `json:"field2" validate:"min=10" error_code:"TS1000"`   // Added error_code tag
}

// Mock Fiber context for testing with a request body
func createMockContextWithBody(app *fiber.App, method, path string, body interface{}) *fiber.Ctx {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Use a dummy handler to capture the context after the request is processed by app.Test
	var capturedCtx *fiber.Ctx
	app.Add(method, path, func(c *fiber.Ctx) error {
		capturedCtx = c
		return nil // We just need to capture the context
	})

	app.Test(req) // Execute the request to capture the context
	return capturedCtx
}

func TestValidatorMiddleware_Success(t *testing.T) {
	app := fiber.New()
	// Use a dummy handler after the middleware to check if it was reached
	app.Post("/", ValidatorMiddleware[TestStruct](), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	validBody := TestStruct{
		Field1: "test",
		Field2: 15,
	}

	reqBody, _ := json.Marshal(validBody)
	req := httptest.NewRequest(fiber.MethodPost, "/", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.Nil(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode, "Handler should be reached on successful validation")
}

func TestValidatorMiddleware_BodyParseError(t *testing.T) {
	app := fiber.New()
	app.Post("/", ValidatorMiddleware[TestStruct](), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK) // This should not be reached
	})

	// Invalid JSON body
	reqBody := []byte(`{"field1": "test", "field2": "not an int"}`)
	req := httptest.NewRequest(fiber.MethodPost, "/", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.Nil(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Status code should be 400 for body parse error")

	var responseBody map[string]string
	bodyBytes, _ := io.ReadAll(resp.Body) // Read the response body
	defer resp.Body.Close()               // Close the response body
	json.Unmarshal(bodyBytes, &responseBody)
	assert.Equal(t, "Invalid request payload", responseBody["error"], "Error message should indicate invalid payload")
}

func TestValidatorMiddleware_ValidationError(t *testing.T) {
	// Create a new Fiber app with the error handler configured
	app := fiber.New(fiber.Config{
		ErrorHandler: FiberErrorHandler, // Use the custom error handler
	})

	// Apply the validator middleware to a route
	app.Post("/test", ValidatorMiddleware[TestStruct](), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK) // This should not be reached on validation error
	})

	// Invalid body (missing required field, field2 too small)
	invalidBody := TestStruct{
		Field2: 5,
	}
	reqBody, _ := json.Marshal(invalidBody)
	req := httptest.NewRequest(fiber.MethodPost, "/test", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute the request and get the response recorder
	resp, err := app.Test(req)

	assert.Nil(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Status code should be 400 for validation error")

	var errorResponse dtos.Error
	bodyBytes, _ := io.ReadAll(resp.Body) // Read the response body
	defer resp.Body.Close()               // Close the response body
	json.Unmarshal(bodyBytes, &errorResponse)

	assert.Equal(t, http.StatusBadRequest, errorResponse.Code, "Error code should be 400")
	assert.Equal(t, "Invalid Request Payload", errorResponse.Err, "Error message should indicate invalid payload")
	assert.NotNil(t, errorResponse.Data, "Error data should contain validation errors")

	// Assert the structure and content of the validation errors in the Data field
	fieldErrors, ok := errorResponse.Data.([]interface{}) // Data is unmarshalled as []interface{}
	assert.True(t, ok, "Error data should be a slice of interface{}")
	assert.Len(t, fieldErrors, 2, "Should have two validation errors")

	// Convert interface{} back to dtos.FieldError for detailed assertions
	var actualFieldErrors []dtos.FieldError
	for _, fe := range fieldErrors {
		var fieldErr dtos.FieldError
		jsonBytes, _ := json.Marshal(fe)
		json.Unmarshal(jsonBytes, &fieldErr)
		actualFieldErrors = append(actualFieldErrors, fieldErr)
	}

	// Assert specific field error details
	assert.Contains(t, actualFieldErrors, dtos.FieldError{Field: "Field1", Message: "Field1 is required", Code: "TS1000", Data: "required"})
	assert.Contains(t, actualFieldErrors, dtos.FieldError{Field: "Field2", Message: "Field2 must be at least 10 characters", Code: "TS1000", Data: "min"})
}
