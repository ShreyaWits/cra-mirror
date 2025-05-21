package errors

import "fmt"

// GetErrorMessage returns the error message related to a specific HTTP status code.
func GetErrorMessage(code int) string {
	statusMessages := map[int]string{
		// 4XX Client Errors
		400: "Bad Request: Invalid request format.",
		401: "Unauthorized: Authentication required.",
		403: "Forbidden: Access is denied.",
		404: "Not Found: Resource does not exist.",
		405: "Method Not Allowed: Unsupported request method.",
		409: "Conflict: Duplicate or conflicting request.",
		422: "Unprocessable Entity: Validation failed.",

		// 5XX Server Errors
		500: "Internal Server Error: Something went wrong.",
		502: "Bad Gateway: Invalid upstream response.",
		503: "Service Unavailable: Server is overloaded or under maintenance.",
		504: "Gateway Timeout: No response from upstream service.",
	}

	if message, exists := statusMessages[code]; exists {
		return message
	}
	return fmt.Sprintf("Something went wrong%d", code)
}

// GetErrorMessage returns the error message related to a specific HTTP status code.
func GetAppErrorMessage(code string) string {
	statusMessages := map[string]string{
		// JWTxxx   (JWT Service)
		JwtErrCreateGCM:          "Failed to create GCM block cipher",
		JwtErrDecodeString:       "Failed to decode base64 string",
		JwtErrCreateCipher:       "Failed to create AES cipher",
		JwtErrCiphertextTooShort: "Ciphertext is too short",
		JwtErrDecryptFailed:      "Failed to decrypt ciphertext",
		JwtErrUnmarshalFailed:    "Failed to parse decrypted JSON",
		JwtErrTokenExpired:       "Link has expired",

		JwtErrInvalidExpireIn:    "Invalid expiration time",
		JwtErrMarshalFailed:      "Failed to marshal JSON",
		JwtErrNonceGenerationFailed: "Failed to generate nonce",

		JwtErrLoadConfigFailed:     "failed to load configuration",
		JwtErrInvalidSecretLength:  "JWT secret must be exactly 32 bytes long",
		JwtErrReadNonce:           "Failed to read nonce from file",
	
	}

	if message, exists := statusMessages[code]; exists {
		return message
	}
	return ""
}
