package services

import (
	"document_processing/pkg/observability"
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
	observability observability.ObservabilityStack
}

func NewGeminiService(apiKey string, minioRepo MinioRepositoryInterface, observability observability.ObservabilityStack) (*GeminiService, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %v", err)
	}

	model := client.GenerativeModel("gemini-2.0-flash")
	return &GeminiService{
		client:    client,
		model:     model,
		minioRepo: minioRepo,
		observability: observability,
	}, nil
}

func (s *GeminiService) ProcessImage(ctx context.Context, base64Image string, extractionFields []string, fileUrl string) (map[string]string, float32, error) {
	traceCtx, span := s.observability.TracerService.StartTracer(ctx, "GeminiService.ProcessImage")
	defer span.End()

	s.observability.TracerService.SetAttributes(span, map[string]string{
		"handler":    "ProcessImage",
		"fileUrl":    fileUrl,
		"fieldCount": fmt.Sprintf("%d", len(extractionFields)),
	})

	s.observability.LoggerService.Info(traceCtx, "Started Gemini image processing", fmt.Sprintf("base64Length=%d", len(base64Image)))

	base64Data, mimeType, err := s.cleanAndValidateBase64(ctx, base64Image)
	if err != nil {
		s.observability.LoggerService.Error(traceCtx, "Base64 validation failed", err.Error())
		s.observability.MetricsService.IncrementCounter(traceCtx, "gemini_process_image.error", 1, map[string]string{"reason": "base64_validation"})
		span.RecordError(err)
		return nil, 0, fmt.Errorf("base64 validation failed: %v", err)
	}

	s.observability.LoggerService.Debug(traceCtx, "Validated base64 image", fmt.Sprintf("mimeType=%s", mimeType))

	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		s.observability.LoggerService.Error(traceCtx, "Base64 decoding failed", err.Error())
		s.observability.MetricsService.IncrementCounter(traceCtx, "gemini_process_image.error", 1, map[string]string{"reason": "decode_failure"})
		span.RecordError(err)
		return nil, 0, fmt.Errorf("base64 decoding failed: %v", err)
	}

	s.observability.LoggerService.Info(traceCtx, "Successfully decoded image", fmt.Sprintf("size=%d bytes", len(imageData)))

	img := genai.ImageData(mimeType, imageData)
	prompt := fmt.Sprintf(`Extract these fields from the document: %s. 
		Return ONLY valid JSON with these exact field names and an overall confidence score.
		If a field is missing, use empty string.
		Example: {"data":{"name":"John Doe","email":"john@example.com"},"confidence":0.95}`,
		strings.Join(extractionFields, ", "))

	var resp *genai.GenerateContentResponse
	maxRetries := 2
	baseDelay := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			err := fmt.Errorf("context canceled while processing with Gemini: %v", ctx.Err())
			s.observability.LoggerService.Warn(traceCtx, "Context canceled", err.Error())
			s.observability.MetricsService.IncrementCounter(traceCtx, "gemini_process_image.error", 1, map[string]string{"reason": "context_cancel"})
			span.RecordError(err)
			return nil, 0, err
		default:
			attemptCtx, cancel := context.WithTimeout(traceCtx, 30*time.Second)
			resp, err = s.model.GenerateContent(attemptCtx, img, genai.Text(prompt))
			cancel()

			if err == nil {
				break
			}

			s.observability.LoggerService.Warn(traceCtx, "Gemini API attempt failed", err.Error())
			span.RecordError(err)

			if i < maxRetries-1 {
				delay := baseDelay * time.Duration(1<<uint(i))
				s.observability.LoggerService.Debug(traceCtx, "Retrying Gemini API", fmt.Sprintf("delay=%v, attempt=%d", delay, i+2))
				select {
				case <-ctx.Done():
					err := fmt.Errorf("context canceled during retry delay: %v", ctx.Err())
					s.observability.LoggerService.Warn(traceCtx, "Retry context canceled", err.Error())
					s.observability.MetricsService.IncrementCounter(traceCtx, "gemini_process_image.error", 1, map[string]string{"reason": "retry_cancel"})
					span.RecordError(err)
					return nil, 0, err
				case <-time.After(delay):
					continue
				}
			}
		}
	}

	if err != nil {
		s.observability.LoggerService.Error(traceCtx, "Gemini API failed after retries", err.Error())
		s.observability.MetricsService.IncrementCounter(traceCtx, "gemini_process_image.error", 1, map[string]string{"reason": "gemini_api_failure"})
		span.RecordError(err)
		if delErr := s.minioRepo.DeleteFile(ctx, fileUrl); delErr != nil {
			s.observability.LoggerService.Warn(traceCtx, "Failed to delete file from MinIO", delErr.Error())
		}
		return nil, 0, fmt.Errorf("gemini API failed after %d retries: %v", maxRetries, err)
	}

	// Parse response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		err := fmt.Errorf("empty response from Gemini")
		s.observability.LoggerService.Error(traceCtx, err.Error(), "")
		s.observability.MetricsService.IncrementCounter(traceCtx, "gemini_process_image.error", 1, map[string]string{"reason": "empty_response"})
		span.RecordError(err)
		return nil, 0, err
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

	s.observability.LoggerService.Debug(traceCtx, "Gemini raw response", responseText)

	type Response struct {
		Data       map[string]string `json:"data"`
		Confidence float32           `json:"confidence"`
	}
	var response Response
	if err := json.Unmarshal([]byte(responseText), &response); err != nil {
		s.observability.LoggerService.Error(traceCtx, "Failed to parse Gemini response", err.Error())
		s.observability.MetricsService.IncrementCounter(traceCtx, "gemini_process_image.error", 1, map[string]string{"reason": "json_parse_failure"})
		span.RecordError(err)
		return nil, 0, fmt.Errorf("failed to parse Gemini response: %v\nResponse: %s", err, responseText)
	}

	for _, field := range extractionFields {
		if _, exists := response.Data[field]; !exists {
			response.Data[field] = ""
		}
	}

	s.observability.LoggerService.Info(traceCtx, "Gemini processing complete", fmt.Sprintf("confidence=%.2f", response.Confidence*100))
	s.observability.MetricsService.IncrementCounter(traceCtx, "gemini_process_image.success", 1, nil)

	return response.Data, response.Confidence * 100, nil
}

func (s *GeminiService) cleanAndValidateBase64(ctx context.Context, base64Str string) (string, string, error) {
	traceCtx, span := s.observability.TracerService.StartTracer(ctx, "GeminiService.cleanAndValidateBase64")
	defer span.End()

	s.observability.LoggerService.Debug(traceCtx, "Starting base64 validation", fmt.Sprintf("inputLength=%d", len(base64Str)))

	if base64Str == "" {
		err := fmt.Errorf("empty input")
		s.observability.LoggerService.Error(traceCtx, "Validation failed", err.Error())
		s.observability.MetricsService.IncrementCounter(traceCtx, "base64_validation.error", 1, map[string]string{"reason": "empty_input"})
		span.RecordError(err)
		return "", "", err
	}

	if strings.HasPrefix(base64Str, "data:") {
		parts := strings.SplitN(base64Str, ",", 2)
		if len(parts) != 2 {
			err := fmt.Errorf("invalid data URL format")
			s.observability.LoggerService.Error(traceCtx, "Validation failed", err.Error())
			s.observability.MetricsService.IncrementCounter(traceCtx, "base64_validation.error", 1, map[string]string{"reason": "invalid_data_url"})
			span.RecordError(err)
			return "", "", err
		}

		mimeParts := strings.Split(parts[0], ";")
		if len(mimeParts) == 0 {
			err := fmt.Errorf("missing MIME type in data URL")
			s.observability.LoggerService.Error(traceCtx, "Validation failed", err.Error())
			s.observability.MetricsService.IncrementCounter(traceCtx, "base64_validation.error", 1, map[string]string{"reason": "missing_mime"})
			span.RecordError(err)
			return "", "", err
		}

		mimeType := strings.TrimPrefix(mimeParts[0], "data:")
		base64Data := strings.TrimSpace(parts[1])
		base64Data = strings.ReplaceAll(base64Data, "\n", "")
		base64Data = strings.ReplaceAll(base64Data, "\r", "")
		base64Data = strings.ReplaceAll(base64Data, " ", "")

		if mod := len(base64Data) % 4; mod != 0 {
			base64Data += strings.Repeat("=", 4-mod)
		}

		s.observability.LoggerService.Debug(traceCtx, "Data URL format validated", fmt.Sprintf("mimeType=%s", mimeType))
		s.observability.MetricsService.IncrementCounter(traceCtx, "base64_validation.success", 1, map[string]string{"format": "data_url"})
		return base64Data, mimeType, nil
	}

	// Raw base64 format
	base64Data := strings.TrimSpace(base64Str)
	base64Data = strings.ReplaceAll(base64Data, "\n", "")
	base64Data = strings.ReplaceAll(base64Data, "\r", "")
	base64Data = strings.ReplaceAll(base64Data, " ", "")

	if mod := len(base64Data) % 4; mod != 0 {
		base64Data += strings.Repeat("=", 4-mod)
	}

	s.observability.LoggerService.Debug(traceCtx, "Raw base64 format validated", "")
	s.observability.MetricsService.IncrementCounter(traceCtx, "base64_validation.success", 1, map[string]string{"format": "raw_base64"})
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
