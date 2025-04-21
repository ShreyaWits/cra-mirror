package handlers

import (
	"encryption_microservice/internal/modules/encryption/api/dtos"
	"encryption_microservice/pkg/errors"
	"encryption_microservice/pkg/logger"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// HandleDecrypt handles the decryption request
func (h *EncryptionHandlerImpl) HandleDecrypt(c *fiber.Ctx) error {
	var req dtos.DecryptRequest
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

	userData, err := h.userService.GetUserData(token)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	// Execute use case
	response, err := h.encryptionUseCase.Decrypt(userData.ID, userData.EDEKPrivate, userData.EDEKPublic, &req)
	if err != nil {
		switch err.(type) {
		case *errors.BadRequestError:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		case *errors.EncryptionError:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		default:
			logger.Error("Unexpected error during decryption", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}
	}

	return c.JSON(response)
}
