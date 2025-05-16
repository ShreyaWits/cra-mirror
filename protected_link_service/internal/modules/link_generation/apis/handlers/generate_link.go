package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"protected_link/internal/common/constants"
	messageUtility "protected_link/internal/common/utils"
	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"
	"protected_link/internal/modules/link_generation/models"
	"protected_link/internal/modules/link_generation/services"
	pb "protected_link/pkg/grpc/proto"
	"protected_link/pkg/validation"
)

// GenerateLinkHandler handles link generation related gRPC requests
type GenerateLinkHandler struct {
	service services.GenerateLinkServiceInterface
	pb.UnimplementedLinkServiceServer
}

// NewGenerateLinkHandler creates a new instance of GenerateLinkHandler
func NewGenerateLinkHandler(service services.GenerateLinkServiceInterface) *GenerateLinkHandler {
	return &GenerateLinkHandler{
		service: service,
	}
}

func (h *GenerateLinkHandler) SaveGeneratedLinkV1(ctx context.Context, req *pb.GenerateUrlRequestV1) (*pb.GenerateUrlResponseV1, error) {
	log.Printf("📥 [gRPC] SaveGeneratedLinkV1 invoked with request: %+v", req)

	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	dataInterfaceMap := make(map[string]interface{}, len(req.Data))
	for k, v := range req.Data {
		dataInterfaceMap[k] = v
	}

	internalReq := &apiDtos.GenerateUrlRequest{
		UserID:      req.UserId,
		Name:        req.Name,
		RequestType: req.RequestType,
		ModelType:   req.ModelType,
		OtpRequired: req.OtpRequired,
		ExpireIn:    req.ExpireIn,
		Email:       req.Email,
		Phone:       req.Phone,
		ChannelType: req.ChannelType,
		Data:        apiDtos.JSONB(dataInterfaceMap),
	}

	fieldErrors, err := validation.ValidateGenerateUrlRequest(*internalReq)
	if err != nil {
		log.Printf("❌ Validation failed: %v", err)
		return &pb.GenerateUrlResponseV1{
			Success: false,
			Message: "Validation failed",
			Data:    nil,
			Error:   fieldErrors,
		}, nil
	}

	response, err := h.service.SaveGeneratedLink(internalReq)
	if err != nil {
		log.Printf("❌ Failed to save generated link: %v", err)
		return nil, fmt.Errorf("failed to save generated link: %w", err)
	}

	log.Printf("✅ Link generated successfully: %s", response.Data.(*models.ProtectedLinkResponse).URL)
	return &pb.GenerateUrlResponseV1{
		Success: response.Success,
		Message: messageUtility.GetMessage(string(constants.LinkGeneratedSuccessfully)),
		Data: &pb.ProtectedLinkResponse{
			Url: response.Data.(*models.ProtectedLinkResponse).URL,
		},
		Error: nil,
	}, nil
}

func (h *GenerateLinkHandler) DeleteGeneratedLinkV1(ctx context.Context, req *pb.DeleteGeneratedLinkRequestV1) (*pb.DeleteGeneratedLinkResponseV1, error) {
	log.Printf("📥 [gRPC] DeleteGeneratedLinkV1 invoked with request: %+v", req)

	if req == nil || req.Link == "" {
		log.Println("❌ Validation failed: link cannot be empty")
		errorMap := map[string]string{
			"validation_errors": "Link cannot be empty",
		}
		jsonErrors, _ := json.Marshal(errorMap)

		return &pb.DeleteGeneratedLinkResponseV1{
			Success: false,
			Message: "Validation failed",
			Data:    nil,
			Error:   string(jsonErrors),
		}, nil
	}

	response, err := h.service.DeleteGeneratedLink(req.Link)
	if err != nil {
		log.Printf("❌ Failed to delete link: %v", err)
		return nil, fmt.Errorf("failed to delete generated link: %w", err)
	}

	log.Printf("✅ Link deleted successfully: %s", response.Data.(*models.ProtectedLinkResponse).URL)
	return &pb.DeleteGeneratedLinkResponseV1{
		Success: response.Success,
		Message: messageUtility.GetMessage(string(constants.ProtectedLinkDeletedSuccessfully)),
		Data: &pb.ProtectedLinkResponse{
			Url: response.Data.(*models.ProtectedLinkResponse).URL,
		},
		Error: "",
	}, nil
}

func (h *GenerateLinkHandler) GetExtractDataV1(ctx context.Context, req *pb.GetExtractDataRequestV1) (*pb.GetExtractDataResponseV1, error) {
	log.Printf("📥 [gRPC] GetExtractDataV1 invoked with request: %+v", req)

	if req == nil || req.Token == "" {
		log.Println("❌ Validation failed: token cannot be empty")
		errorMap := map[string]string{
			"validation_errors": "Token cannot be empty",
		}
		jsonErrors, _ := json.Marshal(errorMap)

		return &pb.GetExtractDataResponseV1{
			Success: false,
			Message: "Validation failed",
			Data:    nil,
			Error:   string(jsonErrors),
		}, nil
	}

	result, err := h.service.GetExtractData(&req.Token)
	if err != nil {
		log.Printf("❌ Failed to retrieve token data: %v", err)
		return nil, fmt.Errorf("failed to retrieve token data: %w", err)
	}

	dataInterfaceMap := make(map[string]string)
	switch data := result.Data.(type) {
	case map[string]interface{}:
		for k, v := range data {
			dataInterfaceMap[k] = fmt.Sprintf("%v", v)
		}
	default:
		jsonBytes, err := json.Marshal(result.Data)
		if err != nil {
			log.Printf("❌ Failed to marshal result.Data: %v", err)
			dataInterfaceMap["message"] = result.Message
		} else {
			var intermediateMap map[string]interface{}
			if err := json.Unmarshal(jsonBytes, &intermediateMap); err != nil {
				log.Printf("❌ Failed to unmarshal into map: %v", err)
				dataInterfaceMap["message"] = result.Message
			} else {
				for k, v := range intermediateMap {
					dataInterfaceMap[k] = fmt.Sprintf("%v", v)
				}
			}
		}
	}

	log.Printf("✅ Data processed successfully: %+v", dataInterfaceMap)
	return &pb.GetExtractDataResponseV1{
		Success: result.Success,
		Message: messageUtility.GetMessage(string(constants.DataFetchedSuccessfully)),
		Data:    dataInterfaceMap,
		Error:   "",
	}, nil
}
