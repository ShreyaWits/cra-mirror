package linkRoutes

import (
	"protected_link/internal/common/api/middlewares"
	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"
	"protected_link/internal/modules/link_generation/apis/handlers"
	middleware "protected_link/internal/modules/link_generation/apis/middlewares"
	"protected_link/internal/modules/link_generation/repositories"
	"protected_link/internal/modules/link_generation/services"

	database "protected_link/pkg/redis"

	"github.com/gofiber/fiber/v2"
)

func InitializeLinkRoutes(app *fiber.App, redis *database.RedisConfig, group fiber.Router) {

	repo := repositories.NewGeneratedRepository(redis)

	services := services.NewGenerateLinkService(repo, redis)

	handlers := handlers.NewGenerateLinkHandler(services)

	group.Post("/generate-link", middlewares.CommonRequestValidator(), middlewares.ValidateBodyDTO(func() interface{} {
		return &apiDtos.GenerateUrlRequest{}
	}), middleware.ValidateDTOKeys(), handlers.CreateSecureURL)

	app.Get("", middlewares.CommonRequestValidator(), handlers.GetExtractData)
	app.Delete("", middlewares.CommonRequestValidator(), handlers.DeleteGeneratedLink)
}
