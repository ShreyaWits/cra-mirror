package handlers

import (
	"third_party_service/internal/modules/third_party/services"

	"github.com/gofiber/fiber/v2"
)

type ThirdPartyHandler struct {
	UserService *services.ThirdPartyService
}

func NewThirdPartyHandler(userService *services.ThirdPartyService) *ThirdPartyHandler {
	return &ThirdPartyHandler{
		UserService: userService,
	}
}

func (uh *ThirdPartyHandler) HandlerSendEmail(c *fiber.Ctx) error {

	return nil
}
