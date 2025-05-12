package http

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/stretchr/testify/assert"
)

func TestTraceMiddleware(t *testing.T) {
	setupTestContainer()

	// Test successful request
	t.Run("successful request", func(t *testing.T) {
		app := fiber.New()
		app.Use(recover.New())
		app.Use(TraceMiddleware())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendString("ok")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	// Test request with error
	t.Run("request with error", func(t *testing.T) {
		app := fiber.New()
		app.Use(recover.New())
		app.Use(TraceMiddleware())
		app.Get("/error", func(c *fiber.Ctx) error {
			return fiber.NewError(fiber.StatusInternalServerError, "test error")
		})

		req := httptest.NewRequest("GET", "/error", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	})

	// Test request with panic
	t.Run("request with panic", func(t *testing.T) {
		app := fiber.New()
		app.Use(recover.New())
		app.Use(TraceMiddleware())
		app.Get("/panic", func(c *fiber.Ctx) error {
			panic("test panic")
		})

		req := httptest.NewRequest("GET", "/panic", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	})

	// Test request with custom headers
	t.Run("request with custom headers", func(t *testing.T) {
		app := fiber.New()
		app.Use(recover.New())
		app.Use(TraceMiddleware())
		app.Get("/headers", func(c *fiber.Ctx) error {
			return c.SendString("ok")
		})

		req := httptest.NewRequest("GET", "/headers", nil)
		req.Header.Set("X-Custom-Header", "test-value")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	// Test request with query parameters
	t.Run("request with query parameters", func(t *testing.T) {
		app := fiber.New()
		app.Use(recover.New())
		app.Use(TraceMiddleware())
		app.Get("/query", func(c *fiber.Ctx) error {
			return c.SendString("ok")
		})

		req := httptest.NewRequest("GET", "/query?param=value", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	// Test request with different HTTP methods
	t.Run("different HTTP methods", func(t *testing.T) {
		app := fiber.New()
		app.Use(recover.New())
		app.Use(TraceMiddleware())
		app.Post("/method", func(c *fiber.Ctx) error {
			return c.SendString("ok")
		})

		req := httptest.NewRequest("POST", "/method", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	// Test request with invalid path
	t.Run("invalid path", func(t *testing.T) {
		app := fiber.New()
		app.Use(recover.New())
		app.Use(TraceMiddleware())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendString("ok")
		})

		req := httptest.NewRequest("GET", "/invalid", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})

	// Test request with invalid method
	t.Run("invalid method", func(t *testing.T) {
		app := fiber.New()
		app.Use(recover.New())
		app.Use(TraceMiddleware())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendString("ok")
		})

		req := httptest.NewRequest("PUT", "/test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusMethodNotAllowed, resp.StatusCode)
	})
}
