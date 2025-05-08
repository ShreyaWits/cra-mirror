package models

// ChannelConfig defines configuration for a single notification service
type ChannelConfig struct {
	Service  string `json:"service"`  // e.g., email, sms, push, whatsapp
	Primary  string `json:"primary"`  // e.g., sendgrid, msg91
	Fallback string `json:"fallback"` // e.g., smtp, textlocal
}

// NotificationConfig holds all channel configurations
type NotificationConfig struct {
	Configs []ChannelConfig `json:"configs"`
}
