package handlers

import (
	"net/http"
	"protected_link/internal/modules/authentication/models"
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

	body := c.Locals("validatedBody").(*models.VerifyOTPRequest)

	//Call the service layer to verify the OTP
	response, err := h.authService.VerifyOTP(body)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(err)
	}

	if !response.Success {
		return c.Status(http.StatusUnauthorized).JSON(response)
	}
	// If OTP verification is successful, return the response

	return c.Status(http.StatusOK).JSON(response)
}
