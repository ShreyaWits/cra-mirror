package dtos

type RegisterWebhookRequest struct {
	URL         string `json:"url"`
	Environment string `json:"environment"`
	ServiceName string `json:"serviceName"`
	Method      string `json:"method"`
}

func NewErrorResponse(statusCode int, message, err string) map[string]interface{} {
	return map[string]interface{}{
		"status_code": statusCode,
		"message":     message,
		"error":       err,
	}
}
