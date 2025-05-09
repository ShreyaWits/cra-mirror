package schemas

type FieldError struct {
	Code    string `json:"code"`
	Field   string `json:"field"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

func (e FieldError) Error() string {
	return e.Message
}
