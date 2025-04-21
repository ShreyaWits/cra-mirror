package dtos

// EncryptRequest represents the request body for encryption
type EncryptRequest struct {
	Data string `json:"data" validate:"required"`
}

// EncryptResponse represents the response body for encryption
type EncryptResponse struct {
	EncryptedData string `json:"encryptedData"`
}

// DecryptRequest represents the request body for decryption
type DecryptRequest struct {
	EncryptedData string `json:"encryptedData" validate:"required"`
}

// DecryptResponse represents the response body for decryption
type DecryptResponse struct {
	Data string `json:"data"`
}

// GenerateEDEKResponse represents the response body for EDEK generation
type GenerateEDEKResponse struct {
	EDEKPrivate string `json:"edekPrivate"`
	EDEKPublic  string `json:"edekPublic"`
}
