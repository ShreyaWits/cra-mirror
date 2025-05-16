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
		// PRQxxx   (Pre API Request)
		PRQErrInvalidRequestFormat: "Invalid request format :-",

		// EHxxx   (Encryption Handler)
		EHErrInvalidRequest: "Bad Request: Invalid request format.",

		// ESxxx   (Encryption Usecase)
		ESErrInvalidEncryptionUsecaseRequest: "Invalid encryption use case request.",
		ESErrRetrieveKEK:                     "Failed to retrieve KEK.",
		ESErrDecryptDEK:                      "Failed to decrypt DEK.",
		ESErrEncryptItemFields:               "Failed to encrypt item fields.",
		ESErrEncrypt:                         "Failed to encrypt data.",
		ESErrDecrypt:                         "Failed to decrypt data.",
		ESErrGenerateEDEK:                    "Failed to generate EDEK.",
		ESErrETYPEKeyMissing:                 "e_type Error",

		// KMGxxx   (Key Management)
		KMGErrInvalidKeyManagementRequest: "Key management request failed.",
		KMGErrStoreKEK:                    "Failed to store KEK.",
		KMGErrRetrieveKEK:                 "Failed to retrieve KEK.",
		KMGErrDeleteKEK:                   "Failed to delete KEK.",
		KMGErrListKEKs:                    "Failed to list KEKs.",

		// ENGxxx   (Encryption Engine)
		ENGErrInvalidEncryptionEngineRequest: "Encryption engine request failed.",
		ENGErrGenerateDEK:                    "Failed to generate DEK.",
		ENGErrEncryptData:                    "Failed to encrypt data.",
		ENGErrDecryptData:                    "Failed to decrypt data.",
		ENGErrEncryptDEK:                     "Failed to encrypt DEK.",
		ENGErrDecryptDEK:                     "Failed to decrypt DEK.",

		// USRxxx   (User Service)
		USRErrInvalidUserServiceRequest: "Bad Request: Invalid request format.",
		USRErrFetchUserData:             "Failed to fetch user data.",
		USRErrParseUserData:             "Failed to parse user data.",
		USRErrCreateUser:                "Failed to create user.",
		USRErrDeleteUser:                "Failed to delete user.",
		USRErrUpdateUser:                "Failed to update user.",

		// KMSxxx   (Key Management Service)
		KMSerrInvalidKeyManagementServiceRequest: "Bad Request: Invalid request format.",
		KMSerrGenerateKEK:                        "Failed to generate KEK.",
		KMSerrRandomKey:                          "Failed to generate random KEK.",
		KMSerrStoreKEK:                           "Failed to store KEK.",
		KMSerrRetrieveKEK:                        "Failed to retrieve KEK.",
		KMSerrDeleteKEK:                          "Failed to delete KEK.",
		KMSerrListKEKs:                           "Failed to list KEKs.",
		KMSerrEncryptData:                        "Failed to encrypt data with Vault.",
		KMSerrDecryptData:                        "Failed to decrypt data with Vault.",

		// CRYPxxx   (Cryptography)
		CRYPerrInvalidCryptographyRequest: "Bad Request: Invalid request format.",
		CRYPerrEncryptData:                "Failed to encrypt data.",
		CRYPerrDecryptData:                "Failed to decrypt data.",
		CRYPerrEncryptBytes:               "Failed to encrypt bytes.",
		CRYPerrDecryptBytes:               "Failed to decrypt bytes.",
	}

	if message, exists := statusMessages[code]; exists {
		return message
	}
	return ""
}
