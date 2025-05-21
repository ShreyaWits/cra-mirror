package smtp

// EmailData holds the details of an email to be sent
type EmailData struct {
	To      string
	Subject string
	Body    string
}
