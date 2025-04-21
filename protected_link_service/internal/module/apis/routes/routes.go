package linkRoutes

import (
	"protected_link/internal/common/api/middlewares"
	database "protected_link/pkg/redis"

	apiDtos "protected_link/internal/module/apis/dtos"
	"protected_link/internal/module/apis/handlers"
	"protected_link/internal/module/apis/middleware"
	"protected_link/internal/module/repositories"
	"protected_link/internal/module/services"

	"github.com/gofiber/fiber/v2"
)

func InitializeLinkRoutes(radis *database.RedisConfig, app fiber.Router) {

	repo := repositories.NewGeneratedRepository(radis)

	services := services.NewGenerateLinkService(repo)

	handlers := handlers.NewGenerateLinkHandler(services)

	app.Post("/generate-link", middlewares.CommonRequestValidator(), middlewares.ValidateBodyDTO(&apiDtos.GenerateUrlRequest{}), middleware.ValidateDTOKeys(), handlers.GetProtectedURL)

}
