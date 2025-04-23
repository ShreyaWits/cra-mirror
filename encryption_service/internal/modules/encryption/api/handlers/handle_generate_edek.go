package handlers

import (
	"encryption_microservice/internal/modules/encryption/api/mapper"

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
	if err != nil {
		return mapper.NewErrorResponse(c, err.Error(), fiber.StatusBadRequest, err.ErrorCode)
	}

	//Fixme :Remove in Production as this will be executed by UserService
	err = h.userService.UpdateUser(userData.ID, response.EDEKPrivate, response.EDEKPublic)

	if err != nil {
		return mapper.NewErrorResponse(c, err.Error(), fiber.StatusBadRequest, err.ErrorCode)
	}

	return mapper.NewResponse(c, "EDEK Generated Successfully", int(fiber.StatusOK), "", response)
}
