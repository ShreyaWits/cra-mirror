package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"Document-Processing/internal/enums"
	"Document-Processing/internal/models"
	"Document-Processing/pkg/observability"
	pb "Document-Processing/proto"
)

// BatchProcessingState tracks the state of a batch processing job

type GeminiServiceInterface interface {
	ProcessImage(ctx context.Context, base64Image string, extractionFields []string, fileUrl string) (map[string]string, float32, error)
	Close() error
}

type LlamaServiceInterface interface {
	ProcessImage(ctx context.Context, base64Image string, extractionFields []string) (map[string]string, error)
}

type MinioRepositoryInterfaces interface {
	StoreFile(ctx context.Context, data []byte, fileType string) (string, error)
	DeleteFile(ctx context.Context, fileUrl string) error
}

type DocumentDataRepository interface {
	CreateDocumentData(data *models.DocumentData) (*models.DocumentData, error)
	GetDocumentDataByID(id string) (*models.DocumentData, error)
}

type BatchProcessingState struct {
	mu             sync.RWMutex
	Status         string
	Message        string
	Error          error
	ProcessedFiles []*pb.ProcessedFileData
}

type DocumentService struct {
	pb.UnimplementedDocumentProcessingServiceV1Server
	geminiService GeminiServiceInterface
	llamaService  LlamaServiceInterface
	minioRepo     MinioRepositoryInterfaces
	yugabyteRepo  DocumentDataRepository
	batchStates   map[string]*BatchProcessingState
	batchStatesMu sync.RWMutex
	observability	observability.ObservabilityStack
}

func NewDocumentService(
	geminiService GeminiServiceInterface,
	llamaService LlamaServiceInterface,
	minioRepo MinioRepositoryInterfaces,
	yugabyteRepo DocumentDataRepository,
	observability observability.ObservabilityStack,
) *DocumentService {
	return &DocumentService{
		geminiService: geminiService,
		llamaService:  llamaService,
		minioRepo:     minioRepo,
		yugabyteRepo:  yugabyteRepo,
		batchStates:   make(map[string]*BatchProcessingState),
		observability: observability,
	}
}

// ProcessBatchFilesV1 initiates batch processing and returns immediately with a batch ID
func (s *DocumentService) ProcessBatchFilesV1(ctx context.Context, req *pb.BatchFileProcessingRequest) (*pb.BatchProcessingAck, error) {
	log.Printf("ProcessBatchFilesV1 called with %d files", len(req.Files))
	traceCtx, span := s.observability.TracerService.StartTracer(ctx, "DocumentService.ProcessBatchFilesV1")
	defer span.End()

	s.observability.TracerService.SetAttributes(span, map[string]string{
		"handler":   "ProcessBatchFilesV1",
		"file.count": fmt.Sprintf("%d", len(req.Files)),
	})
	s.observability.LoggerService.Info(traceCtx, "Processing batch files", fmt.Sprintf("%d files", len(req.Files)))
	s.observability.MetricsService.IncrementCounter(traceCtx, "batch_files_requested", 1, map[string]string{
		"handler": "ProcessBatchFilesV1",
	})

	if len(req.Files) == 0 {
		s.observability.LoggerService.Error(traceCtx, "No files provided in request", "ProcessBatchFilesV1")
		return nil, fmt.Errorf("no files provided in request")
	}

	// Use provided batch_id or generate one if not provided
	batchID := req.BatchId
	if batchID == "" {
		batchID = fmt.Sprintf("batch-%d", time.Now().UnixNano())
	}

	// Store all files in Minio and collect their URLs
	var fileURLs []string
	for _, fileReq := range req.Files {
		if fileReq.File == nil {
			continue
		}

		fileData, err := base64.StdEncoding.DecodeString(fileReq.File.Base64File)
		if err != nil {
			s.observability.LoggerService.Error(traceCtx, "Failed to decode base64 data", err.Error())
			s.observability.MetricsService.IncrementCounter(traceCtx, "base64_decoding_failed", 1, nil)
			span.RecordError(err)
			return nil, fmt.Errorf("failed to decode base64 data: %v", err)
		}

		fileURL, err := s.minioRepo.StoreFile(ctx, fileData, fileReq.File.FileType)
		if err != nil {
			s.observability.LoggerService.Error(traceCtx, "Failed to store file in Minio", err.Error())
			s.observability.MetricsService.IncrementCounter(traceCtx, "minio_storage_failed", 1, nil)
			span.RecordError(err)
			return nil, fmt.Errorf("failed to store file in Minio: %v", err)
		}

		fileURLs = append(fileURLs, fileURL)
	}

	// Create and store the batch state
	batchState := &BatchProcessingState{
		Status:  "ACCEPTED",
		Message: "Batch processing accepted",
	}

	s.batchStatesMu.Lock()
	s.batchStates[batchID] = batchState
	s.batchStatesMu.Unlock()

	// Start processing in the background
	go s.processBatchInBackground(ctx, batchID, req, fileURLs)
	s.observability.LoggerService.Info(traceCtx, "Batch processing started", batchID)
	s.observability.MetricsService.IncrementCounter(traceCtx, "batch_processing_started", 1, map[string]string{
		"handler": "ProcessBatchFilesV1",
		"batch_id": batchID,
	})
	s.observability.TracerService.SetAttributes(span, map[string]string{
		"handler":   "ProcessBatchFilesV1",
		"batch_id":  batchID,
		"file.count": fmt.Sprintf("%d", len(req.Files)),
	})
	return &pb.BatchProcessingAck{
		BatchId:  batchID,
		Status:   "ACCEPTED",
		Message:  "Batch processing started",
		FileUrls: fileURLs,
	}, nil
}

// GetBatchStatusV1 returns the current status of a batch processing job
func (s *DocumentService) GetBatchStatusV1(ctx context.Context, req *pb.BatchStatusRequest) (*pb.BatchFileProcessingResponse, error) {
	log.Printf("GetBatchStatusV1 called for batch_id: %s", req.BatchId)

	traceCtx, span := s.observability.TracerService.StartTracer(ctx, "DocumentService.GetBatchStatusV1")
	defer span.End()

	s.observability.TracerService.SetAttributes(span, map[string]string{
		"handler":  "GetBatchStatusV1",
		"batch_id": req.BatchId,
	})
	s.observability.LoggerService.Info(traceCtx, "Retrieving batch status", req.BatchId)
	s.observability.MetricsService.IncrementCounter(traceCtx, "batch_status_requested", 1, map[string]string{
		"handler":  "GetBatchStatusV1",
		"batch_id": req.BatchId,
	})

	data, err := s.yugabyteRepo.GetDocumentDataByID(req.BatchId)
	if err != nil {
		s.observability.LoggerService.Error(traceCtx, "Failed to retrieve batch data from Yugabyte", err.Error())
		s.observability.MetricsService.IncrementCounter(traceCtx, "yugabyte_query_failed", 1, map[string]string{
			"batch_id": req.BatchId,
		})
		span.RecordError(err)

		return &pb.BatchFileProcessingResponse{
			Success: false,
			Message: "Failed to retrieve batch data.",
		}, nil
	}

	if data == nil {
		s.observability.LoggerService.Warn(traceCtx, "No data found or still processing", req.BatchId)
		s.observability.MetricsService.IncrementCounter(traceCtx, "batch_processing_pending", 1, map[string]string{
			"batch_id": req.BatchId,
		})

		return &pb.BatchFileProcessingResponse{
			Success: false,
			Message: "Document is currently being processed.",
		}, nil
	}

	// decoding JSON
	var decodeExtractedData []map[string]interface{}
	decodeErr := json.Unmarshal([]byte(data.Data), &decodeExtractedData)
	if decodeErr != nil {
		s.observability.LoggerService.Error(traceCtx, "Failed to decode JSON from DB", decodeErr.Error())
		s.observability.MetricsService.IncrementCounter(traceCtx, "json_unmarshal_failed", 1, map[string]string{
			"batch_id": req.BatchId,
		})
		span.RecordError(decodeErr)

		return &pb.BatchFileProcessingResponse{
			Success: false,
			Message: "Internal error: Failed to decode processed data: " + decodeErr.Error(),
		}, nil
	}

	// converting map to proto
	var processedFiles []*pb.ProcessedFileData

	for _, item := range decodeExtractedData {
		processedFile := &pb.ProcessedFileData{
			FileId:     item["file_id"].(string),
			FileUrl:    item["file_url"].(string),
			Confidence: float32(item["confidence"].(float64)),
		}

		// Handle extracted_data: Convert it to map[string]string
		extractedData := map[string]string{}
		if extractedDataRaw, ok := item["extracted_data"].(map[string]interface{}); ok {
			for key, value := range extractedDataRaw {
				extractedData[key] = value.(string)
			}
		}
		processedFile.ExtractedData = extractedData
		processedFiles = append(processedFiles, processedFile)
	}

	s.observability.LoggerService.Info(traceCtx, "Batch processing completed successfully", req.BatchId)
	s.observability.MetricsService.IncrementCounter(traceCtx, "batch_status_success", 1, map[string]string{
		"batch_id": req.BatchId,
	})

	return &pb.BatchFileProcessingResponse{
		Success:        data.Status == "COMPLETED",
		Message:        data.Message,
		ProcessedFiles: processedFiles,
		CreatedAt:      data.CreatedAt.Format(time.RFC3339),
	}, nil
}


// processBatchInBackground handles the actual file processing
func (s *DocumentService) processBatchInBackground(ctx context.Context, batchID string, req *pb.BatchFileProcessingRequest, fileURLs []string) {
	traceCtx, span := s.observability.TracerService.StartTracer(ctx, "DocumentService.processBatchInBackground")
	defer span.End()

	s.observability.TracerService.SetAttributes(span, map[string]string{
		"handler":  "processBatchInBackground",
		"batch_id": batchID,
	})
	s.observability.LoggerService.Info(traceCtx, "Started background batch processing", fmt.Sprintf("batchID: %s", batchID))
	s.observability.MetricsService.IncrementCounter(traceCtx, "background_batch_processing_started", 1, map[string]string{
		"batch_id": batchID,
	})

	s.batchStatesMu.RLock()
	batchState := s.batchStates[batchID]
	s.batchStatesMu.RUnlock()

	batchState.mu.Lock()
	batchState.Status = "PROCESSING"
	batchState.Message = "Processing files"
	batchState.mu.Unlock()

	// Create a new context with a longer timeout for the entire batch
	batchCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var wg sync.WaitGroup
	errChan := make(chan error, len(req.Files))
	processedFileChan := make(chan *pb.ProcessedFileData, len(req.Files))

	for i, fileReq := range req.Files {
		wg.Add(1)
		go func(fileReq *pb.FileProcessingRequest, index int) {
			defer wg.Done()

			fileCtx, fileCancel := context.WithTimeout(batchCtx, 2*time.Minute)
			defer fileCancel()

			s.observability.LoggerService.Info(traceCtx, "Processing file", fmt.Sprintf("Index: %d, FileType: %s", index, fileReq.File.FileType))

			processedFile, err := s.processFile(fileCtx, fileReq, req.Classifier, fileURLs[index])
			if err != nil {
				s.observability.LoggerService.Error(traceCtx, fmt.Sprintf("Failed to process file %d", index), err.Error())
				s.observability.MetricsService.IncrementCounter(traceCtx, "file_processing_failed", 1, map[string]string{
					"file_index": fmt.Sprintf("%d", index),
					"batch_id":   batchID,
				})
				span.RecordError(err)
				errChan <- err
				return
			}

			s.observability.LoggerService.Info(traceCtx, "File processed successfully", fmt.Sprintf("Index: %d", index))
			processedFileChan <- processedFile
		}(fileReq, i)
	}

	wg.Wait()
	close(errChan)
	close(processedFileChan)

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	var processedFiles []*pb.ProcessedFileData
	for file := range processedFileChan {
		processedFiles = append(processedFiles, file)
	}

	batchState.mu.Lock()
	if len(errors) > 0 {
		batchState.Status = "FAILED"
		batchState.Message = fmt.Sprintf("Error processing files: %v", errors[0])
		batchState.Error = errors[0]

		s.observability.LoggerService.Error(traceCtx, "Batch processing failed", errors[0].Error())
		s.observability.MetricsService.IncrementCounter(traceCtx, "batch_processing_failed", 1, map[string]string{
			"batch_id": batchID,
		})
		span.RecordError(errors[0])
	} else {
		batchState.Status = "COMPLETED"
		batchState.Message = "All files processed successfully"
		batchState.ProcessedFiles = processedFiles

		s.observability.LoggerService.Info(traceCtx, "Batch processing completed successfully", batchID)
		s.observability.MetricsService.IncrementCounter(traceCtx, "batch_processing_completed", 1, map[string]string{
			"batch_id": batchID,
		})
	}

	var extractedDataJSON []map[string]interface{}
	for _, f := range processedFiles {
		item := map[string]interface{}{
			"file_id":        f.FileId,
			"file_url":       f.FileUrl,
			"extracted_data": f.ExtractedData,
			"confidence":     f.Confidence,
		}
		extractedDataJSON = append(extractedDataJSON, item)
	}

	jsonBytes, jsonErr := json.Marshal(extractedDataJSON)
	if jsonErr != nil {
		s.observability.LoggerService.Error(traceCtx, "Failed to marshal extracted data to JSON", jsonErr.Error())
		span.RecordError(jsonErr)
	} else {
		extractedData := &models.DocumentData{
			Batch_id:  batchID,
			Data:      string(jsonBytes),
			Status:    batchState.Status,
			Message:   batchState.Message,
			CreatedAt: time.Now(),
		}

		_, err := s.yugabyteRepo.CreateDocumentData(extractedData)
		if err != nil {
			s.observability.LoggerService.Error(traceCtx, "Error saving processed data to DB", err.Error())
			s.observability.MetricsService.IncrementCounter(traceCtx, "db_write_failed", 1, map[string]string{
				"batch_id": batchID,
			})
			span.RecordError(err)
			batchState.Error = fmt.Errorf("failed to save processed data to DB: %v", err)
		} else {
			s.observability.LoggerService.Info(traceCtx, "Processed data saved to DB successfully", batchID)
			s.observability.MetricsService.IncrementCounter(traceCtx, "db_write_success", 1, map[string]string{
				"batch_id": batchID,
			})
		}
	}

	batchState.mu.Unlock()
}

// GeminiService returns the Gemini service instance
func (s *DocumentService) GeminiService() GeminiServiceInterface {
	return s.geminiService
}

func (s *DocumentService) processFile(ctx context.Context, req *pb.FileProcessingRequest, classifier string, fileUrl string) (*pb.ProcessedFileData, error) {
	traceCtx, span := s.observability.TracerService.StartTracer(ctx, "DocumentService.processFile")
	defer span.End()

	s.observability.TracerService.SetAttributes(span, map[string]string{
		"handler":  "processFile",
		"fileUrl":  fileUrl,
		"mimeType": req.File.FileType,
	})

	base64Data := req.File.Base64File
	mimeType := req.File.FileType

	s.observability.LoggerService.Info(traceCtx, "Received base64 data", fmt.Sprintf("length=%d, mimeType=%s", len(base64Data), mimeType))

	// Handle data URL format
	if strings.HasPrefix(base64Data, "data:") {
		parts := strings.SplitN(base64Data, ",", 2)
		if len(parts) != 2 {
			err := fmt.Errorf("invalid data URL format")
			s.observability.LoggerService.Error(traceCtx, err.Error(), "processFile")
			s.observability.MetricsService.IncrementCounter(traceCtx, "process_file.error", 1, map[string]string{"reason": "invalid_data_url"})
			span.RecordError(err)
			return nil, err
		}
		mimeParts := strings.Split(parts[0], ";")
		if len(mimeParts) > 0 {
			mimeType = strings.TrimPrefix(mimeParts[0], "data:")
		}
		base64Data = parts[1]
		s.observability.LoggerService.Debug(traceCtx, "Extracted base64 data from data URL", fmt.Sprintf("new length=%d", len(base64Data)))
	}

	// Clean base64 string
	base64Data = strings.TrimSpace(base64Data)
	base64Data = strings.ReplaceAll(base64Data, "\n", "")
	base64Data = strings.ReplaceAll(base64Data, "\r", "")
	base64Data = strings.ReplaceAll(base64Data, " ", "")

	// Add padding if necessary
	if mod := len(base64Data) % 4; mod != 0 {
		base64Data += strings.Repeat("=", 4-mod)
	}

	s.observability.LoggerService.Debug(traceCtx, "Cleaned base64 data", fmt.Sprintf("length=%d", len(base64Data)))

	if len(base64Data) == 0 {
		err := fmt.Errorf("empty base64 data")
		s.observability.LoggerService.Error(traceCtx, err.Error(), "processFile")
		// Attempt delete file
		if delErr := s.minioRepo.DeleteFile(traceCtx, fileUrl); delErr != nil {
			s.observability.LoggerService.Warn(traceCtx, "Failed to delete file from MinIO", delErr.Error())
		}
		s.observability.MetricsService.IncrementCounter(traceCtx, "process_file.error", 1, map[string]string{"reason": "empty_base64"})
		span.RecordError(err)
		return nil, err
	}

	// Log first 100 chars if long
	if len(base64Data) > 100 {
		s.observability.LoggerService.Debug(traceCtx, "Base64 data (first 100 chars)", base64Data[:100])
	} else {
		s.observability.LoggerService.Debug(traceCtx, "Base64 data", base64Data)
	}

	// Validate base64 characters
	for i, c := range base64Data {
		if !IsValidBase64Char(c) {
			err := fmt.Errorf("invalid base64 character at position %d: %c", i, c)
			s.observability.LoggerService.Error(traceCtx, err.Error(), "processFile")
			// Attempt delete file
			if delErr := s.minioRepo.DeleteFile(traceCtx, fileUrl); delErr != nil {
				s.observability.LoggerService.Warn(traceCtx, "Failed to delete file from MinIO", delErr.Error())
			}
			s.observability.MetricsService.IncrementCounter(traceCtx, "process_file.error", 1, map[string]string{"reason": "invalid_base64_char"})
			span.RecordError(err)
			return nil, err
		}
	}

	// Decode base64
	fileData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		s.observability.LoggerService.Error(traceCtx, "Base64 decoding failed", err.Error())
		s.observability.MetricsService.IncrementCounter(traceCtx, "process_file.error", 1, map[string]string{"reason": "decode_failure"})
		span.RecordError(err)
		// Attempt delete file
		if delErr := s.minioRepo.DeleteFile(traceCtx, fileUrl); delErr != nil {
			s.observability.LoggerService.Warn(traceCtx, "Failed to delete file from MinIO", delErr.Error())
		}
		return nil, fmt.Errorf("failed to decode base64 data: %v", err)
	}

	s.observability.LoggerService.Info(traceCtx, "Successfully decoded base64 data", fmt.Sprintf("file size=%d bytes", len(fileData)))

	var extractedData map[string]string
	var confidence float32

	if classifier == string(enums.Llama) {
		extractedData, err = s.llamaService.ProcessImage(traceCtx, base64Data, req.ExtractionFields)
		if err != nil {
			s.observability.LoggerService.Error(traceCtx, "LLaMA processing failed", err.Error())
			s.observability.MetricsService.IncrementCounter(traceCtx, "process_file.error", 1, map[string]string{"reason": "llama_processing_failed"})
			span.RecordError(err)
			if delErr := s.minioRepo.DeleteFile(traceCtx, fileUrl); delErr != nil {
				s.observability.LoggerService.Warn(traceCtx, "Failed to delete file from MinIO", delErr.Error())
			}
			return nil, fmt.Errorf("failed to process with LLaMA: %v", err)
		}
		confidence = 100.0 // default confidence for LLaMA
	} else {
		extractedData, confidence, err = s.geminiService.ProcessImage(traceCtx, base64Data, req.ExtractionFields, fileUrl)
		if err != nil {
			s.observability.LoggerService.Error(traceCtx, "Gemini processing failed", err.Error())
			s.observability.MetricsService.IncrementCounter(traceCtx, "process_file.error", 1, map[string]string{"reason": "gemini_processing_failed"})
			span.RecordError(err)
			// Uncomment if you want to delete file on Gemini failure
			// if delErr := s.minioRepo.DeleteFile(traceCtx, fileUrl); delErr != nil {
			// 	s.observability.LoggerService.Warn(traceCtx, "Failed to delete file from MinIO", delErr.Error())
			// }
			return nil, fmt.Errorf("failed to process with Gemini: %v", err)
		}
	}

	s.observability.LoggerService.Info(traceCtx, "File processed successfully", fmt.Sprintf("confidence=%.2f", confidence))

	return &pb.ProcessedFileData{
		FileId:        fmt.Sprintf("file-%d", time.Now().UnixNano()),
		FileUrl:       fileUrl,
		ExtractedData: extractedData,
		Confidence:    confidence,
	}, nil
}

// IsValidBase64Char checks if a character is valid in base64 encoding
func IsValidBase64Char(c rune) bool {
	return (c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9') ||
		c == '+' || c == '/' || c == '='
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *DocumentService) GetBatchStateForTest(batchID string) (*BatchProcessingState, bool) {
	s.batchStatesMu.RLock()
	defer s.batchStatesMu.RUnlock()
	state, ok := s.batchStates[batchID]
	return state, ok
}