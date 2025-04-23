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
	response, encErr := h.encryptionUseCase.Encrypt(c.Context(), userData.ID, userData.EDEKPrivate, userData.EDEKPublic, &req)
	if encErr != nil {
		return mapper.NewErrorResponse(c, encErr.Error(), fiber.StatusBadRequest, encErr.ErrorCode)
	}

	return mapper.NewResponse(c, "Encryption Succesful", int(fiber.StatusOK), "", response.Data)
}

func (h *EncryptionHandlerImpl) Encrypt(ctx context.Context, req *pb.EncryptDataRequest) (*pb.EncryptDataResponse, error) {
	token := req.Token

	userData, customErr := h.userService.GetUserData(token)
	if customErr != nil {
		return nil, status.Error(codes.Unauthenticated, customErr.Error())
	}

	mappedRequest := mapper.ConvertFromStructPB(req.Data)

	encryptedReq := &dtos.EncryptRequest{
		Data: mappedRequest,
	}

	// Execute use case
	response, customErr := h.encryptionUseCase.Encrypt(ctx, userData.ID, userData.EDEKPrivate, userData.EDEKPublic, encryptedReq)
	if customErr != nil {
		return nil, status.Error(codes.Unavailable, customErr.Error())
	}

	mappedResponse := mapper.ConvertToStructPB(response.Data)
	return &pb.EncryptDataResponse{
		Data: mappedResponse,
	}, nil
}
