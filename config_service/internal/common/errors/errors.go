package common

type AppError struct {
	Status int    `json:"status"`
	Code   string `json:"errorCode"`
	Msg    string `json:"errorMsg"`
}

func (e AppError) Error() string {
	return e.Msg
}

var Errors = map[string]string{
	"CNF001":   "Invalid config data",
	"CNF002":   "Config key not found",
	"CNF003":   "ETCD connection failed",
	"CNF004":   "Invalid request body",
	"CNF005":   "Admin Secret not configured in environment",
	"CNF006":   "Invalid password",
	"CNF007":   "JWT secret not configured in environment",
	"AUTH001":  "Authorization header missing",
	"AUTH002":  "Invalid authorization header format",
	"AUTH003":  "Invalid or expired token",
	"CNF008":   "Invalid username",
	"CNF009":   "username not configured in environment",
	"CNF010":   "Invalid secret key",
	"CNF011":   "Database connection failed",
	"CNF012":   "username (email) should not be blank",
	"CNF013":   "password should not be blank",
	"CNF014":   "secret should not be blank",
	"ADMIN001": "Admin user already exists",
	"ADMIN002": "Admin user not found",
	"ADMIN003": "Invalid admin credentials",
	"AUTH004":  "Forbidden: You do not have permission to access this resource",
	"ADMIN010" :"User not found",
}

func ThrowError(status int, code string) AppError {
	msg, ok := Errors[code]
	if !ok {
		return AppError{400, "UNKNOWN", "Unknown error"}
	}
	return AppError{status, code, msg}
}
