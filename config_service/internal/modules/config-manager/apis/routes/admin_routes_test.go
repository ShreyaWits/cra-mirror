package routes

import (
	handler "nps-config-service/internal/modules/config-manager/apis/handlers"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestRegisterAdminRoutes(t *testing.T) {
	// Create a new Fiber app
	app := fiber.New()

	// Create a mock admin handler
	mockHandler := &handler.AdminHandler{}

	// Register admin routes
	adminGroup := app.Group("/api")
	RegisterAdminRoutes(adminGroup, mockHandler)

	// Test if routes are registered correctly
	routes := app.GetRoutes()
	assert.Equal(t, 2, len(routes), "Expected 2 routes to be registered")

	// Verify signup route
	signupRoute := routes[0]
	assert.Equal(t, "POST", signupRoute.Method, "Expected POST method for signup route")
	assert.Equal(t, "/api/admin/signup", signupRoute.Path, "Expected correct path for signup route")

	// Verify login route
	loginRoute := routes[1]
	assert.Equal(t, "POST", loginRoute.Method, "Expected POST method for login route")
	assert.Equal(t, "/api/admin/login", loginRoute.Path, "Expected correct path for login route")
}
