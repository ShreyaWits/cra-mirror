package services

import (
    "Document-Processing/internal/utils"
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

type LlamaService struct {
    modelName string
    apiURL    string
    client    *http.Client
}

func NewLlamaService() *LlamaService {
    return &LlamaService{
        modelName: "llama3.2-vision",
        apiURL:    "http://localhost:11434/api/chat",
        client:    &http.Client{Timeout: 2 * time.Minute},
    }
}

func (s *LlamaService) ProcessImage(ctx context.Context, base64Image string, extractionFields []string) (map[string]string, error) {
    fmt.Printf("Llama Process Start")
    base64Data, _, err := cleanAndValidateBase64(base64Image)
    if err != nil {
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