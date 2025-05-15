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
	"Document-Processing/internal/repository"
	pb "Document-Processing/proto"
)

// BatchProcessingState tracks the state of a batch processing job
type BatchProcessingState struct {
	mu             sync.RWMutex
	Status         string
	Message        string
	Error          error
	ProcessedFiles []*pb.ProcessedFileData
}

type DocumentService struct {
	pb.UnimplementedDocumentProcessingServiceV1Server
	geminiService *GeminiService
	llamaService  *LlamaService
	minioRepo     *repository.MinioRepository
	yugabyteRepo  repository.DocumentDataRepository
	batchStates   map[string]*BatchProcessingState
	batchStatesMu sync.RWMutex
}

func NewDocumentService(geminiService *GeminiService, llamaService *LlamaService, minioRepo *repository.MinioRepository, yugabyteRepo repository.DocumentDataRepository) *DocumentService {
	return &DocumentService{
		geminiService: geminiService,
		llamaService:  llamaService,
		minioRepo:     minioRepo,
		yugabyteRepo:  yugabyteRepo,
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

	// Store all files in Minio and collect their URLs
	var fileURLs []string
	for _, fileReq := range req.Files {
		if fileReq.File == nil {
			continue
		}

		fileData, err := base64.StdEncoding.DecodeString(fileReq.File.Base64File)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 data: %v", err)
		}

		fileURL, err := s.minioRepo.StoreFile(ctx, fileData, fileReq.File.FileType)
		if err != nil {
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

	return &pb.BatchProcessingAck{
		BatchId:  batchID,
		Status:   "ACCEPTED",
		Message:  "Batch processing started",
		FileUrls: fileURLs,
	}, nil
}

// GetBatchStatusV1 returns the current status of a batch processing job
func (s *DocumentService) GetBatchStatusV1(ctx context.Context, req *pb.BatchStatusRequest) (*pb.BatchFileProcessingResponse, error) {
	data, err := s.yugabyteRepo.GetDocumentDataByID(req.BatchId)
	if err != nil {
		log.Printf("Error retrieving batch data for ID %s: %v", req.BatchId, err)
		return &pb.BatchFileProcessingResponse{
			Success: false,
			Message: "Failed to retrieve batch data.",
		}, nil
	}

	if data == nil {
		log.Printf("No batch data found or still processing for ID %s", req.BatchId)
		return &pb.BatchFileProcessingResponse{
			Success: false,
			Message: "Document is currently being processed.",
		}, nil
	}

	// decoding JSON

	var decodeExtractedData []map[string]interface{}

	decodeErr := json.Unmarshal([]byte(data.Data), &decodeExtractedData)
	if decodeErr != nil {
		log.Fatal("Failed to decode JSON:", err)
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

		// Step 3: Append the converted proto message to the list
		processedFiles = append(processedFiles, processedFile)
	}

	return &pb.BatchFileProcessingResponse{
		Success:        data.Status == "COMPLETED",
		Message:        data.Message,
		ProcessedFiles: processedFiles,
		CreatedAt:      data.CreatedAt.Format(time.RFC3339),
	}, nil
}

// processBatchInBackground handles the actual file processing
func (s *DocumentService) processBatchInBackground(ctx context.Context, batchID string, req *pb.BatchFileProcessingRequest, fileURLs []string) {
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

			processedFile, err := s.processFile(fileCtx, fileReq, req.Classifier, fileURLs[index])
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

	var extracted_data []map[string]interface{}
	for _, f := range processedFiles {
		item := map[string]interface{}{
			"file_id":        f.FileId,
			"file_url":       f.FileUrl,
			"extracted_data": f.ExtractedData,
			"confidence":     f.Confidence,
		}
		extracted_data = append(extracted_data, item)
	}

	jsonBytes, jsonBytesErr := json.Marshal(extracted_data) // Use json.MarshalIndent for pretty format
	if jsonBytesErr != nil {
		log.Fatalf("Failed to encode JSON: %v", jsonBytesErr)
	}

	extractedData := &models.DocumentData{
		Batch_id:  batchID,
		Data:      string(jsonBytes),
		Status:    batchState.Status,
		Message:   batchState.Message,
		CreatedAt: time.Now(),
	}

	_, err := s.yugabyteRepo.CreateDocumentData(extractedData)

	if err != nil {
		log.Printf("Error saving processed data to DB: %v", err)
		batchState.Error = fmt.Errorf("failed to save processed data to DB: %v", err)
	} else {
		log.Printf("Processed data saved to DB successfully")
	}

	batchState.mu.Unlock()
}

// GeminiService returns the Gemini service instance
func (s *DocumentService) GeminiService() *GeminiService {
	return s.geminiService
}

func (s *DocumentService) processFile(ctx context.Context, req *pb.FileProcessingRequest, classifier string, fileUrl string) (*pb.ProcessedFileData, error) {
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
		// Delete file from MinIO if base64 data is empty
		if err := s.minioRepo.DeleteFile(ctx, fileUrl); err != nil {
			log.Printf("Failed to delete file from MinIO: %v", err)
		}
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
			// Delete file from MinIO if base64 data is invalid
			if err := s.minioRepo.DeleteFile(ctx, fileUrl); err != nil {
				log.Printf("Failed to delete file from MinIO: %v", err)
			}
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
		// Delete file from MinIO if base64 decoding fails
		if err := s.minioRepo.DeleteFile(ctx, fileUrl); err != nil {
			log.Printf("Failed to delete file from MinIO: %v", err)
		}
		return nil, fmt.Errorf("failed to decode base64 data: %v (input length: %d)", err, len(base64Data))
	}

	log.Printf("Successfully decoded base64, file size: %d bytes", len(fileData))

	var extractedData map[string]string
	var confidence float32

	// Llama Doc Processing AI
	if classifier == string(enums.ClassifierLlama) {
		extractedData, err = s.llamaService.ProcessImage(ctx, base64Data, req.ExtractionFields)
		if err != nil {
			// Delete file from MinIO if Llama processing fails
			if err := s.minioRepo.DeleteFile(ctx, fileUrl); err != nil {
				log.Printf("Failed to delete file from MinIO: %v", err)
			}
			return nil, fmt.Errorf("failed to process with LLaMA: %v", err)
		}
		// LLaMA doesn't provide confidence scores, so we'll use a default value
		confidence = 100.0
	} else {
		// Gemini Doc Processing AI
		extractedData, confidence, err = s.geminiService.ProcessImage(ctx, base64Data, req.ExtractionFields, fileUrl)
		if err != nil {
			// Delete file from MinIO if Gemini processing fails
			// if err := s.minioRepo.DeleteFile(ctx, fileUrl); err != nil {
			// 	log.Printf("Failed to delete file from MinIO: %v", err)
			// }
			return nil, fmt.Errorf("failed to process with Gemini: %v", err)
		}
	}

	return &pb.ProcessedFileData{
		FileId:        fmt.Sprintf("file-%d", time.Now().UnixNano()),
		FileUrl:       fileUrl,
		ExtractedData: extractedData,
		Confidence:    confidence,
	}, nil
}

// isValidBase64Char checks if a character is valid in base64 encoding
func isValidBase64Char(c rune) bool {
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
