package common

type AppError struct {
	Code string `json:"errorCode"`
	Msg  string `json:"errorMsg"`
}

func (e AppError) Error() string {
	return e.Msg
}

var Errors = map[string]string{
	"CNF001": "Invalid config data",
	"CNF002": "Config key not found",
	"CNF003": "ETCD connection failed",
}

func ThrowError(code string) AppError {
	msg, ok := Errors[code]
	if !ok {
		return AppError{"UNKNOWN", "Unknown error"}
	}
	return AppError{code, msg}
}
