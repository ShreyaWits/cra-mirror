package models

type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type SMSRequest struct {
	To   string `json:"to"`
	Body string `json:"body"`
}
