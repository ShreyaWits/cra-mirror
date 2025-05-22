package handlers

import (
	"context"
	"fmt"

	"Document-Processing/pkg/observability"
	pb "Document-Processing/proto"
	"go.opentelemetry.io/otel/codes"
)

// DocumentServiceHandlerInterface defines the methods used by the DocumentHandler
type DocumentServiceHandlerInterface interface {
	ProcessBatchFilesV1(ctx context.Context, req *pb.BatchFileProcessingRequest) (*pb.BatchProcessingAck, error)
	GetBatchStatusV1(ctx context.Context, req *pb.BatchStatusRequest) (*pb.BatchFileProcessingResponse, error)
}

type DocumentHandler struct {
	pb.UnimplementedDocumentProcessingServiceV1Server
	documentService DocumentServiceHandlerInterface
	observability   observability.ObservabilityStack
}

func NewDocumentHandler(documentService DocumentServiceHandlerInterface, obs observability.ObservabilityStack) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
		observability:   obs,
	}
}

// ProcessBatchFiles initiates batch processing and returns immediately with a batch ID
func (h *DocumentHandler) ProcessBatchFilesV1(ctx context.Context, req *pb.BatchFileProcessingRequest) (*pb.BatchProcessingAck, error) {
	traceCtx, span := h.observability.TracerService.StartTracer(ctx, "DocumentHandler.ProcessBatchFilesV1")
	defer span.End()

	h.observability.TracerService.SetAttributes(span, map[string]string{
		"handler":   "ProcessBatchFilesV1",
		"file.count": fmt.Sprintf("%d", len(req.Files)),
	})

	h.observability.LoggerService.Info(traceCtx, "Processing batch files", fmt.Sprintf("%d files", len(req.Files)))
	h.observability.MetricsService.IncrementCounter(traceCtx, "batch_files_requested", 1, map[string]string{
		"handler": "ProcessBatchFilesV1",
	})

	resp, err := h.documentService.ProcessBatchFilesV1(traceCtx, req)
	if err != nil {
		h.observability.LoggerService.Error(traceCtx, "Batch processing failed", err.Error())
		h.observability.MetricsService.IncrementCounter(traceCtx, "batch_files_failed", 1, nil)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	h.observability.MetricsService.IncrementCounter(traceCtx, "batch_files_success", 1, nil)
	span.SetStatus(codes.Ok, "Batch processed successfully")
	return resp, nil
}

// GetBatchStatus returns the current status of a batch processing job
func (h *DocumentHandler) GetBatchStatusV1(ctx context.Context, req *pb.BatchStatusRequest) (*pb.BatchFileProcessingResponse, error) {
	traceCtx, span := h.observability.TracerService.StartTracer(ctx, "DocumentHandler.GetBatchStatusV1")
	defer span.End()

	h.observability.TracerService.SetAttributes(span, map[string]string{
		"handler":  "GetBatchStatusV1",
		"batch_id": req.BatchId,
	})

	h.observability.LoggerService.Info(traceCtx, "Getting batch status", req.BatchId)
	h.observability.MetricsService.IncrementCounter(traceCtx, "batch_status_requested", 1, nil)

	resp, err := h.documentService.GetBatchStatusV1(traceCtx, req)
	if err != nil {
		h.observability.LoggerService.Error(traceCtx, "Failed to get batch status", err.Error())
		h.observability.MetricsService.IncrementCounter(traceCtx, "batch_status_failed", 1, nil)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	h.observability.MetricsService.IncrementCounter(traceCtx, "batch_status_success", 1, nil)
	span.SetStatus(codes.Ok, "Batch status fetched successfully")
	return resp, nil
}
