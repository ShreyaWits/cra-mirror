package constants

type ServerErrorMessageCode string

const (
	RequestBodyInvalidContentType ServerErrorMessageCode = "SRV001"
	RequestBodyInvalidJSON        ServerErrorMessageCode = "SRV002"
	RequestInvalidFormat          ServerErrorMessageCode = "SRV003"
	RequestValidationFailed       ServerErrorMessageCode = "SRV004"

	InternalServerError   ServerErrorMessageCode = "SRV005"
	FaliedToSaveLink      ServerErrorMessageCode = "SRV006"
	InvalidRequestPayload ServerErrorMessageCode = "SRV007"
	FailedToDeleteLink    ServerErrorMessageCode = "SRV008"
)

type SuccessMessageCode string

const (
	LinkGeneratedSuccessfully          SuccessMessageCode = "SUC001"
	ProtectedLinkGeneratedSuccessfully SuccessMessageCode = "SUC002"
	DataFetchedSuccessfully            SuccessMessageCode = "SUC003"
	ProtectedLinkDeletedSuccessfully   SuccessMessageCode = "SUC004"
)

type ValidationMessageCode string

const (
	VldUser                       ValidationMessageCode = "VLD013"
	VldMobileRequired             ValidationMessageCode = "VLD014"
	VldInvalidEmailFormat         ValidationMessageCode = "VLD015"
	VldUserIDAndOTPRequired       ValidationMessageCode = "VLD016"
	InvalidValueShouldBeOneOfList ValidationMessageCode = "VLD017"
	ValidationFailedFor           ValidationMessageCode = "VLD018"
)

type AuthenticationMessageCode string

const (
	UnauthorizedAccess      AuthenticationMessageCode = "AUTH001"
	InvalidCredentials      AuthenticationMessageCode = "AUTH002"
	AccessTokenExpired      AuthenticationMessageCode = "AUTH003"
	RefreshTokenExpired     AuthenticationMessageCode = "AUTH004"
	InvalidAccessToken      AuthenticationMessageCode = "AUTH005"
	InvalidRefreshToken     AuthenticationMessageCode = "AUTH006"
	AuthTokenMissing        AuthenticationMessageCode = "AUTH007"
	AccessDenied            AuthenticationMessageCode = "AUTH008"
	LoginSuccess            AuthenticationMessageCode = "AUTH009"
	LogoutSuccess           AuthenticationMessageCode = "AUTH010"
	TokenGenerationFailed   AuthenticationMessageCode = "AUTH011"
	TokenVerificationFailed AuthenticationMessageCode = "AUTH012"
	AccountInactive         AuthenticationMessageCode = "AUTH013"
	AccountLocked           AuthenticationMessageCode = "AUTH014"
	SessionExpired          AuthenticationMessageCode = "AUTH015"
	InvalidUser             AuthenticationMessageCode = "VLD019"

	OtpSentSuccessfully     AuthenticationMessageCode = "AUTH016"
	OtpSendFailed           AuthenticationMessageCode = "AUTH017"
	OtpVerificationFailed   AuthenticationMessageCode = "AUTH018"
	OtpExpired              AuthenticationMessageCode = "AUTH019"
	OtpVerifiedSuccessfully AuthenticationMessageCode = "AUTH020"
	FailedToSaveOTP         AuthenticationMessageCode = "AUTH021"
	OtpInvalid              AuthenticationMessageCode = "AUTH022"

	UserRegistrationSuccess    AuthenticationMessageCode = "AUTH023"
	EmailAlreadyExists         AuthenticationMessageCode = "AUTH024"
	MobileAlreadyExists        AuthenticationMessageCode = "AUTH025"
	UsernameAlreadyExists      AuthenticationMessageCode = "AUTH026"
	RegistrationValidationFail AuthenticationMessageCode = "AUTH027"
	UserCreationFailed         AuthenticationMessageCode = "AUTH028"
	FailedToRetrieveOTP        AuthenticationMessageCode = "AUTH029"

	RequestLinkExpiredDesc  AuthenticationMessageCode = "AUTH030"
	RequestLinkExpiredTitle AuthenticationMessageCode = "AUTH031"
	OtpInvalidDescription   AuthenticationMessageCode = "AUTH032"
	OTPIncorrect            AuthenticationMessageCode = "AUTH033"
)
