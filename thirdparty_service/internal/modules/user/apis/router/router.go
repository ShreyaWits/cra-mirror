package router

import (
	"thirdparty_service/internal/app"
	"thirdparty_service/internal/middleware"
	"thirdparty_service/internal/modules/user/apis/dtos"
	handlers "thirdparty_service/internal/modules/user/apis/handler"

	"github.com/gofiber/fiber/v2"
)

func SetupUserRoutes(router fiber.Router) error {
	userPrefix := router.Group("/user")

	userService := app.GetUserService()
	userHandlers := handlers.NewUserHandler(userService)

	userPrefix.Post("/", middleware.ValidatorMiddleware[dtos.CreateUserRequest](), userHandlers.HandleCreateUser)

	return nil
}
