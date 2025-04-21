package handlers

import (
	"net/http"

	commonDtos "protected_link/internal/common/api/dtos"
	apiDtos "protected_link/internal/module/apis/dtos"
	"protected_link/internal/module/services"

	"github.com/gofiber/fiber/v2"
)

type GenerateLinkHandler struct {
	services *services.GenerateLinkService
}

// NewGenerateLinkHandler initializes the handler
func NewGenerateLinkHandler(services *services.GenerateLinkService) *GenerateLinkHandler {
	return &GenerateLinkHandler{
		services: services,
	}
}

// GetProtectedURL handles generation and retrieval of the protected link
func (h *GenerateLinkHandler) GetProtectedURL(c *fiber.Ctx) error {
	body := c.Locals("validatedBody").(*apiDtos.GenerateUrlRequest)

	// Save link using the service
	savedLink, err := h.services.SaveGeneratedLink(body)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(commonDtos.ApiResponseDto{
			Success: false,
			Message: "Failed to save link",
			Data:    err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(savedLink)
}
