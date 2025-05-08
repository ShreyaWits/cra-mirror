package repositories

import "third_party_service/internal/modules/third_party/models"

type ThirdPartyRepository interface {
	// SendEmail(emailRequest models.EmailRequest) error
	// SendSMS(smsRequest models.SMSRequest) error
	// SendPushNotification(notificationRequest models.PushNotificationRequest) error
	// SendWhatsAppMessage(whatsappRequest models.WhatsAppRequest) error
	// SendTelegramMessage(telegramRequest models.TelegramRequest) error
	// SendSlackMessage(slackRequest models.SlackRequest) error
	// SendDiscordMessage(discordRequest models.DiscordRequest) error
	// SendKafkaMessage(kafkaRequest models.KafkaRequest) error
	// SendRabbitMQMessage(rabbitMQRequest models.RabbitMQRequest) error
	// SendRedisMessage(redisRequest models.RedisRequest) error
	// SendSQSMessage(sqsRequest models.SQSRequest) error
	// SendSNSMessage(snsRequest models.SNSRequest) error
	// SendGCPMessage(gcpRequest models.GCPRequest) error
	// SendAzureMessage(azureRequest models.AzureRequest) error
	// SendTwilioMessage(twilioRequest models.TwilioRequest) error
	// SendSendGridMessage(sendGridRequest models.SendGridRequest) error
	// SendMailgunMessage(mailgunRequest models.MailgunRequest) error
	// SendAWSMessage(awsRequest models.AWSRequest) error
	// SendGoogleMessage(googleRequest models.GoogleRequest) error
	// SendAzureMessage(azureRequest models.AzureRequest) error
	// SendFirebaseMessage(firebaseRequest models.FirebaseRequest) error
	// SendOneSignalMessage(oneSignalRequest models.OneSignalRequest) error
	// SendPusherMessage(pusherRequest models.PusherRequest) error

}

type ThirdPartyKafkaRepository struct {
}

func NewThirdPartyRepository() *ThirdPartyKafkaRepository {
	return &ThirdPartyKafkaRepository{}
}

func (r *ThirdPartyKafkaRepository) SendEmail(emailReq models.EmailRequest) (models.EmailRequest, error) {
	// add logic for the send email via sendgrid
	return emailReq, nil
}
