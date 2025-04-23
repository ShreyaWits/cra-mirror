package handlers

import (
	"context"
	pb "encryption_microservice/internal/common/proto_gen"
	"encryption_microservice/internal/modules/encryption/api/mapper"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	response, err := h.encryptionUseCase.GenerateEDEK(c.Context(), userData.ID)
	if err != nil {
		return mapper.NewErrorResponse(c, err.Error(), fiber.StatusBadRequest, err.ErrorCode)
	}

	//Fixme :Remove in Production as this will be executed by UserService
	h.userService.UpdateUser(userData.ID, response.EDEKPrivate, response.EDEKPublic)

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

// HandleGenerateEDEK handles the EDEK generation request
func (h *EncryptionHandlerImpl) GenerateEDEK(ctx context.Context, req *pb.GenerateEDEKRequest) (*pb.GenerateEDEKResponse, error) {
	token := req.Token

	userData, customErr := h.userService.GetUserData(token)
	if customErr != nil {
		return nil, status.Error(codes.Unauthenticated, customErr.Error())
	}

	// Execute use case
	response, customErr := h.encryptionUseCase.GenerateEDEK(ctx, userData.ID)
	if customErr != nil {
		return nil, status.Error(codes.Unavailable, customErr.Error())
	}

	fmt.Println(response, userData)

	// Fixme: Remove in Production as this will be executed by UserService
	if updateErr := h.userService.UpdateUser(userData.ID, response.EDEKPrivate, response.EDEKPublic); updateErr != nil {
		return nil, status.Error(codes.Unavailable, updateErr.Error())
	}

	return &pb.GenerateEDEKResponse{
		EDEKPrivate: response.EDEKPrivate,
		EDEKPublic:  response.EDEKPublic,
	}, nil
}
