package link_utils

import (
	"encoding/json"
	"fmt"
	"math/rand"
	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateShortCode(length int) string {
	rand.Seed(time.Now().UnixNano())
	code := make([]byte, length)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}

func ConvertToGenerateUrlRequest(data interface{}) (*apiDtos.GenerateUrlRequest, error) {
	// Step 1: Marshal interface{} into JSON
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal interface{}: %w", err)
	}

	// Step 2: Unmarshal into specific DTO
	var request apiDtos.GenerateUrlRequest
	if err := json.Unmarshal(bytes, &request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to GenerateUrlRequest: %w", err)
	}

	return &request, nil
}

func ParseExpiration(expireIn string) (int64, error) {
	duration, err := time.ParseDuration(expireIn)
	if err != nil {
		return 0, fmt.Errorf("invalid expire_in format: %w", err)
	}
	return time.Now().Add(duration).Unix(), nil
}
