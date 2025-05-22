package handlers_test

import (
	"context"
	"testing"

	"Document-Processing/internal/handlers"
	"Document-Processing/pkg/observability"
	pb "Document-Processing/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDocumentService is a mock implementation of the DocumentService
type MockDocumentService struct {
	mock.Mock
}

// ProcessBatchFilesV1 mocks the corresponding method in DocumentService
func (m *MockDocumentService) ProcessBatchFilesV1(ctx context.Context, req *pb.BatchFileProcessingRequest) (*pb.BatchProcessingAck, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*pb.BatchProcessingAck), args.Error(1)
}

// GetBatchStatusV1 mocks the corresponding method in DocumentService
func (m *MockDocumentService) GetBatchStatusV1(ctx context.Context, req *pb.BatchStatusRequest) (*pb.BatchFileProcessingResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*pb.BatchFileProcessingResponse), args.Error(1)
}

// TestProcessBatchFilesV1 tests the ProcessBatchFilesV1 handler method
func TestProcessBatchFilesV1(t *testing.T) {
	mockService := new(MockDocumentService)
	var _ handlers.DocumentServiceHandlerInterface = (*MockDocumentService)(nil) // Interface assertion
	obs := *observability.NewObservabilityStack()
	handler := handlers.NewDocumentHandler(mockService, obs)

	req := &pb.BatchFileProcessingRequest{
		Files: []*pb.FileProcessingRequest{
			{
				File: &pb.FileData{
					Base64File: "testbase64",
					FileType:   "image/jpeg",
				},
				ExtractionFields: []string{"field1", "field2"},
			},
		},
		Classifier: "gemini",
		BatchId:    "test-batch-id",
	}

	expectedAck := &pb.BatchProcessingAck{
		BatchId:  "test-batch-id",
		Status:   "ACCEPTED",
		Message:  "Batch processing started",
		FileUrls: []string{"http://example.com/file1"},
	}

	// Set up expectations
	mockService.On("ProcessBatchFilesV1", mock.Anything, req).Return(expectedAck, nil)

	// Call the handler method
	ack, err := handler.ProcessBatchFilesV1(context.Background(), req)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedAck, ack)
	mockService.AssertExpectations(t)
}

// TestGetBatchStatusV1 tests the GetBatchStatusV1 handler method
func TestGetBatchStatusV1(t *testing.T) {
	mockService := new(MockDocumentService)
	obs := *observability.NewObservabilityStack()
	handler := handlers.NewDocumentHandler(mockService, obs)

	req := &pb.BatchStatusRequest{
		BatchId: "test-batch-id",
	}

	expectedResponse := &pb.BatchFileProcessingResponse{
		Success: true,
		Message: "Batch processing completed",
		ProcessedFiles: []*pb.ProcessedFileData{
			{
				FileId:        "file-1",
				FileUrl:       "http://example.com/file1",
				ExtractedData: map[string]string{"field1": "value1"},
				Confidence:    99.9,
			},
		},
		CreatedAt: "2023-10-27T10:00:00Z", // Example timestamp
	}

	// Set up expectations
	mockService.On("GetBatchStatusV1", mock.Anything, req).Return(expectedResponse, nil)

	// Call the handler method
	res, err := handler.GetBatchStatusV1(context.Background(), req)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, res)
	mockService.AssertExpectations(t)
}
