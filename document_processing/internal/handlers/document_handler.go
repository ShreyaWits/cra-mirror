package handlers

import (
	"context"
	"log"

	pb "Document-Processing/proto"
)

// DocumentServiceHandlerInterface defines the methods used by the DocumentHandler
type DocumentServiceHandlerInterface interface {
	ProcessBatchFilesV1(ctx context.Context, req *pb.BatchFileProcessingRequest) (*pb.BatchProcessingAck, error)
	GetBatchStatusV1(ctx context.Context, req *pb.BatchStatusRequest) (*pb.BatchFileProcessingResponse, error)
}

type DocumentHandler struct {
	pb.UnimplementedDocumentProcessingServiceV1Server
	documentService DocumentServiceHandlerInterface
}

func NewDocumentHandler(documentService DocumentServiceHandlerInterface) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
	}
}

// ProcessBatchFiles initiates batch processing and returns immediately with a batch ID
func (h *DocumentHandler) ProcessBatchFilesV1(ctx context.Context, req *pb.BatchFileProcessingRequest) (*pb.BatchProcessingAck, error) {
	log.Printf("Processing batch request with %d files", len(req.Files))
	return h.documentService.ProcessBatchFilesV1(ctx, req)
}

// GetBatchStatus returns the current status of a batch processing job
func (h *DocumentHandler) GetBatchStatusV1(ctx context.Context, req *pb.BatchStatusRequest) (*pb.BatchFileProcessingResponse, error) {
	log.Printf("Getting status for batch: %s", req.BatchId)
	return h.documentService.GetBatchStatusV1(ctx, req)
}
