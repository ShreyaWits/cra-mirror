package models

import "time"


type Config struct {
	ServiceName string                 `json:"service_name"`
	Environment string                 `json:"environment"`
	ConfigData  map[string]interface{} `json:"config_data"`
	Version     string                 `json:"version"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}


type ConfigMetadata struct {
	LastModifiedBy string    `json:"last_modified_by"`
	ChangeHistory  []string  `json:"change_history"`
	LastModifiedAt time.Time `json:"last_modified_at"`
}
