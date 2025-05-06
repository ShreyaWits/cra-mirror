package apiDtos

type SecurePayload struct {
	Data      interface{} `json:"data"`
	ExpiresAt int64       `json:"expires_at"`
	ModelType string      `json:"model_type"`
}
