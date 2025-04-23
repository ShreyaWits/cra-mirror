package apiRoutes

import (
	"protected_link/internal/common/api/middlewares"

	authHandler "protected_link/internal/modules/authentication/api/handlers"
	"protected_link/internal/modules/authentication/models"
	authRepository "protected_link/internal/modules/authentication/repositories"
	authService "protected_link/internal/modules/authentication/services"

	database "protected_link/pkg/redis"

	"github.com/gofiber/fiber/v2"
)

func InitializeLinkRoutes(app *fiber.App, redis *database.RedisConfig, group fiber.Router) {

	repo := authRepository.NewOTPRepository(redis)

	services := authService.NewAuthenticationService(repo)

	handlers := authHandler.NewAuthHandler(services)

	//middlewares.ValidateBodyDTO(&apiDtos.GenerateUrlRequest{})

	group.Post("/verify-otp", middlewares.CommonRequestValidator(), middlewares.ValidateBodyDTO(&models.VerifyOTPRequest{}), handlers.VerifyOTPHandler)

}
