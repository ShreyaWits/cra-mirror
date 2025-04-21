package handlers

import (
	"encryption_microservice/pkg/errors"
	"encryption_microservice/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

// HandleGenerateEDEK handles the EDEK generation request
func (h *EncryptionHandlerImpl) HandleGenerateEDEK(c *fiber.Ctx) error {

	token := c.Locals("token").(string)

	userData, err := h.userService.GetUserData(token)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Execute use case
	response, err := h.encryptionUseCase.GenerateEDEK(userData.ID)

	//Fixme :Remove in Production as this will be executed by UserService
	h.userService.UpdateUser(userData.ID, response.EDEKPrivate, response.EDEKPublic)

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
