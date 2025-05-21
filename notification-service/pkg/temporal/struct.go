package temporal

type EmailProvider string

const (
	EmailSMTPProvider     EmailProvider = "SMTP"
	EmailSendGridProvider EmailProvider = "SENDGRID"
)

type SMSProvider string

const (
	SMSTwilioProvider SMSProvider = "TWILIO"
)

type WhatsappProvider string

const (
	WhatsappTwilioProvider WhatsappProvider = "TWILIO"
)

type PushProvider string

const (
	PushFirebaseProvider PushProvider = "FIREBASE"
)
