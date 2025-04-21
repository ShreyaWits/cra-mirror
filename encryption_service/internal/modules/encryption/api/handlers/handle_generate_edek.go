package handlers

import (
	"encryption_microservice/pkg/errors"
	"encryption_microservice/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

// HandleGenerateEDEK handles the EDEK generation request
func (h *EncryptionHandlerImpl) HandleGenerateEDEK(c *fiber.Ctx) error {

	// Execute use case
	response, err := h.encryptionUseCase.GenerateEDEK()
	if err != nil {
		switch err.(type) {
		case *errors.BadRequestError:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		case *errors.KeyManagementError:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		case *errors.EncryptionError:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		default:
			logger.Error("Unexpected error during EDEK generation", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}
	}

	return c.JSON(response)
}
