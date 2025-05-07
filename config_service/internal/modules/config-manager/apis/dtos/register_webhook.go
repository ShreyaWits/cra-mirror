package dtos

type RegisterWebhookRequest struct {
	URL         string `json:"url" validate:"required"`
	Environment string `json:"environment" validate:"required"`
	ServiceName string `json:"serviceName" validate:"required"`
	Method      string `json:"method" validate:"required"`
}

func NewErrorResponse(statusCode int, message, err string) map[string]interface{} {
	return map[string]interface{}{
		"status_code": statusCode,
		"message":     message,
		"error":       err,
	}
}
