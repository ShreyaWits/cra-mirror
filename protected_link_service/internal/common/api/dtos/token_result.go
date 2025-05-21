package commonDtos

// TokenResult represents the decrypted token data
type TokenResult struct {
	Data      interface{} `json:"data"`
	ExpiresAt int64       `json:"expires_at"`
	ModelType string      `json:"model_type"`
}
