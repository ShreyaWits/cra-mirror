package errors

const (
	// EHxxx   (Encryption Handler)
	EHErrInvalidRequest = "EH001"

	// PRQxxx   (Pre API Request)
	PRQErrInvalidRequestFormat = "PRQ001"

	// KMGxxx   (Key Management)
	KMGErrInvalidKeyManagementRequest = "KMG001"
	KMGErrStoreKEK                    = "KMG002"
	KMGErrRetrieveKEK                 = "KMG003"
	KMGErrDeleteKEK                   = "KMG004"
	KMGErrListKEKs                    = "KMG005"
	KMGErrGenrateKEK                  = "KMG006"

	// ENGxxx   (Encryption Engine)
	ENGErrInvalidEncryptionEngineRequest = "ENG001"
	ENGErrGenerateDEK                    = "ENG002"
	ENGErrEncryptData                    = "ENG003"
	ENGErrDecryptData                    = "ENG004"
	ENGErrEncryptDEK                     = "ENG005"
	ENGErrDecryptDEK                     = "ENG006"

	// USRxxx   (User Service)
	USRErrInvalidUserServiceRequest = "USR001"
	USRErrFetchUserData             = "USR002"
	USRErrParseUserData             = "USR003"
	USRErrCreateUser                = "USR004"
	USRErrDeleteUser                = "USR005"
	USRErrUpdateUser                = "USR006"
	USRErrTokenRequired             = "USR007"

	// KMSxxx   (Key Management Service)
	KMSerrInvalidKeyManagementServiceRequest = "KMS001"
	KMSerrGenerateKEK                        = "KMS002"
	KMSerrRandomKey                          = "KMS003"
	KMSerrStoreKEK                           = "KMS004"
	KMSerrRetrieveKEK                        = "KMS005"
	KMSerrDeleteKEK                          = "KMS006"
	KMSerrListKEKs                           = "KMS007"
	KMSerrEncryptData                        = "KMS008"
	KMSerrDecryptData                        = "KMS009"

	// CRYPxxx   (Cryptography)
	CRYPerrInvalidCryptographyRequest = "CRYP001"
	CRYPerrEncryptData                = "CRYP002"
	CRYPerrDecryptData                = "CRYP003"
	CRYPerrEncryptBytes               = "CRYP004"
	CRYPerrDecryptBytes               = "CRYP005"

	// ESxxx   (Encryption Usecase)
	ESErrInvalidEncryptionUsecaseRequest = "ES001"
	ESErrRetrieveKEK                     = "ES002"
	ESErrDecryptDEK                      = "ES003"
	ESErrEncryptItemFields               = "ES004"
	ESErrEncrypt                         = "ES005"
	ESErrDecrypt                         = "ES006"
	ESErrGenerateEDEK                    = "ES007"
	ESErrETYPEKeyMissing                 = "ES008"
)
