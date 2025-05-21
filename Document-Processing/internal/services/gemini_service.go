package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiModelInterface defines the interface for the Gemini model
type GeminiModelInterface interface {
	GenerateContent(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error)
}

// MinioRepositoryInterface defines the interface for MinIO operations
type MinioRepositoryInterface interface {
	StoreFile(ctx context.Context, fileData []byte, fileType string) (string, error)
	GetFile(ctx context.Context, fileName string) ([]byte, error)
	DeleteFile(ctx context.Context, fileUrl string) error
}

type GeminiService struct {
	client    *genai.Client
	model     GeminiModelInterface
	minioRepo MinioRepositoryInterface
}

func NewGeminiService(apiKey string, minioRepo MinioRepositoryInterface) (*GeminiService, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %v", err)
	}

	model := client.GenerativeModel("gemini-1.5-flash")
	return &GeminiService{
		client:    client,
		model:     model,
		minioRepo: minioRepo,
	}, nil
}

func (s *GeminiService) ProcessImage(ctx context.Context, base64Image string, extractionFields []string, fileUrl string) (map[string]string, float32, error) {
	// Step 1: Clean and validate base64 string
	base64Data, mimeType, err := cleanAndValidateBase64(base64Image)
	if err != nil {
		return nil, 0, fmt.Errorf("base64 validation failed: %v", err)
	}

	// Log base64 data details for debugging
	fmt.Printf("Processing image with base64 length: %d, MIME type: %s\n", len(base64Data), mimeType)

	// Step 2: Decode base64 image
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		fmt.Printf("Base64 decoding failed: %v\n", err)
		fmt.Printf("Base64 data length: %d\n", len(base64Data))
		fmt.Printf("First 100 chars: %s\n", base64Data[:min(100, len(base64Data))])
		return nil, 0, fmt.Errorf("base64 decoding failed: %v (input length: %d)", err, len(base64Data))
	}

	fmt.Printf("Successfully decoded image, size: %d bytes\n", len(imageData))

	// Step 3: Create image data for Gemini
	img := genai.ImageData(mimeType, imageData)

	// Step 4: Create extraction prompt
	prompt := fmt.Sprintf(`Extract these fields from the document: %s. 
		Return ONLY valid JSON with these exact field names and an overall confidence score.
		If a field is missing, use empty string.
		Example: {"data":{"name":"John Doe","email":"john@example.com"},"confidence":0.95}`,
		strings.Join(extractionFields, ", "))

	// Step 5: Call Gemini with retry logic and exponential backoff
	var resp *genai.GenerateContentResponse
	maxRetries := 2
	baseDelay := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return nil, 0, fmt.Errorf("context canceled while processing with Gemini: %v", ctx.Err())
		default:
			// Create a new context with timeout for this attempt
			attemptCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			resp, err = s.model.GenerateContent(attemptCtx, img, genai.Text(prompt))
			cancel()

			if err == nil {
				break
			}

			fmt.Printf("Gemini API attempt %d failed: %v\n", i+1, err)

			if i < maxRetries-1 {
				// Calculate delay with exponential backoff
				delay := baseDelay * time.Duration(1<<uint(i))
				fmt.Printf("Retrying in %v...\n", delay)
				select {
				case <-ctx.Done():
					return nil, 0, fmt.Errorf("context canceled while waiting to retry: %v", ctx.Err())
				case <-time.After(delay):
					continue
				}
			}
		}
	}

	if err != nil {
		if err := s.minioRepo.DeleteFile(ctx, fileUrl); err != nil {
			fmt.Printf("Failed to delete file from MinIO: %v", err)
		}
		return nil, 0, fmt.Errorf("gemini API failed after %d retries: %v", maxRetries, err)
	}

	// Step 6: Parse response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, 0, fmt.Errorf("empty response from Gemini")
	}

	var responseText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			responseText += string(textPart)
		}
	}

	// Clean JSON response
	responseText = strings.TrimSpace(responseText)
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	fmt.Printf("Gemini response: %s\n", responseText)

	// Parse JSON
	type Response struct {
		Data       map[string]string `json:"data"`
		Confidence float32           `json:"confidence"`
	}
	var response Response
	if err := json.Unmarshal([]byte(responseText), &response); err != nil {
		return nil, 0, fmt.Errorf("failed to parse Gemini response: %v\nResponse: %s", err, responseText)
	}

	// Ensure all requested fields are present
	for _, field := range extractionFields {
		if _, exists := response.Data[field]; !exists {
			response.Data[field] = ""
		}
	}

	return response.Data, response.Confidence * 100, nil // Convert to percentage
}

func cleanAndValidateBase64(base64Str string) (string, string, error) {
	// Handle empty input
	if base64Str == "" {
		return "", "", fmt.Errorf("empty input")
	}

	// Handle data URL format
	if strings.HasPrefix(base64Str, "data:") {
		parts := strings.SplitN(base64Str, ",", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid data URL format")
		}

		// Extract MIME type
		mimeParts := strings.Split(parts[0], ";")
		if len(mimeParts) == 0 {
			return "", "", fmt.Errorf("missing MIME type in data URL")
		}
		mimeType := strings.TrimPrefix(mimeParts[0], "data:")

		// Clean base64 data
		base64Data := strings.TrimSpace(parts[1])
		base64Data = strings.ReplaceAll(base64Data, "\n", "")
		base64Data = strings.ReplaceAll(base64Data, "\r", "")
		base64Data = strings.ReplaceAll(base64Data, " ", "")

		// Add padding if necessary
		if mod := len(base64Data) % 4; mod != 0 {
			base64Data += strings.Repeat("=", 4-mod)
		}

		return base64Data, mimeType, nil
	}

	// For raw base64 without data URL prefix
	base64Data := strings.TrimSpace(base64Str)
	base64Data = strings.ReplaceAll(base64Data, "\n", "")
	base64Data = strings.ReplaceAll(base64Data, "\r", "")
	base64Data = strings.ReplaceAll(base64Data, " ", "")

	// Add padding if necessary
	if mod := len(base64Data) % 4; mod != 0 {
		base64Data += strings.Repeat("=", 4-mod)
	}

	// Default MIME type for raw base64
	return base64Data, "image/jpeg", nil
}

// Helper function to get minimum of two integers
// func min(a, b int) int {
// 	if a < b {
// 		return a
// 	}
// 	return b
// }

func (s *GeminiService) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}
