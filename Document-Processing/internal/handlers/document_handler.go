package handlers

import (
	"context"

	"Document-Processing/internal/services"
	pb "Document-Processing/proto"
)

type DocumentHandler struct {
	pb.UnimplementedDocumentProcessingServiceServer
	documentService services.DocumentService
}

func NewDocumentHandler(documentService services.DocumentService) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
	}
}

func (h *DocumentHandler) ProcessDocument(ctx context.Context, req *pb.ProcessDocumentRequest) (*pb.ProcessDocumentResponse, error) {
	processedDoc, err := h.documentService.ProcessDocument(ctx, req.Document, req.DocumentType)
	if err != nil {
		return &pb.ProcessDocumentResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.ProcessDocumentResponse{
		Success:           true,
		Message:           "Document processed successfully",
		ProcessedDocument: processedDoc,
	}, nil
}
