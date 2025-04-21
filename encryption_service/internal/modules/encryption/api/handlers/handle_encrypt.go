package handlers

import (
	"encryption_microservice/internal/modules/encryption/api/dtos"
	"encryption_microservice/pkg/errors"
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

	// Execute use case
	response, err := h.encryptionUseCase.Encrypt(&req)
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
			logger.Error("Unexpected error during encryption", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}
	}

	return c.JSON(response)
}
