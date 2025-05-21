package smtp

// SMTPConfig holds the configuration for connecting to the SMTP server
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}
