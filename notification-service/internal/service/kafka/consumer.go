package service

import (
	"context"
	"encoding/json"
	"fmt"
	"notification-service/internal/common/api/dtos"
	"notification-service/internal/common/models"
	"notification-service/internal/constant"

	"notification-service/pkg/logger"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/segmentio/kafka-go"
)

// Refactored QueueHandlers to use interfaces
// (fields are now exported)
type QueueHandlers struct {
	NotificationRedis NotificationRedisInterface
	ConfigRepo        ConfigRepositoryInterface
	TemporalClient    TemporalClientInterface
	NotificationRepo  NotificationRepositoryInterface
	KafkaClient       KafkaPublisherInterface
}

// New constructor for dependency injection
func NewQueueHandlers(
	redis NotificationRedisInterface,
	configRepo ConfigRepositoryInterface,
	temporalClient TemporalClientInterface,
	notificationRepo NotificationRepositoryInterface,
	kafkaClient KafkaPublisherInterface,
) *QueueHandlers {
	return &QueueHandlers{
		NotificationRedis: redis,
		ConfigRepo:        configRepo,
		TemporalClient:    temporalClient,
		NotificationRepo:  notificationRepo,
		KafkaClient:       kafkaClient,
	}
}

type NotificationData struct {
	Type    string `json:"type"`
	Email   string `json:"email"`
	Body    string `json:"body"`
	Subject string `json:"subject"`
}

func (kh *QueueHandlers) HandleSendNotification(msg kafka.Message) {
	// convert json to struct
	notificationPayload := &dtos.NotificationRequest{}
	err := json.Unmarshal(msg.Value, notificationPayload)

	if err != nil {
		logger.Log.Println("Error in json unmarshal at kafka consumer", err.Error())
		return
	}

	logger.Log.Println("notificationPayload", notificationPayload)

	validate := validator.New()
	dtos.RegisterValidations(validate)
	// validate the payload
	if err := validate.Struct(notificationPayload); err != nil {
		logger.Log.Println("Error in validating the payload", err.Error())
		return
	}

	logger.Log.Printf("Valid Notification payload received")

	//check config from redis service
	configuration, err := kh.NotificationRedis.GetNotificationConfig("notification-config")

	if err != nil {
		logger.Log.Println("Error in getting notification config", err.Error())
	}

	var providerConfig []dtos.ChannelConfig

	if configuration == nil || len(configuration) == 0 {
		config, err := kh.ConfigRepo.GetConfig()
		if err != nil {
			logger.Log.Println("Error in getting notification config", err.Error())
		}
		if len(config) == 0 {
			logger.Log.Println("Notification config not found in DB")
			providerConfig = []dtos.ChannelConfig{
				{Service: "email", Primary: "sendgrid", Fallback: "smtp"},
				{Service: "sms", Primary: "twilio", Fallback: "fast2sms"},
				{Service: "whatsapp", Primary: "twilio", Fallback: "fast2sms"},
			}
		} else {
			providerConfig = config
		}
		// set data in redis
		kh.NotificationRedis.SetNotificationConfig("notification-config", providerConfig)
	} else {
		providerConfig = configuration
	}

	channelMaps := ChannelConfigSliceToMap(providerConfig)

	//loop and create object according to temporal needs
	var temporalPayload []dtos.NotificationTemporalPayload

	for _, recipient := range notificationPayload.Recipients {
		for _, channel := range notificationPayload.Channels {
			config, ok := channelMaps[channel]
			if !ok {
				logger.Log.Println(fmt.Sprintf("Invalid channel: %s", channel))
				continue
			}
			// Add job for primary
			temporalPayload = append(temporalPayload, dtos.NotificationTemporalPayload{
				TrackingID:       notificationPayload.TrackingID,
				Recipient:        recipient,
				Channel:          channel,
				PrimaryService:   config.Primary,
				FallbackService:  config.Fallback,
				Meta:             notificationPayload.Meta,
				Tags:             notificationPayload.Tags,
				TemplateID:       notificationPayload.TemplateID,
				RetryCount:       3,
				ExecutionTimeout: 10000,
			})
		}
	}

	logger.Log.Println("channelMaps", channelMaps)
	logger.Log.Println("temporalPayload", temporalPayload)

	// loop and execute temporal service for each payload
	for _, payload := range temporalPayload {

		jsonPayload, _ := json.Marshal(payload)
		var data map[string]interface{}
		err := json.Unmarshal(jsonPayload, &data)
		if err != nil {
			logger.Log.Println("Error in json unmarshal", err.Error())
		}

		// save to db and send the data to temporal
		notificationId, err := kh.NotificationRepo.SaveNotification(&models.Notification{
			Recipient: map[string]string{
				"user_id":         payload.Recipient.UserID,
				"email":           payload.Recipient.Email,
				"phone":           payload.Recipient.Phone,
				"whatsapp_number": payload.Recipient.WhatsappNumber,
			},
			Channel:            payload.Channel,
			PrimaryService:     payload.PrimaryService,
			FallbackService:    payload.FallbackService,
			Meta:               payload.Meta,
			Tags:               payload.Tags,
			TemplateID:         payload.TemplateID,
			PrimaryRetryCount:  0,
			FallbackRetryCount: 0,
			TotalExecutionTime: 0,
			NotificationStatus: models.NotificationStatusPending,
			TrackingId:         payload.TrackingID,
		})
		if err != nil {
			logger.Log.Println("Error saving notification", err.Error())
			return
		}
		data["id"] = *notificationId

		timeNow := time.Now().UnixMilli()
		switch payload.Channel {
		case "email":
			if payload.Recipient.Email != "" {
				kh.TemporalClient.ExecuteEmailWorkflow(data)
			} else {
				logger.Log.Println("Email not present, skipping email workflow")
			}
		case "sms":
			if payload.Recipient.Phone != "" {
				kh.TemporalClient.ExecuteSmsWorkflow(data)
			} else {
				logger.Log.Println("Phone number not present, skipping SMS workflow")
			}
		case "push":
			// Assuming push doesn't need validation, or add condition if needed
			kh.TemporalClient.ExecutePushNotificationWorkflow(data)
		case "whatsapp":
			if payload.Recipient.WhatsappNumber != "" {
				kh.TemporalClient.ExecuteWhatsappWorkflow(data)
			} else {
				logger.Log.Println("WhatsApp number not present, skipping WhatsApp workflow")
			}
		default:
			err := kh.NotificationRepo.UpdateNotificationByID(*notificationId, &models.Notification{
				NotificationStatus: models.NotificationStatusFailed,
				PrimaryRetryCount:  0,
				TotalExecutionTime: int(timeNow - time.Now().UnixMilli()),
				ErrorMessage:       "Invalid Channel",
				IsPrimaryFailed:    true,
				IsFallbackFailed:   true,
				IsMovedToDlq:       true,
			})
			if err != nil {
				logger.Log.Println("Error updating notification status", err.Error())
			}
			data["retryCount"] = int(data["retryCount"].(float64))
			data["notificationStatus"] = models.NotificationStatusFailed
			data["isFallbackFailed"] = true
			data["error_message"] = "Invalid Channel"
			data["isMovedToDlq"] = true

			jsonByte, err := json.Marshal(data)

			if err != nil {
				logger.Log.Println("Error marshaling notification data:", err.Error())
			}

			kh.KafkaClient.Publish(context.Background(), string(constant.NOTIFICATION_TOPIC), []byte(data["trackingId"].(string)), jsonByte)
			logger.Log.Println("Invalid channel")
		}

	}

}

func ChannelConfigSliceToMap(configs []dtos.ChannelConfig) map[string]dtos.ChannelConfig {
	mapped := make(map[string]dtos.ChannelConfig)
	for _, cfg := range configs {
		mapped[cfg.Service] = cfg
	}
	return mapped
}
