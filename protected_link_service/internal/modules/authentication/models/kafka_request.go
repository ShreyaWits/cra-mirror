package models

type MessagePayload struct {
	TemplateID string      `json:"templateID"`
	Channels   []string    `json:"channels"`
	Recipients []Recipient `json:"recipients"`
}

type Recipient struct {
	UserID string            `json:"userID"`
	Email  string            `json:"email"`
	Phone  string            `json:"phone"`
	Data   map[string]string `json:"data"`
}
