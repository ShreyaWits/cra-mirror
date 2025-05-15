package errors

const (
	// JWTxxx   (JWT Service Errors)
	JwtErrCreateGCM          = "JWT001"
	JwtErrDecodeString       = "JWT002"
	JwtErrCreateCipher       = "JWT003"
	JwtErrCiphertextTooShort = "JWT004"
	JwtErrDecryptFailed      = "JWT005"
	JwtErrUnmarshalFailed    = "JWT006"
	JwtErrTokenExpired       = "JWT007"

	JwtErrInvalidExpireIn        = "JWT008"
	JwtErrMarshalFailed          = "JWT009"
	JwtErrNonceGenerationFailed  = "JWT010"

	JwtErrLoadConfigFailed     = "JWT011"
	JwtErrInvalidSecretLength  = "JWT012"
	JwtErrReadNonce 		= "JWT013"
)


