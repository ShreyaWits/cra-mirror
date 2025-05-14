package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	appErrors "template-services/internal/pkg/errors"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// Dummy error code and message mapping for testing
const (
	DummyErrCode = "DummyErrCode"
)

// No patching of appErrors.GetAppErrorMessage - use the real function

// Test struct for validation
type TestRequest struct {
	Name    string `json:"name" validate:"required,nonempty" error_code:"DummyErrCode"`
	Channel string `json:"channel" validate:"required,oneof=email sms" error_code:"DummyErrCode"`
	Age     int    `json:"age" validate:"required" error_code:"DummyErrCode"`
}

func TestValidateStruct_Success(t *testing.T) {
	req := TestRequest{
		Name:    "John",
		Channel: "email",
		Age:     30,
	}
	errMap, ok := ValidateStruct(req)
	assert.True(t, ok)
	assert.Nil(t, errMap)
}

func TestValidateStruct_MissingFields(t *testing.T) {
	req := TestRequest{
		Name:    "",
		Channel: "",
		Age:     0,
	}
	errMap, ok := ValidateStruct(req)
	assert.False(t, ok)
	assert.Contains(t, errMap, "name")
	assert.Contains(t, errMap, "channel")
	assert.Contains(t, errMap, "age")
	assert.Equal(t, appErrors.GetAppErrorMessage(DummyErrCode), errMap["name"])
}

func TestValidateStruct_InvalidChannel(t *testing.T) {
	req := TestRequest{
		Name:    "John",
		Channel: "push", // invalid
		Age:     25,
	}
	errMap, ok := ValidateStruct(req)
	assert.False(t, ok)
	assert.Contains(t, errMap, "channel")
}

func TestValidateBody_Success(t *testing.T) {
	app := fiber.New()
	app.Post("/test", ValidateBody[TestRequest](), func(c *fiber.Ctx) error {
		body := c.Locals("body").(TestRequest)
		return c.JSON(fiber.Map{
			"success": true,
			"name":    body.Name,
			"channel": body.Channel,
			"age":     body.Age,
		})
	})

	payload := TestRequest{Name: "Alice", Channel: "sms", Age: 22}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.True(t, result["success"].(bool))
	assert.Equal(t, "Alice", result["name"])
	assert.Equal(t, "sms", result["channel"])
	assert.EqualValues(t, 22, result["age"])
}

func TestValidateBody_InvalidJSON(t *testing.T) {
	app := fiber.New()
	app.Post("/test", ValidateBody[TestRequest]())

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.False(t, result["success"].(bool))
	assert.Contains(t, result["message"].(map[string]interface{}), "body")
}

func TestValidateBody_ValidationError(t *testing.T) {
	app := fiber.New()
	app.Post("/test", ValidateBody[TestRequest]())

	payload := TestRequest{Name: "", Channel: "sms", Age: 0}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.False(t, result["success"].(bool))
	assert.Contains(t, result["message"].(map[string]interface{}), "name")
	assert.Contains(t, result["message"].(map[string]interface{}), "age")
}
