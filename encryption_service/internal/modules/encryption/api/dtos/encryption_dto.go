package dtos

// EncryptRequest represents the request body for encryption
type EncryptRequest struct {
	Data []map[string]interface{} `json:"data" validate:"required,dive,required"`
}

// EncryptResponse represents the response body for encryption
type EncryptResponse struct {
	Data []map[string]interface{} `json:"data" validate:"required,dive,required"`
}

// DecryptRequest represents the request body for decryption
type DecryptRequest struct {
	Data []map[string]interface{} `json:"data" validate:"required,dive,required"`
}

// DecryptResponse represents the response body for decryption
type DecryptResponse struct {
	Data []map[string]interface{} `json:"data" validate:"required,dive,required"`
}

// GenerateEDEKResponse represents the response body for EDEK generation
type GenerateEDEKResponse struct {
	EDEKPrivate string `json:"edekPrivate"`
	EDEKPublic  string `json:"edekPublic"`
}


