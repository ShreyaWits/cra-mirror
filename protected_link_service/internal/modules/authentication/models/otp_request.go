package models

type MessagePayload struct {
	Channels   []string    `json:"channels"`
	Recipients []Recipient `json:"recipients"`
}

type Recipient struct {
	UserID string            `json:"userID"`
	Email  string            `json:"email"`
	Phone  string            `json:"phone"`
	Data   map[string]string `json:"data"`
}
