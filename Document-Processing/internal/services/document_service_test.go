package services_test

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"Document-Processing/internal/models"
	"Document-Processing/internal/services"

	pb "Document-Processing/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock GeminiServiceInterface
type MockGeminiService struct {
	mock.Mock
}

func (m *MockGeminiService) ProcessImage(ctx context.Context, base64Image string, extractionFields []string, fileUrl string) (map[string]string, float32, error) {
	args := m.Called(ctx, base64Image, extractionFields, fileUrl)
	return args.Get(0).(map[string]string), args.Get(1).(float32), args.Error(2)
}

func (m *MockGeminiService) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Mock LlamaServiceInterface
type MockLlamaService struct {
	mock.Mock
}

func (m *MockLlamaService) ProcessImage(ctx context.Context, base64Image string, extractionFields []string) (map[string]string, error) {
	args := m.Called(ctx, base64Image, extractionFields)
	return args.Get(0).(map[string]string), args.Error(1)
}

// Mock MinioRepositoryInterface
type MockMinioRepo struct {
	mock.Mock
}

func (m *MockMinioRepo) StoreFile(ctx context.Context, data []byte, fileType string) (string, error) {
	args := m.Called(ctx, data, fileType)
	return args.String(0), args.Error(1)
}

func (m *MockMinioRepo) DeleteFile(ctx context.Context, fileUrl string) error {
	args := m.Called(ctx, fileUrl)
	return args.Error(0)
}

// Mock DocumentDataRepository
type MockDocDataRepo struct {
	mock.Mock
}

func (m *MockDocDataRepo) GetDocumentDataByID(id string) (*models.DocumentData, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DocumentData), args.Error(1)
}

func (m *MockDocDataRepo) CreateDocumentData(data *models.DocumentData) (*models.DocumentData, error) {
	args := m.Called(data)
	// Get(0) should be *models.DocumentData, so cast it accordingly
	result, _ := args.Get(0).(*models.DocumentData)
	return result, args.Error(1)
}

func TestProcessBatchFilesV1_Success(t *testing.T) {
	ctx := context.Background()

	mockGemini := new(MockGeminiService)
	mockLlama := new(MockLlamaService)
	mockMinio := new(MockMinioRepo)
	mockDocRepo := new(MockDocDataRepo)

	mockGemini.On("ProcessImage", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("string")).
		Return(map[string]string{"key": "value"}, float32(0.99), nil)

	sampleData := base64.StdEncoding.EncodeToString([]byte("dummy file content"))

	req := &pb.BatchFileProcessingRequest{
		Files: []*pb.FileProcessingRequest{
			{
				File: &pb.FileData{
					Base64File: sampleData,
					FileType:   "image/png",
				},
			},
		},
	}

	mockMinio.On("StoreFile", mock.Anything, mock.AnythingOfType("[]uint8"), "image/png").
		Return("https://minio/fakefile.png", nil)

	mockDocRepo.On("CreateDocumentData", mock.Anything).
		Return(&models.DocumentData{}, nil)

	svc := services.NewDocumentService(mockGemini, mockLlama, mockMinio, mockDocRepo)

	ack, err := svc.ProcessBatchFilesV1(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, ack)
	assert.Equal(t, "ACCEPTED", ack.Status)
	assert.Len(t, ack.FileUrls, 1)
	assert.Contains(t, ack.FileUrls[0], "https://minio")

	mockMinio.AssertExpectations(t)
}

func TestProcessBatchFilesV1_NoFiles(t *testing.T) {
	ctx := context.Background()

	mockGemini := new(MockGeminiService)
	mockLlama := new(MockLlamaService)
	mockMinio := new(MockMinioRepo)
	mockDocRepo := new(MockDocDataRepo)

	req := &pb.BatchFileProcessingRequest{
		Files: []*pb.FileProcessingRequest{},
	}

	svc := services.NewDocumentService(mockGemini, mockLlama, mockMinio, mockDocRepo)

	ack, err := svc.ProcessBatchFilesV1(ctx, req)

	assert.Nil(t, ack)
	assert.Error(t, err)
	assert.Equal(t, "no files provided in request", err.Error())
}

func TestProcessBatchFilesV1_MinIOStoreError(t *testing.T) {
	ctx := context.Background()

	mockGemini := new(MockGeminiService)
	mockLlama := new(MockLlamaService)
	mockMinio := new(MockMinioRepo)
	mockDocRepo := new(MockDocDataRepo)

	sampleData := base64.StdEncoding.EncodeToString([]byte("dummy file content"))

	req := &pb.BatchFileProcessingRequest{
		Files: []*pb.FileProcessingRequest{
			{
				File: &pb.FileData{
					Base64File: sampleData,
					FileType:   "image/png",
				},
			},
		},
	}

	mockMinio.On("StoreFile", mock.Anything, mock.AnythingOfType("[]uint8"), "image/png").
		Return("", errors.New("minio storage failed"))

	svc := services.NewDocumentService(mockGemini, mockLlama, mockMinio, mockDocRepo)

	ack, err := svc.ProcessBatchFilesV1(ctx, req)

	assert.Nil(t, ack)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to store file in Minio")

	mockMinio.AssertExpectations(t)
}

func TestProcessBatchFilesV1_Base64DecodeError(t *testing.T) {
	ctx := context.Background()

	mockGemini := new(MockGeminiService)
	mockLlama := new(MockLlamaService)
	mockMinio := new(MockMinioRepo)
	mockDocRepo := new(MockDocDataRepo)

	// Invalid base64 data
	sampleData := "invalid-base64-data"

	req := &pb.BatchFileProcessingRequest{
		Files: []*pb.FileProcessingRequest{
			{
				File: &pb.FileData{
					Base64File: sampleData,
					FileType:   "image/png",
				},
			},
		},
	}

	svc := services.NewDocumentService(mockGemini, mockLlama, mockMinio, mockDocRepo)

	ack, err := svc.ProcessBatchFilesV1(ctx, req)

	assert.Nil(t, ack)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode base64 data")

	mockMinio.AssertExpectations(t)
	mockDocRepo.AssertExpectations(t)
	mockGemini.AssertExpectations(t)
	mockLlama.AssertExpectations(t)
}

func TestProcessBatchFilesV1_GeneratedBatchID(t *testing.T) {
	ctx := context.Background()

	mockGemini := new(MockGeminiService)
	mockLlama := new(MockLlamaService)
	mockMinio := new(MockMinioRepo)
	mockDocRepo := new(MockDocDataRepo)

	mockGemini.On("ProcessImage", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("string")).
		Return(map[string]string{"key": "value"}, float32(0.99), nil)

	sampleData := base64.StdEncoding.EncodeToString([]byte("dummy file content"))

	req := &pb.BatchFileProcessingRequest{
		Files: []*pb.FileProcessingRequest{
			{
				File: &pb.FileData{
					Base64File: sampleData,
					FileType:   "image/png",
				},
			},
		},
		// BatchId is intentionally not provided
	}

	mockMinio.On("StoreFile", mock.Anything, mock.AnythingOfType("[]uint8"), "image/png").
		Return("https://minio/fakefile.png", nil)

	mockDocRepo.On("CreateDocumentData", mock.Anything).
		Return(&models.DocumentData{}, nil)

	svc := services.NewDocumentService(mockGemini, mockLlama, mockMinio, mockDocRepo)

	ack, err := svc.ProcessBatchFilesV1(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, ack)
	assert.Equal(t, "ACCEPTED", ack.Status)
	assert.Len(t, ack.FileUrls, 1)
	assert.Contains(t, ack.FileUrls[0], "https://minio")
	assert.NotEmpty(t, ack.BatchId)                          // Assert that a batch ID was generated
	assert.True(t, strings.HasPrefix(ack.BatchId, "batch-")) // Assert the format of the generated ID

	mockMinio.AssertExpectations(t)
	// Note: Background processing is not waited for in this test,
	// as we are only testing the immediate ACK response and batch ID generation.
}

func TestProcessBatchFilesV1_GeminiProcessError(t *testing.T) {
	ctx := context.Background()

	mockGemini := new(MockGeminiService)
	mockLlama := new(MockLlamaService)
	mockMinio := new(MockMinioRepo)
	mockDocRepo := new(MockDocDataRepo)

	sampleData := base64.StdEncoding.EncodeToString([]byte("dummy file content"))

	req := &pb.BatchFileProcessingRequest{
		Files: []*pb.FileProcessingRequest{
			{
				File: &pb.FileData{
					Base64File: sampleData,
					FileType:   "image/jpeg",
				},
				ExtractionFields: []string{"field1"},
			},
		},
		Classifier: "gemini",
	}

	mockMinio.On("StoreFile", mock.Anything, mock.AnythingOfType("[]uint8"), "image/jpeg").
		Return("https://minio/fakefile.jpeg", nil)

	// ✅ FIX: Return empty map instead of nil to avoid type assertion panic
	mockGemini.On("ProcessImage", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("[]string"), mock.AnythingOfType("string")).
		Return(map[string]string{}, float32(0), errors.New("gemini processing failed"))

	mockDocRepo.On("CreateDocumentData", mock.AnythingOfType("*models.DocumentData")).
		Return(&models.DocumentData{}, nil)

	svc := services.NewDocumentService(mockGemini, mockLlama, mockMinio, mockDocRepo)

	ack, err := svc.ProcessBatchFilesV1(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, ack)
	assert.Equal(t, "ACCEPTED", ack.Status)

	// Wait for background processing to complete (adjust sleep if necessary)
	time.Sleep(100 * time.Millisecond)

	// Get the batch state and verify expected error handling
	batchState, ok := svc.GetBatchStateForTest(ack.BatchId)

	assert.True(t, ok)
	assert.Equal(t, "FAILED", batchState.Status)
	assert.Contains(t, batchState.Message, "gemini processing failed")

	mockMinio.AssertExpectations(t)
	mockGemini.AssertExpectations(t)
	mockDocRepo.AssertExpectations(t)
	mockLlama.AssertExpectations(t) // Ensure Llama was not called
}

func TestProcessBatchFilesV1_LlamaProcessError(t *testing.T) {
	ctx := context.Background()

	mockGemini := new(MockGeminiService)
	mockLlama := new(MockLlamaService)
	mockMinio := new(MockMinioRepo)
	mockDocRepo := new(MockDocDataRepo)

	sampleData := base64.StdEncoding.EncodeToString([]byte("dummy file content"))

	req := &pb.BatchFileProcessingRequest{
		Files: []*pb.FileProcessingRequest{
			{
				File: &pb.FileData{
					Base64File: sampleData,
					FileType:   "image/png",
				},
				ExtractionFields: []string{"field1"},
			},
		},
		Classifier: "llama", // Use Llama classifier
	}

	// Mock Minio StoreFile success
	mockMinio.On("StoreFile", mock.Anything, mock.AnythingOfType("[]uint8"), "image/png").
		Return("https://minio/fakefile.png", nil)

	// Mock Minio DeleteFile success (needed to avoid panic)
	mockMinio.On("DeleteFile", mock.Anything, mock.AnythingOfType("string")).
		Return(nil)

	// Mock LlamaService ProcessImage to return an error and empty map (to avoid interface{} nil panic)
	mockLlama.On("ProcessImage", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("[]string")).
		Return(make(map[string]string), errors.New("llama processing failed"))

	// Expect DocRepo CreateDocumentData to be called with FAILED status
	mockDocRepo.On("CreateDocumentData", mock.AnythingOfType("*models.DocumentData")).
		Return(&models.DocumentData{}, nil)

	svc := services.NewDocumentService(mockGemini, mockLlama, mockMinio, mockDocRepo)

	ack, err := svc.ProcessBatchFilesV1(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, ack)
	assert.Equal(t, "ACCEPTED", ack.Status)

	// Wait for background processing to complete (adjust duration if needed)
	time.Sleep(100 * time.Millisecond)

	batchState, ok := svc.GetBatchStateForTest(ack.BatchId)

	assert.True(t, ok)
	assert.Equal(t, "FAILED", batchState.Status)
	assert.Contains(t, batchState.Message, "llama processing failed")

	mockMinio.AssertExpectations(t)
	mockLlama.AssertExpectations(t)
	mockDocRepo.AssertExpectations(t)
	mockGemini.AssertExpectations(t)
}

func TestGetBatchStatusV1_Success(t *testing.T) {
	ctx := context.Background()

	mockGemini := new(MockGeminiService)
	mockLlama := new(MockLlamaService)
	mockMinio := new(MockMinioRepo)
	mockDocRepo := new(MockDocDataRepo)

	// Prepare dummy document data with JSON string
	docData := &models.DocumentData{
		Batch_id: "batch123",
		Data: `[{
			"file_id":"file1",
			"file_url":"https://minio/file1.png",
			"confidence":0.95,
			"extracted_data":{"key1":"value1"}
		}]`,
		Status:    "COMPLETED",
		Message:   "Done",
		CreatedAt: time.Now(),
	}

	mockDocRepo.On("GetDocumentDataByID", "batch123").Return(docData, nil)

	svc := services.NewDocumentService(mockGemini, mockLlama, mockMinio, mockDocRepo)

	resp, err := svc.GetBatchStatusV1(ctx, &pb.BatchStatusRequest{BatchId: "batch123"})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "Done", resp.Message)
	assert.Len(t, resp.ProcessedFiles, 1)
	assert.Equal(t, "file1", resp.ProcessedFiles[0].FileId)
	assert.Equal(t, "value1", resp.ProcessedFiles[0].ExtractedData["key1"])

	mockDocRepo.AssertExpectations(t)
}

func TestGetBatchStatusV1_NotFound(t *testing.T) {
	ctx := context.Background()

	mockGemini := new(MockGeminiService)
	mockLlama := new(MockLlamaService)
	mockMinio := new(MockMinioRepo)
	mockDocRepo := new(MockDocDataRepo)

	mockDocRepo.On("GetDocumentDataByID", "batchNotFound").Return(nil, nil)
	svc := services.NewDocumentService(mockGemini, mockLlama, mockMinio, mockDocRepo)

	resp, err := svc.GetBatchStatusV1(ctx, &pb.BatchStatusRequest{BatchId: "batchNotFound"})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.Equal(t, "Document is currently being processed.", resp.Message)

	mockDocRepo.AssertExpectations(t)
}

func TestGetBatchStatusV1_DBError(t *testing.T) {
	ctx := context.Background()

	mockGemini := new(MockGeminiService)
	mockLlama := new(MockLlamaService)
	mockMinio := new(MockMinioRepo)
	mockDocRepo := new(MockDocDataRepo)

	mockDocRepo.On("GetDocumentDataByID", "batchError").Return(nil, errors.New("db failure"))

	svc := services.NewDocumentService(mockGemini, mockLlama, mockMinio, mockDocRepo)

	resp, err := svc.GetBatchStatusV1(ctx, &pb.BatchStatusRequest{BatchId: "batchError"})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.Equal(t, "Failed to retrieve batch data.", resp.Message) // Corrected assertion
	mockDocRepo.AssertExpectations(t)                               // Added assertion
}

func TestGetBatchStatusV1_JSONDecodeError(t *testing.T) {
	ctx := context.Background()

	mockGemini := new(MockGeminiService)
	mockLlama := new(MockLlamaService)
	mockMinio := new(MockMinioRepo)
	mockDocRepo := new(MockDocDataRepo)

	// Prepare dummy document data with invalid JSON string
	docData := &models.DocumentData{
		Batch_id:  "batch123",
		Data:      `invalid json`, // Invalid JSON
		Status:    "COMPLETED",
		Message:   "Done",
		CreatedAt: time.Now(),
	}

	mockDocRepo.On("GetDocumentDataByID", "batch123").Return(docData, nil)

	svc := services.NewDocumentService(mockGemini, mockLlama, mockMinio, mockDocRepo)

	resp, err := svc.GetBatchStatusV1(ctx, &pb.BatchStatusRequest{BatchId: "batch123"})

	assert.NoError(t, err) // Expect no gRPC-level error, but an error indicated in the response
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Message, "Internal error: Failed to decode processed data:")

	mockDocRepo.AssertExpectations(t)
}

func TestProcessBatchFilesV1_ProcessFile_EmptyBase64(t *testing.T) {
	ctx := context.Background()

	mockGemini := new(MockGeminiService)
	mockLlama := new(MockLlamaService)
	mockMinio := new(MockMinioRepo)
	mockDocRepo := new(MockDocDataRepo)

	// Prepare request with empty base64 data
	req := &pb.BatchFileProcessingRequest{
		Files: []*pb.FileProcessingRequest{
			{
				File: &pb.FileData{
					Base64File: "", // Empty base64 data
					FileType:   "image/png",
				},
			},
		},
	}

	// Expect StoreFile to be called (it will return a fake URL)
	mockMinio.On("StoreFile", mock.Anything, mock.AnythingOfType("[]uint8"), "image/png").
		Return("https://minio/fakefile.png", nil)

	// Expect DeleteFile to be called when processFile encounters empty base64
	mockMinio.On("DeleteFile", mock.Anything, "https://minio/fakefile.png").
		Return(nil).Once()

	// Expect DocRepo CreateDocumentData to be called with FAILED status
	mockDocRepo.On("CreateDocumentData", mock.AnythingOfType("*models.DocumentData")).
		Return(&models.DocumentData{}, nil)

	svc := services.NewDocumentService(mockGemini, mockLlama, mockMinio, mockDocRepo)

	ack, err := svc.ProcessBatchFilesV1(ctx, req)

	assert.NoError(t, err) // ProcessBatchFilesV1 should return ACK even if background processing fails
	assert.NotNil(t, ack)
	assert.Equal(t, "ACCEPTED", ack.Status)

	// Wait for background processing to complete (with a timeout)
	time.Sleep(100 * time.Millisecond) // Adjust sleep duration if needed

	// Check the batch status using the exported test method
	batchState, ok := svc.GetBatchStateForTest(ack.BatchId)

	assert.True(t, ok)
	assert.Equal(t, "FAILED", batchState.Status)
	assert.Contains(t, batchState.Message, "empty base64 data")

	mockMinio.AssertExpectations(t)
	mockGemini.AssertExpectations(t) // Assert that Gemini was not called
	mockLlama.AssertExpectations(t)  // Assert that Llama was not called
	mockDocRepo.AssertExpectations(t)
}
