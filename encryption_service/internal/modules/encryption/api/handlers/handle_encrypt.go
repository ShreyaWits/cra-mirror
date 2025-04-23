package handlers

import (
	"encryption_microservice/internal/modules/encryption/api/dtos"
	"encryption_microservice/internal/modules/encryption/api/mapper"
	"encryption_microservice/pkg/logger"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// HandleEncrypt handles the encryption request
func (h *EncryptionHandlerImpl) HandleEncrypt(c *fiber.Ctx) error {
	var req dtos.EncryptRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Error("Failed to parse request body", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		logger.Error("Validation failed", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Validation failed",
		})
	}
	token := c.Locals("token").(string)

	// Fetch user data and handle errors using the structured mapper
	userData, userErr := h.userService.GetUserData(token)
	if userErr != nil {
		return mapper.NewErrorResponse(c, userErr.Error(), fiber.StatusBadRequest, userErr.ErrorCode)
	}
	// Execute use case and handle errors
	response, encErr := h.encryptionUseCase.Encrypt(userData.ID, userData.EDEKPrivate, userData.EDEKPublic, &req)
	if encErr != nil {
		return mapper.NewErrorResponse(c, encErr.Error(), fiber.StatusBadRequest, encErr.ErrorCode)
	}

	return mapper.NewResponse(c, "Encryption Succesful", int(fiber.StatusOK), "", response.Data)
}
