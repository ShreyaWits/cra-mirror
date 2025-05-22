package services

import (
	"Document-Processing/internal/utils"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type LlamaService struct {
	modelName string
	apiURL    string
	client    *http.Client
}

// modelName: "llama3.2-vision",
// apiURL:    "http://localhost:11434/api/chat",

func NewLlamaService(modelName string, apiURL string) *LlamaService {
	return &LlamaService{
		modelName: modelName,
		apiURL:    apiURL,
		client:    &http.Client{Timeout: 2 * time.Minute},
	}
}

// NewLlamaServiceWithClient creates a new LlamaService with a custom HTTP client
func NewLlamaServiceWithClient(modelName string, apiURL string, client *http.Client) *LlamaService {
	return &LlamaService{
		modelName: modelName,
		apiURL:    apiURL,
		client:    client,
	}
}

func (s *LlamaService) ProcessImage(ctx context.Context, base64Image string, extractionFields []string) (map[string]string, error) {
	fmt.Printf("Llama Process Start")

	// Validate base64 data
	if base64Image == "" {
		return nil, fmt.Errorf("base64 validation failed: empty input")
	}

	// Handle data URL format
	var base64Data string
	if strings.HasPrefix(base64Image, "data:") {
		parts := strings.SplitN(base64Image, ",", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("base64 validation failed: invalid data URL format")
		}
		base64Data = parts[1]
	} else {
		base64Data = base64Image
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

	// Validate base64 characters
	for i, c := range base64Data {
		if !IsValidBase64Char(c) {
			return nil, fmt.Errorf("base64 validation failed: invalid character at position %d: %c", i, c)
		}
	}

	// Try to decode base64 to validate
	if _, err := base64.StdEncoding.DecodeString(base64Data); err != nil {
		return nil, fmt.Errorf("base64 validation failed: %v", err)
	}

	requestBody := map[string]interface{}{
		"model": s.modelName,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": utils.GetIdentityUserDetailsPrompt(extractionFields),
				"images":  []string{base64Data},
			},
		},
		"stream": false,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.apiURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("non-success HTTP status: %s", resp.Status)
	}

	responseData, _ := io.ReadAll(resp.Body)

	var llamaResponse llamaResponse
	dataParseErr := json.Unmarshal(responseData, &llamaResponse)

	if dataParseErr != nil {

		return map[string]string{}, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	var result map[string]string

	parseErr := json.Unmarshal([]byte(llamaResponse.Message.Content), &result)
	if parseErr != nil {
		fmt.Println("Error:", err)
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", parseErr)
	}

	fmt.Printf("Clean AI data : %s", result)

	return result, nil

}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type llamaResponse struct {
	Model      string  `json:"model"`
	CreatedAt  string  `json:"created_at"`
	Message    message `json:"message"`
	DoneReason string  `json:"done_reason"`
	Done       bool    `json:"done"`
}
