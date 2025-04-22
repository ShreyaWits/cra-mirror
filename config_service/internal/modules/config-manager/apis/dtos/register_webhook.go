package dtos

type RegisterWebhookRequest struct {
	URL         string `json:"url"`
	Environment string `json:"environment"`
	ServiceName string `json:"serviceName"`
	Method      string `json:"method"`
}
