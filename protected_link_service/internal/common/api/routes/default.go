package routes

import (
	linkRoutes "protected_link/internal/module/apis/routes"
	database "protected_link/pkg/redis"

	"github.com/gofiber/fiber/v2"
)

// InitializeServerRoutes initializes the server routes.
func InitializeServerRoutes(radis *database.RedisConfig, app *fiber.App) {

	userGroup := app.Group("/api/v1")
	println("sadadsasd", userGroup)
	linkRoutes.InitializeLinkRoutes(radis, userGroup)

}
