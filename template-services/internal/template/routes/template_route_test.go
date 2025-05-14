package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"template-services/internal/template/dto"
	"template-services/internal/template/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// MockDeleteTemplateHandler simulates the DeleteTemplate handler.
func MockDeleteTemplateHandler(c *fiber.Ctx) error {
	var req dto.DeleteTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}
	// Simulate successful deletion
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Template deleted successfully",
	})
}

func TestTemplateDeleteRoute_Success(t *testing.T) {
	app := fiber.New()

	// Register the route with the mock handler and real validation middleware
	app.Post("/delete/", middleware.ValidateBody[dto.DeleteTemplateRequest](), MockDeleteTemplateHandler)

	payload := dto.DeleteTemplateRequest{
		TemplateID: "template-123",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/delete/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.True(t, result["success"].(bool))
	assert.Equal(t, "Template deleted successfully", result["message"])
}

func TestTemplateDeleteRoute_ValidationError(t *testing.T) {
	app := fiber.New()

	app.Post("/delete/", middleware.ValidateBody[dto.DeleteTemplateRequest](), MockDeleteTemplateHandler)

	// Missing ID field
	payload := map[string]string{}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/delete/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.False(t, result["success"].(bool))

	// Updated assertion: message is a map with "templateid" as key
	msgMap, ok := result["message"].(map[string]interface{})
	assert.True(t, ok, "message should be a map")
	assert.Contains(t, msgMap, "templateid")
}
