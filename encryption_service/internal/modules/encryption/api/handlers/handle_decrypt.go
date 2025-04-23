package handlers

import (
	"context"
	pb "encryption_microservice/internal/common/proto_gen"
	"encryption_microservice/internal/modules/encryption/api/dtos"
	"encryption_microservice/internal/modules/encryption/api/mapper"
	"encryption_microservice/pkg/logger"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	response, err := h.encryptionUseCase.Decrypt(c.Context(), userData.ID, userData.EDEKPrivate, userData.EDEKPublic, &req)
	if err != nil {
		return mapper.NewErrorResponse(c, err.Error(), fiber.StatusBadRequest, err.ErrorCode)
	}

	return mapper.NewResponse(c, "Decryption Succesful", int(fiber.StatusOK), "", response.Data)
}

// HandleDecrypt handles the decryption request
func (h *EncryptionHandlerImpl) Decrypt(ctx context.Context, req *pb.DecryptRequest) (*pb.DecryptResponse, error) {
	token := req.Token

	userData, customErr := h.userService.GetUserData(token)
	if customErr != nil {
		return nil, status.Error(codes.Unauthenticated, customErr.Error())
	}

	mappedRequest := mapper.ConvertFromStructPB(req.Data)
	decryptRequest := &dtos.DecryptRequest{Data: mappedRequest}

	response, customErr := h.encryptionUseCase.Decrypt(ctx, userData.ID, userData.EDEKPrivate, userData.EDEKPublic, decryptRequest)
	if customErr != nil {
		return nil, status.Error(codes.Unavailable, customErr.Error())
	}

	mappedResponse := mapper.ConvertToStructPB(response.Data)
	return &pb.DecryptResponse{Data: mappedResponse}, nil
}
