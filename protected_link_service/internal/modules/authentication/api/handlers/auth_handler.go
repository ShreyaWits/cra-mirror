package handlers

import (
	"net/http"
	commonDtos "protected_link/internal/common/api/dtos"
	authModels "protected_link/internal/modules/authentication/models"
	authService "protected_link/internal/modules/authentication/services"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService *authService.AuthenticationService
}

// NewAuthHandler initializes the auth handler with its service
func NewAuthHandler(authService *authService.AuthenticationService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) VerifyOTPHandler(c *fiber.Ctx) error {
	body := new(authModels.VerifyOTPRequest)
	if err := c.BodyParser(body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(commonDtos.ApiResponseDto{
			Success: false,
			Message: "Invalid request payload",
			Data:    err.Error(),
		})
	}

	//Call the service layer to verify the OTP
	response, err := h.authService.VerifyOTP(body)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(commonDtos.ApiResponseDto{
			Success: false,
			Message: "OTP verification failed",
			Data:    err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(response)
}
