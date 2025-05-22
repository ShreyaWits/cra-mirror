package services

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/generative-ai-go/genai"
)

type mockMinioRepository struct {
	deleteFileErr error
	storeFileErr  error
	getFileErr    error
	storeFileURL  string
	getFileData   []byte
}

func (m *mockMinioRepository) DeleteFile(ctx context.Context, fileUrl string) error {
	return m.deleteFileErr
}

func (m *mockMinioRepository) StoreFile(ctx context.Context, fileData []byte, fileType string) (string, error) {
	if m.storeFileErr != nil {
		return "", m.storeFileErr
	}
	return m.storeFileURL, nil
}

func (m *mockMinioRepository) GetFile(ctx context.Context, fileName string) ([]byte, error) {
	if m.getFileErr != nil {
		return nil, m.getFileErr
	}
	return m.getFileData, nil
}

type mockGeminiModel struct {
	generateContentFunc func(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error)
}

func (m *mockGeminiModel) GenerateContent(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	return m.generateContentFunc(ctx, parts...)
}

func TestGeminiService_ProcessImage(t *testing.T) {
	tests := []struct {
		name             string
		base64Image      string
		extractionFields []string
		fileUrl          string
		wantErr          bool
		wantFields       map[string]string
		wantConfidence   float32
		mockResponse     map[string]interface{}
		mockStatusCode   int
	}{
		{
			name:             "valid base64 image",
			base64Image:      "data:image/jpeg;base64,/9j/4AAQSkZJRg==",
			extractionFields: []string{"name", "email"},
			fileUrl:          "test.jpg",
			wantErr:          false,
			wantFields: map[string]string{
				"name":  "John Doe",
				"email": "john@example.com",
			},
			wantConfidence: 95.0,
			mockResponse: map[string]interface{}{
				"candidates": []map[string]interface{}{
					{
						"content": map[string]interface{}{
							"parts": []map[string]interface{}{
								{
									"text": `{"data":{"name":"John Doe","email":"john@example.com"},"confidence":0.95}`,
								},
							},
						},
					},
				},
			},
			mockStatusCode: http.StatusOK,
		},
		{
			name:             "invalid base64 image",
			base64Image:      "invalid-base64",
			extractionFields: []string{"name"},
			fileUrl:          "test.jpg",
			wantErr:          true,
		},
		{
			name:             "empty base64 image",
			base64Image:      "",
			extractionFields: []string{"name"},
			fileUrl:          "test.jpg",
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMinio := &mockMinioRepository{}
			service, err := NewGeminiService("test-api-key", mockMinio)
			if err != nil {
				t.Fatalf("Failed to create GeminiService: %v", err)
			}

			// Skip API call for invalid/empty base64 tests
			if tt.wantErr && (tt.base64Image == "" || tt.base64Image == "invalid-base64") {
				_, _, err := service.ProcessImage(context.Background(), tt.base64Image, tt.extractionFields, tt.fileUrl)
				if (err != nil) != tt.wantErr {
					t.Errorf("GeminiService.ProcessImage() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				return
			}

			// Set up mock model
			service.model = &mockGeminiModel{
				generateContentFunc: func(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
					return &genai.GenerateContentResponse{
						Candidates: []*genai.Candidate{
							{
								Content: &genai.Content{
									Parts: []genai.Part{
										genai.Text(`{"data":{"name":"John Doe","email":"john@example.com"},"confidence":0.95}`),
									},
								},
							},
						},
					}, nil
				},
			}

			got, confidence, err := service.ProcessImage(context.Background(), tt.base64Image, tt.extractionFields, tt.fileUrl)

			if (err != nil) != tt.wantErr {
				t.Errorf("GeminiService.ProcessImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if !compareMaps(got, tt.wantFields) {
					t.Errorf("GeminiService.ProcessImage() = %v, want %v", got, tt.wantFields)
				}
				if confidence != tt.wantConfidence {
					t.Errorf("GeminiService.ProcessImage() confidence = %v, want %v", confidence, tt.wantConfidence)
				}
			}
		})
	}
}

func TestCleanAndValidateBase64(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData string
		wantMime string
		wantErr  bool
	}{
		{
			name:     "valid data URL",
			input:    "data:image/jpeg;base64,/9j/4AAQSkZJRg==",
			wantData: "/9j/4AAQSkZJRg==",
			wantMime: "image/jpeg",
			wantErr:  false,
		},
		{
			name:     "valid raw base64",
			input:    "/9j/4AAQSkZJRg==",
			wantData: "/9j/4AAQSkZJRg==",
			wantMime: "image/jpeg",
			wantErr:  false,
		},
		{
			name:    "invalid data URL format",
			input:   "data:image/jpeg",
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotData, gotMime, err := cleanAndValidateBase64(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("cleanAndValidateBase64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if gotData != tt.wantData {
					t.Errorf("cleanAndValidateBase64() data = %v, want %v", gotData, tt.wantData)
				}
				if gotMime != tt.wantMime {
					t.Errorf("cleanAndValidateBase64() mime = %v, want %v", gotMime, tt.wantMime)
				}
			}
		})
	}
}
