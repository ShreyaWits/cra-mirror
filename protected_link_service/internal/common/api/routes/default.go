package routes

import (
	linkRoutes "protected_link/internal/modules/protected_link_generation/apis/routes"
	database "protected_link/pkg/redis"

	"github.com/gofiber/fiber/v2"
)

// InitializeServerRoutes initializes the server routes.
func InitializeServerRoutes(redis *database.RedisConfig, app *fiber.App) {

	userGroup := app.Group("/api/v1")

	linkRoutes.InitializeLinkRoutes(app, redis, userGroup)

}
