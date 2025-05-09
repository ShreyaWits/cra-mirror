package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"Document-Processing/internal/repository"
	"Document-Processing/internal/utils"
	pb "Document-Processing/proto"
)

// BatchProcessingState tracks the state of a batch processing job
type BatchProcessingState struct {
	Status         string
	Message        string
	ProcessedFiles []*pb.ProcessedFileData
	Error          error
	mu             sync.RWMutex
}

type DocumentService struct {
	pb.UnimplementedDocumentProcessingServiceV1Server
	geminiService *GeminiService
	minioRepo     *repository.MinioRepository
	// Map to store batch processing states
	batchStates   map[string]*BatchProcessingState
	batchStatesMu sync.RWMutex
}

func NewDocumentService(geminiService *GeminiService, minioRepo *repository.MinioRepository) *DocumentService {
	return &DocumentService{
		geminiService: geminiService,
		minioRepo:     minioRepo,
		batchStates:   make(map[string]*BatchProcessingState),
	}
}

// ProcessBatchFilesV1 initiates batch processing and returns immediately with a batch ID
func (s *DocumentService) ProcessBatchFilesV1(ctx context.Context, req *pb.BatchFileProcessingRequest) (*pb.BatchProcessingAck, error) {
	log.Printf("ProcessBatchFilesV1 called with %d files", len(req.Files))

	if len(req.Files) == 0 {
		return nil, fmt.Errorf("no files provided in request")
	}

	// Use provided batch_id or generate one if not provided
	batchID := req.BatchId
	if batchID == "" {
		batchID = fmt.Sprintf("batch-%d", time.Now().UnixNano())
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
	go s.processBatchInBackground(ctx, batchID, req)

	return &pb.BatchProcessingAck{
		BatchId: batchID,
		Status:  "ACCEPTED",
		Message: "Batch processing started",
	}, nil
}

// GetBatchStatusV1 returns the current status of a batch processing job
func (s *DocumentService) GetBatchStatusV1(ctx context.Context, req *pb.BatchStatusRequest) (*pb.BatchFileProcessingResponse, error) {
	log.Printf("GetBatchStatusV1 called for batch: %s", req.BatchId)

	s.batchStatesMu.RLock()
	batchState, exists := s.batchStates[req.BatchId]
	s.batchStatesMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("batch ID not found: %s", req.BatchId)
	}

	batchState.mu.RLock()
	defer batchState.mu.RUnlock()

	if batchState.Error != nil {
		return &pb.BatchFileProcessingResponse{
			Success: false,
			Message: fmt.Sprintf("Error processing batch: %v", batchState.Error),
		}, nil
	}

	return &pb.BatchFileProcessingResponse{
		Success:        batchState.Status == "COMPLETED",
		Message:        batchState.Message,
		ProcessedFiles: batchState.ProcessedFiles,
	}, nil
}

// processBatchInBackground handles the actual file processing
func (s *DocumentService) processBatchInBackground(ctx context.Context, batchID string, req *pb.BatchFileProcessingRequest) {
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
			log.Printf("Starting to process file %d with type: %s", index, fileReq.File.FileType)

			// Create a new context for each file with a timeout
			fileCtx, fileCancel := context.WithTimeout(batchCtx, 2*time.Minute)
			defer fileCancel()

			processedFile, err := s.processFile(fileCtx, fileReq)
			if err != nil {
				log.Printf("Error processing file %d: %v", index, err)
				errChan <- err
				return
			}

			processedFileChan <- processedFile
			log.Printf("Successfully processed file %d", index)
		}(fileReq, i)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errChan)
	close(processedFileChan)

	// Collect results
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
	} else {
		batchState.Status = "COMPLETED"
		batchState.Message = "All files processed successfully"
		batchState.ProcessedFiles = processedFiles
	}
	batchState.mu.Unlock()
}

// GeminiService returns the Gemini service instance
func (s *DocumentService) GeminiService() *GeminiService {
	return s.geminiService
}

func (s *DocumentService) processFile(ctx context.Context, req *pb.FileProcessingRequest) (*pb.ProcessedFileData, error) {
	base64Data := req.File.Base64File
	mimeType := req.File.FileType

	log.Printf("Received base64 data length: %d", len(base64Data))
	log.Printf("Received MIME type: %s", mimeType)

	// Handle data URL format first
	if strings.HasPrefix(base64Data, "data:") {
		parts := strings.SplitN(base64Data, ",", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid data URL format: %s", req.File.Base64File)
		}
		mimeParts := strings.Split(parts[0], ";")
		if len(mimeParts) > 0 {
			mimeType = strings.TrimPrefix(mimeParts[0], "data:")
		}
		base64Data = parts[1]
		log.Printf("Extracted base64 data from data URL, new length: %d", len(base64Data))
	}

	// Clean base64 string before decoding
	base64Data = strings.TrimSpace(base64Data)
	base64Data = strings.ReplaceAll(base64Data, "\n", "")
	base64Data = strings.ReplaceAll(base64Data, "\r", "")
	base64Data = strings.ReplaceAll(base64Data, " ", "")

	// Add padding if necessary
	if mod := len(base64Data) % 4; mod != 0 {
		base64Data += strings.Repeat("=", 4-mod)
	}

	log.Printf("Cleaned base64 data length: %d", len(base64Data))

	if len(base64Data) == 0 {
		return nil, fmt.Errorf("empty base64 data")
	}

	// Log first few characters for debugging
	if len(base64Data) > 100 {
		log.Printf("First 100 chars of base64: %s", base64Data[:100])
	} else {
		log.Printf("Base64 data: %s", base64Data)
	}

	// Validate base64 characters
	for i, c := range base64Data {
		if !isValidBase64Char(c) {
			log.Printf("Invalid base64 character at position %d: %c (ASCII: %d)", i, c, c)
			return nil, fmt.Errorf("invalid base64 character at position %d: %c", i, c)
		}
	}

	// Decode base64
	fileData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		log.Printf("Base64 decoding failed. Error: %v", err)
		log.Printf("Base64 data length: %d", len(base64Data))
		if len(base64Data) > 100 {
			log.Printf("First 100 chars: %s", base64Data[:100])
		}
		return nil, fmt.Errorf("failed to decode base64 data: %v (input length: %d)", err, len(base64Data))
	}

	log.Printf("Successfully decoded base64, file size: %d bytes", len(fileData))

	// Save to temp file
	tempFilePath, err := utils.SaveBytesToTempFile(fileData, mimeType)
	if err != nil {
		return nil, fmt.Errorf("failed to save temp image: %v", err)
	}
	log.Printf("Temp file saved at: %s", tempFilePath)

	// Continue with Gemini processing
	extractedData, err := s.geminiService.ProcessImage(ctx, base64Data, req.ExtractionFields)
	if err != nil {
		return nil, fmt.Errorf("failed to process with Gemini: %v", err)
	}

	fileURL, err := s.minioRepo.StoreFile(ctx, fileData, mimeType)
	if err != nil {
		return nil, fmt.Errorf("failed to store file: %v", err)
	}

	return &pb.ProcessedFileData{
		FileId:        fmt.Sprintf("file-%d", time.Now().UnixNano()),
		FileUrl:       fileURL,
		ExtractedData: extractedData,
	}, nil
}

// isValidBase64Char checks if a character is valid in base64 encoding
func isValidBase64Char(c rune) bool {
	return (c >= 'A' && c <= 'Z') || // A-Z
		(c >= 'a' && c <= 'z') || // a-z
		(c >= '0' && c <= '9') || // 0-9
		c == '+' || c == '/' || c == '=' // + / =
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
