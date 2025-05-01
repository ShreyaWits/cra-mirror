package handlers

import (
	"fmt"
	"net/http"
	commonDtos "protected_link/internal/common/api/dtos"
	"protected_link/internal/common/constants"
	"protected_link/internal/common/utils"
	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"
	"protected_link/internal/modules/link_generation/services"
	"protected_link/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

type GenerateLinkHandler struct {
	services *services.GenerateLinkService
}

// NewGenerateLinkHandler initializes the handler
func NewGenerateLinkHandler(services *services.GenerateLinkService) *GenerateLinkHandler {
	logger.InitLogger()
	return &GenerateLinkHandler{
		services: services,
	}
}

// CreateSecureURL handles generation and retrieval of the protected link5
func (h *GenerateLinkHandler) CreateSecureURL(c *fiber.Ctx) error {
	body := c.Locals("validatedBody").(*apiDtos.GenerateUrlRequest)

	// Save link using the service
	savedLink, err := h.services.SaveGeneratedLink(body)
	logger.Error("GENRATE LINK ERROR", err)
	fmt.Println("Error in Generate Link Handler:", err)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(commonDtos.ApiResponseDto{
			Success: false,
			Message: utils.GetMessage(string(constants.FaliedToSaveLink)),
			Data:    err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(savedLink)
}

func (h *GenerateLinkHandler) DeleteGeneratedLink(c *fiber.Ctx) error {
	token := c.Query("token")

	// Save link using the service
	savedLink, err := h.services.DeleteGeneratedLink(token)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(commonDtos.ApiResponseDto{
			Success: false,
			Message: "Failed to delete link",
			Data:    err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(savedLink)
}

func (h *GenerateLinkHandler) GetExtractData(c *fiber.Ctx) error {
	token := c.Query("token") // Extract token from query param

	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(commonDtos.ApiResponseDto{
			Success: false,
			Message: utils.GetMessage(string(constants.AuthTokenMissing)),
			Error:   utils.GetMessage(string(constants.AuthTokenMissing)),
		})
	}

	// Save link using the service
	savedLink, err := h.services.GetExtractData(&token)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(commonDtos.ApiResponseDto{
			Success: false,
			Message: utils.GetMessage(string(constants.RequestLinkExpiredTitle)),
			Error:   utils.GetMessage(string(constants.RequestLinkExpiredDesc)),
		})
	}

	return c.Status(http.StatusOK).JSON(savedLink)
}
