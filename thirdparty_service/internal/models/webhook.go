package models

type Webhook struct {
	URL         string            `json:"url"`
	Environment string            `json:"environment"`
	ServiceName string            `json:"serviceName"`
	Method      string            `json:"method"`
	Values      map[string]string `json:"values"`
}
