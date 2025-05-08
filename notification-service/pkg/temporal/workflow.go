package temporal

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"notification-service/internal/common/models"
	"notification-service/internal/common/repositories"
	"notification-service/pkg/config"
	"notification-service/pkg/firebase"
	"notification-service/pkg/kafka"
	"notification-service/pkg/sendgrid"
	"notification-service/pkg/smtp"
	"notification-service/pkg/twilio"
	"notification-service/pkg/whatsapp"
	"strconv"
	"strings"
	"time"

	temporalSDK "go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type TemporalWorkflow struct {
	Client                 *http.Client
	ActivityRequestTimeout *int
	ActivityRetryCount     *int
	notificationRepo       repositories.NotificationRepositoryInterface
	kafkaClient            *kafka.KafkaPublisher
}

func NewTemporalWorkflow(ActivityRequestTimeout, ActivityRetryCount *int, notificationRepo repositories.NotificationRepositoryInterface, kafkaClient *kafka.KafkaPublisher) *TemporalWorkflow {
	log.Println("Initialized Email Executor Workflow.")
	return &TemporalWorkflow{
		Client:                 &http.Client{},
		ActivityRequestTimeout: ActivityRequestTimeout,
		ActivityRetryCount:     ActivityRetryCount,
		notificationRepo:       notificationRepo,
		kafkaClient:            kafkaClient,
	}
}

func (w *TemporalWorkflow) ExecuteEmailWorkflow(ctx workflow.Context, props map[string]interface{}) (map[string]interface{}, error) {

	log.Println("Executing Email Executor Workflow......", props)

	activityRetryCount := *w.ActivityRetryCount
	activityRequestTimeout := *w.ActivityRequestTimeout
	RuleRetryExponentialBackoffInterval := 0

	if props["retryCount"] != 0 {
		activityRetryCount = int(props["retryCount"].(float64))
	}

	if props["executionTimeout"] != 0 {
		activityRequestTimeout = int(props["executionTimeout"].(float64))
	}

	// Define activity options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Millisecond * time.Duration(activityRequestTimeout),
		RetryPolicy: &temporalSDK.RetryPolicy{
			InitialInterval: time.Millisecond * time.Duration(RuleRetryExponentialBackoffInterval),
			MaximumAttempts: int32(activityRetryCount),
		},
	}

	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var activityResult map[string]interface{}
	var err error

	exeTimeStamp := time.Now().UnixMilli()

	// Call appropriate activity based on type
	switch strings.ToUpper(props["primaryService"].(string)) {
	case string(EmailSMTPProvider):
		err = workflow.ExecuteActivity(ctx, w.InitSMTPActivity, props).Get(ctx, &activityResult)
	case string(EmailSendGridProvider):
		err = workflow.ExecuteActivity(ctx, w.InitSendGridActivity, props).Get(ctx, &activityResult)
	default:
		log.Println("Unknown Email Provider")
		err = errors.New("Unknown Email Provider")
	}

	log.Println("Activity Result----------------------", activityResult)
	log.Println("Activity err------------------------", err)

	// Handle activity execution result and errors
	if err != nil {
		log.Println("Failed to execute activity", "Error", err)
		// update the db of failure
		w.notificationRepo.UpdateNotificationByID(props["id"].(string), &models.Notification{
			PrimaryRetryCount:  int(props["retryCount"].(float64)),
			NotificationStatus: models.NotificationStatusFailed,
			IsPrimaryFailed:    true,
			ErrorMessage:       err.Error(),
		})
		// if fallback service exist retry fallback service
		if props["fallbackService"] != "" {
			var fallbackResult map[string]interface{}
			var err error
			// Call appropriate activity based on type
			switch strings.ToUpper(props["fallbackService"].(string)) {
			case string(EmailSMTPProvider):
				err = workflow.ExecuteActivity(ctx, w.InitSMTPActivity, props).Get(ctx, &fallbackResult)
			case string(EmailSendGridProvider):
				err = workflow.ExecuteActivity(ctx, w.InitSendGridActivity, props).Get(ctx, &fallbackResult)
			default:
				log.Println("Unknown Email Provider")
				err = errors.New("Unknown Email Provider Inside Fallback")
			}

			if err != nil {
				// update the db of failure
				w.notificationRepo.UpdateNotificationByID(props["id"].(string), &models.Notification{
					FallbackRetryCount: int(props["retryCount"].(float64)),
					NotificationStatus: models.NotificationStatusFailed,
					IsFallbackFailed:   true,
					ErrorMessage:       err.Error(),
					IsMovedToDlq:       true,
				})

				props["retryCount"] = int(props["retryCount"].(float64))
				props["notificationStatus"] = models.NotificationStatusFailed
				props["isFallbackFailed"] = true
				props["error_message"] = err.Error()
				props["isMovedToDlq"] = true

				jsonByte, err := json.Marshal(props)

				if err != nil {
					log.Println("Error marshaling notification data:", err.Error())
				}

				w.kafkaClient.Publish(context.Background(), "notification_dlq", []byte(props["trackingId"].(string)), jsonByte)

			}

			if fallErrMap, ok := fallbackResult["error"].(map[string]interface{}); ok {
				// update the db of failure
				w.notificationRepo.UpdateNotificationByID(props["id"].(string), &models.Notification{
					FallbackRetryCount: int(props["retryCount"].(float64)),
					NotificationStatus: models.NotificationStatusFailed,
					IsFallbackFailed:   true,
					ErrorMessage:       fallErrMap["message"].(string),
					IsMovedToDlq:       true,
				})

				props["retryCount"] = int(props["retryCount"].(float64))
				props["notificationStatus"] = models.NotificationStatusFailed
				props["isFallbackFailed"] = true
				props["error_message"] = fallErrMap["message"].(string)
				props["isMovedToDlq"] = true

				jsonByte, err := json.Marshal(props)

				if err != nil {
					log.Println("Error marshaling notification data:", err.Error())
				}

				w.kafkaClient.Publish(context.Background(), "notification_dlq", []byte(props["trackingId"].(string)), jsonByte)
			}

		}

		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	// Check if activity result contains an error
	if errorMap, ok := activityResult["error"].(map[string]interface{}); ok {
		log.Println("Failed to execute activity", "Error", err)
		// update the db of failure
		w.notificationRepo.UpdateNotificationByID(props["id"].(string), &models.Notification{
			PrimaryRetryCount:  int(props["retryCount"].(float64)),
			NotificationStatus: models.NotificationStatusFailed,
			IsPrimaryFailed:    true,
			ErrorMessage:       errorMap["message"].(string),
		})
		// if fallback service exist retry fallback service
		if props["fallbackService"] != "" {
			var fallbackResult map[string]interface{}
			var err error
			// Call appropriate activity based on type
			switch strings.ToUpper(props["fallbackService"].(string)) {
			case string(EmailSMTPProvider):
				err = workflow.ExecuteActivity(ctx, w.InitSMTPActivity, props).Get(ctx, &fallbackResult)
			case string(EmailSendGridProvider):
				err = workflow.ExecuteActivity(ctx, w.InitSendGridActivity, props).Get(ctx, &fallbackResult)
			default:
				log.Println("Unknown Email Provider")
				err = errors.New("Unknown Email Provider Inside Fallback")
			}

			if err != nil {
				// update the db of failure
				w.notificationRepo.UpdateNotificationByID(props["id"].(string), &models.Notification{
					FallbackRetryCount: int(props["retryCount"].(float64)),
					NotificationStatus: models.NotificationStatusFailed,
					IsFallbackFailed:   true,
					ErrorMessage:       err.Error(),
					TotalExecutionTime: int(time.Now().UnixMilli()),
					IsMovedToDlq:       true,
				})

				props["retryCount"] = int(props["retryCount"].(float64))
				props["notificationStatus"] = models.NotificationStatusFailed
				props["isFallbackFailed"] = true
				props["error_message"] = err.Error()
				props["isMovedToDlq"] = true

				jsonByte, err := json.Marshal(props)

				if err != nil {
					log.Println("Error marshaling notification data:", err.Error())
				}

				w.kafkaClient.Publish(context.Background(), "notification_dlq", []byte(props["trackingId"].(string)), jsonByte)
			}

			if fallbackErrMap, ok := fallbackResult["error"].(map[string]interface{}); ok {
				// update the db of failure
				w.notificationRepo.UpdateNotificationByID(props["id"].(string), &models.Notification{
					FallbackRetryCount: int(props["retryCount"].(float64)),
					NotificationStatus: models.NotificationStatusFailed,
					IsFallbackFailed:   true,
					ErrorMessage:       fallbackErrMap["message"].(string),
					TotalExecutionTime: int(exeTimeStamp - time.Now().UnixMilli()),
					IsMovedToDlq:       true,
				})
				props["retryCount"] = int(props["retryCount"].(float64))
				props["notificationStatus"] = models.NotificationStatusFailed
				props["isFallbackFailed"] = true
				props["error_message"] = fallbackErrMap["message"].(string)
				props["isMovedToDlq"] = true

				jsonByte, err := json.Marshal(props)

				if err != nil {
					log.Println("Error marshaling notification data:", err.Error())
				}

				w.kafkaClient.Publish(context.Background(), "notification_dlq", []byte(props["trackingId"].(string)), jsonByte)
			}

			// update the db on success
			w.notificationRepo.UpdateNotificationByID(props["id"].(string), &models.Notification{
				FallbackRetryCount: 0,
				NotificationStatus: models.NotificationStatusSuccess,
				TotalExecutionTime: int(exeTimeStamp - time.Now().UnixMilli()),
			})

			// Successfully executed activity
			return map[string]interface{}{
				"success": true,
				"result":  fallbackResult,
			}, nil

		}
		if errorMsg, ok := errorMap["message"].(string); ok && errorMsg != "" {
			log.Println("Activity resulted in error", "Error", errorMsg)
			return map[string]interface{}{
				"success": false,
				"error":   errorMsg,
			}, nil
		}
	}

	// update the db on success
	w.notificationRepo.UpdateNotificationByID(props["id"].(string), &models.Notification{
		FallbackRetryCount: 0,
		NotificationStatus: models.NotificationStatusSuccess,
		TotalExecutionTime: int(exeTimeStamp - time.Now().UnixMilli()),
	})

	// Successfully executed activity
	return map[string]interface{}{
		"success": true,
		"result":  activityResult,
	}, nil
}
func (w *TemporalWorkflow) ExecuteSMSWorkflow(ctx workflow.Context, props map[string]interface{}) (map[string]interface{}, error) {

	activityRetryCount := *w.ActivityRetryCount
	activityRequestTimeout := *w.ActivityRequestTimeout
	RuleRetryExponentialBackoffInterval := 0

	if props["retryCount"] != 0 {
		activityRetryCount = int(props["retryCount"].(float64))
	}

	if props["executionTimeout"] != 0 {
		activityRequestTimeout = int(props["executionTimeout"].(float64))
	}

	exeTimeStamp := time.Now().UnixMilli()

	// Define activity options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Millisecond * time.Duration(activityRequestTimeout),
		RetryPolicy: &temporalSDK.RetryPolicy{
			InitialInterval: time.Millisecond * time.Duration(RuleRetryExponentialBackoffInterval),
			MaximumAttempts: int32(activityRetryCount),
		},
	}

	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var activityResult map[string]interface{}
	var err error

	// Call appropriate activity based on type
	switch strings.ToUpper(props["primaryService"].(string)) {
	case string(SMSTwilioProvider):
		err = workflow.ExecuteActivity(ctx, w.InitTwilioSMSActivity, props).Get(ctx, &activityResult)
	default:
		return map[string]interface{}{
			"success": false,
			"error":   errors.New("Unknown SMS Provider"),
		}, nil
	}

	// Handle activity execution result and errors
	if err != nil {
		log.Println("Failed to execute activity", "Error", err)
		// update the db of failure
		w.notificationRepo.UpdateNotificationByID(props["id"].(string), &models.Notification{
			PrimaryRetryCount:  int(props["retryCount"].(float64)),
			NotificationStatus: models.NotificationStatusFailed,
			IsPrimaryFailed:    true,
			ErrorMessage:       err.Error(),
			IsMovedToDlq:       true,
		})

		props["retryCount"] = int(props["retryCount"].(float64))
		props["notificationStatus"] = models.NotificationStatusFailed
		props["isFallbackFailed"] = true
		props["error_message"] = err.Error()
		props["isMovedToDlq"] = true

		jsonByte, err := json.Marshal(props)

		if err != nil {
			log.Println("Error marshaling notification data:", err.Error())
		}

		w.kafkaClient.Publish(context.Background(), "notification_dlq", []byte(props["trackingId"].(string)), jsonByte)

		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	// Check if activity result contains an error
	if errorMap, ok := activityResult["error"].(map[string]interface{}); ok {
		log.Println("Failed to execute activity", "Error", err)
		// update the db of failure
		w.notificationRepo.UpdateNotificationByID(props["id"].(string), &models.Notification{
			PrimaryRetryCount:  int(props["retryCount"].(float64)),
			NotificationStatus: models.NotificationStatusFailed,
			IsPrimaryFailed:    true,
			ErrorMessage:       errorMap["message"].(string),
			IsMovedToDlq:       true,
		})
		props["retryCount"] = int(props["retryCount"].(float64))
		props["notificationStatus"] = models.NotificationStatusFailed
		props["isFallbackFailed"] = true
		props["error_message"] = errorMap["message"].(string)
		props["isMovedToDlq"] = true

		jsonByte, err := json.Marshal(props)

		if err != nil {
			log.Println("Error marshaling notification data:", err.Error())
		}

		w.kafkaClient.Publish(context.Background(), "notification_dlq", []byte(props["trackingId"].(string)), jsonByte)
		if errorMsg, ok := errorMap["message"].(string); ok && errorMsg != "" {
			log.Println("Activity resulted in error", "Error", errorMsg)
			return map[string]interface{}{
				"success": false,
				"error":   errorMsg,
			}, nil
		}
	}

	// update the db on success
	w.notificationRepo.UpdateNotificationByID(props["id"].(string), &models.Notification{
		FallbackRetryCount: 0,
		NotificationStatus: models.NotificationStatusSuccess,
		TotalExecutionTime: int(exeTimeStamp - time.Now().UnixMilli()),
	})

	// Successfully executed activity
	return map[string]interface{}{
		"success": true,
		"result":  activityResult,
	}, nil
}
func (w *TemporalWorkflow) ExecuteWhatsappWorkflow(ctx workflow.Context, props map[string]interface{}) (map[string]interface{}, error) {
	log.Println("Executing Whatsapp Executor Workflow......", props)
	activityRetryCount := *w.ActivityRetryCount
	activityRequestTimeout := *w.ActivityRequestTimeout
	RuleRetryExponentialBackoffInterval := 0

	if props["retryCount"] != 0 {
		activityRetryCount = int(props["retryCount"].(float64))
	}

	if props["executionTimeout"] != 0 {
		activityRequestTimeout = int(props["executionTimeout"].(float64))
	}

	exeTimeStamp := time.Now().UnixMilli()

	// Define activity options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Millisecond * time.Duration(activityRequestTimeout),
		RetryPolicy: &temporalSDK.RetryPolicy{
			InitialInterval: time.Millisecond * time.Duration(RuleRetryExponentialBackoffInterval),
			MaximumAttempts: int32(activityRetryCount),
		},
	}

	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var activityResult map[string]interface{}
	var err error
	log.Println("Service value in props:", props["service"])

	// Call appropriate activity based on type
	switch strings.ToUpper(props["primaryService"].(string)) {
	case string(WhatsappTwilioProvider):
		err = workflow.ExecuteActivity(ctx, w.InitTwilioWhatsappActivity, props).Get(ctx, &activityResult)
	default:
		return map[string]interface{}{
			"success": false,
			"error":   errors.New("Unknown SMS Provider"),
		}, nil
	}

	// Handle activity execution result and errors
	if err != nil {
		log.Println("Failed to execute activity", "Error", err)
		// update the db of failure
		w.notificationRepo.UpdateNotificationByID(props["id"].(string), &models.Notification{
			PrimaryRetryCount:  int(props["retryCount"].(float64)),
			NotificationStatus: models.NotificationStatusFailed,
			IsPrimaryFailed:    true,
			ErrorMessage:       err.Error(),
			IsMovedToDlq:       true,
		})

		props["retryCount"] = int(props["retryCount"].(float64))
		props["notificationStatus"] = models.NotificationStatusFailed
		props["isFallbackFailed"] = true
		props["error_message"] = err.Error()
		props["isMovedToDlq"] = true

		jsonByte, err := json.Marshal(props)

		if err != nil {
			log.Println("Error marshaling notification data:", err.Error())
		}

		w.kafkaClient.Publish(context.Background(), "notification_dlq", []byte(props["trackingId"].(string)), jsonByte)
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	// Check if activity result contains an error
	if errorMap, ok := activityResult["error"].(map[string]interface{}); ok {
		log.Println("Failed to execute activity", "Error", err)
		// update the db of failure
		w.notificationRepo.UpdateNotificationByID(props["id"].(string), &models.Notification{
			PrimaryRetryCount:  int(props["retryCount"].(float64)),
			NotificationStatus: models.NotificationStatusFailed,
			IsPrimaryFailed:    true,
			ErrorMessage:       errorMap["message"].(string),
			IsMovedToDlq:       true,
		})

		props["retryCount"] = int(props["retryCount"].(float64))
		props["notificationStatus"] = models.NotificationStatusFailed
		props["isFallbackFailed"] = true
		props["error_message"] = errorMap["message"].(string)
		props["isMovedToDlq"] = true

		jsonByte, err := json.Marshal(props)

		if err != nil {
			log.Println("Error marshaling notification data:", err.Error())
		}

		w.kafkaClient.Publish(context.Background(), "notification_dlq", []byte(props["trackingId"].(string)), jsonByte)

		if errorMsg, ok := errorMap["message"].(string); ok && errorMsg != "" {
			log.Println("Activity resulted in error", "Error", errorMsg)
			return map[string]interface{}{
				"success": false,
				"error":   errorMsg,
			}, nil
		}
	}

	// update the db on success
	w.notificationRepo.UpdateNotificationByID(props["id"].(string), &models.Notification{
		FallbackRetryCount: 0,
		NotificationStatus: models.NotificationStatusSuccess,
		TotalExecutionTime: int(exeTimeStamp - time.Now().UnixMilli()),
	})

	// Successfully executed activity
	return map[string]interface{}{
		"success": true,
		"result":  activityResult,
	}, nil
}
func (w *TemporalWorkflow) ExecutePushNotificationWorkflow(ctx workflow.Context, props map[string]interface{}) (map[string]interface{}, error) {

	activityRetryCount := *w.ActivityRetryCount
	activityRequestTimeout := *w.ActivityRequestTimeout
	RuleRetryExponentialBackoffInterval := 0

	if props["retryCount"] != 0 {
		activityRetryCount = int(props["retryCount"].(float64))
	}

	if props["executionTimeout"] != 0 {
		activityRequestTimeout = int(props["executionTimeout"].(float64))
	}

	exeTimeStamp := time.Now().UnixMilli()

	// Define activity options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Millisecond * time.Duration(activityRequestTimeout),
		RetryPolicy: &temporalSDK.RetryPolicy{
			InitialInterval: time.Millisecond * time.Duration(RuleRetryExponentialBackoffInterval),
			MaximumAttempts: int32(activityRetryCount),
		},
	}

	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var activityResult map[string]interface{}
	var err error

	// Call appropriate activity based on type
	switch strings.ToUpper(props["primaryService"].(string)) {
	case string(PushFirebaseProvider):
		err = workflow.ExecuteActivity(ctx, w.InitFirebasePushActivity, props).Get(ctx, &activityResult)
	default:
		return map[string]interface{}{
			"success": false,
			"error":   errors.New("Unknown SMS Provider"),
		}, nil
	}

	// Handle activity execution result and errors
	if err != nil {
		log.Println("Failed to execute activity", "Error", err)
		// update the db of failure
		w.notificationRepo.UpdateNotificationByID(props["id"].(string), &models.Notification{
			PrimaryRetryCount:  int(props["retryCount"].(float64)),
			NotificationStatus: models.NotificationStatusFailed,
			IsPrimaryFailed:    true,
			ErrorMessage:       err.Error(),
			IsMovedToDlq:       true,
		})

		props["retryCount"] = int(props["retryCount"].(float64))
		props["notificationStatus"] = models.NotificationStatusFailed
		props["isFallbackFailed"] = true
		props["error_message"] = err.Error()
		props["isMovedToDlq"] = true

		jsonByte, err := json.Marshal(props)

		if err != nil {
			log.Println("Error marshaling notification data:", err.Error())
		}

		w.kafkaClient.Publish(context.Background(), "notification_dlq", []byte(props["trackingId"].(string)), jsonByte)

		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	// Check if activity result contains an error
	if errorMap, ok := activityResult["error"].(map[string]interface{}); ok {
		log.Println("Failed to execute activity", "Error", err)
		// update the db of failure
		w.notificationRepo.UpdateNotificationByID(props["id"].(string), &models.Notification{
			PrimaryRetryCount:  int(props["retryCount"].(float64)),
			NotificationStatus: models.NotificationStatusFailed,
			IsPrimaryFailed:    true,
			ErrorMessage:       errorMap["message"].(string),
			IsMovedToDlq:       true,
		})

		props["retryCount"] = int(props["retryCount"].(float64))
		props["notificationStatus"] = models.NotificationStatusFailed
		props["isFallbackFailed"] = true
		props["error_message"] = errorMap["message"].(string)
		props["isMovedToDlq"] = true

		jsonByte, err := json.Marshal(props)

		if err != nil {
			log.Println("Error marshaling notification data:", err.Error())
		}

		w.kafkaClient.Publish(context.Background(), "notification-dlq", []byte(props["trackingId"].(string)), jsonByte)

		if errorMsg, ok := errorMap["message"].(string); ok && errorMsg != "" {
			log.Println("Activity resulted in error", "Error", errorMsg)
			return map[string]interface{}{
				"success": false,
				"error":   errorMsg,
			}, nil
		}
	}

	// update the db on success
	w.notificationRepo.UpdateNotificationByID(props["id"].(string), &models.Notification{
		FallbackRetryCount: 0,
		NotificationStatus: models.NotificationStatusSuccess,
		TotalExecutionTime: int(exeTimeStamp - time.Now().UnixMilli()),
	})

	// Successfully executed activity
	return map[string]interface{}{
		"success": true,
		"result":  activityResult,
	}, nil
}

func (w *TemporalWorkflow) InitSMTPActivity(props map[string]interface{}) (map[string]interface{}, error) {
	log.Println("InitSMTPActivity", props)

	smtpHost := config.GetEnv("SMTP_HOST", "smtp.google.com")
	smtpPort := config.GetEnv("SMTP_PORT", "443")
	smtpUser := config.GetEnv("SMTP_USER", "loushikk.giri@gmail.com")
	smtpPassword := config.GetEnv("SMTP_PASSWORD", "11477885655")

	smtpPortNum, err := strconv.Atoi(smtpPort)
	if err != nil {
		log.Println("PORT IS NOT A VALID INT")
		smtpPortNum = 443
	}

	if props == nil {
		errMsg := "Email Payload is Empty"
		return nil, errors.New(errMsg)
	}

	recipient, ok := props["recipient"].(map[string]interface{})
	if !ok {
		errMsg := "Invalid recipient format"
		return nil, errors.New(errMsg)
	}

	emailSensingErr := smtp.SMTP(&smtp.EmailConfig{
		SMTPHost:     smtpHost,
		SMTPPort:     smtpPortNum,
		SMTPUser:     smtpUser,
		SMTPPassword: smtpPassword,
	}, recipient["email"].(string), "Test Email", "Hello There")

	if emailSensingErr != nil {
		return nil, emailSensingErr
	}

	result := map[string]interface{}{
		"success": true,
	}
	return result, nil
}

func (w *TemporalWorkflow) InitSendGridActivity(props map[string]interface{}) (map[string]interface{}, error) {

	log.Println("InitSendGridActivity", props)

	sendgridApiKey := config.GetEnv("SENDGRID_APIKEY", "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")
	fromEmail := config.GetEnv("SENDGRID_EMAIL", "kajal.raj@thewitslab.com")

	if props == nil {
		errMsg := "Email Payload is Empty"
		return nil, errors.New(errMsg)
	}

	recipient, ok := props["recipient"].(map[string]interface{})
	if !ok {
		errMsg := "Invalid recipient format"
		return nil, errors.New(errMsg)
	}

	log.Println("recipient", recipient)

	SDService := sendgrid.InitSendGrid(sendgridApiKey,
		fromEmail,
		recipient["email"].(string),
		"Test Email",
		"Hello There")
	_, err := SDService.SendEmail()

	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{
		"success": true,
	}
	return result, nil
}
func (w *TemporalWorkflow) InitTwilioSMSActivity(props map[string]interface{}) (map[string]interface{}, error) {
	log.Println("InitTwilioActivity", props)

	accountSid := config.GetEnv("TWILIO_ACCOUNT_SID", "AC2f0b338bfe121bd56a62e358d39b2ad3")
	authToken := config.GetEnv("TWILIO_AUTH_TOKEN", "d66315cdbebbf8c1e5f7183682b09760")
	fromPhone := config.GetEnv("TWILIO_PHONE_NUMBER", "+15055788243")

	if props == nil {
		errMsg := "Payload is Empty"
		return nil, errors.New(errMsg)
	}
	recipient, ok := props["recipient"].(map[string]interface{})
	if !ok {
		errMsg := "Invalid recipient format"
		return nil, errors.New(errMsg)
	}

	twilioClient := twilio.NewTwilioClient(accountSid, authToken, fromPhone)

	err := twilioClient.SendSMS(recipient["phone"].(string), "Hello There")

	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{
		"success": true,
	}
	return result, nil
}

func (w *TemporalWorkflow) InitTwilioWhatsappActivity(props map[string]interface{}) (map[string]interface{}, error) {

	log.Println("InitTwilioWhatsappActivity", props)

	if props == nil {
		errMsg := "Payload is Empty"
		return nil, errors.New(errMsg)
	}
	recipient, ok := props["recipient"].(map[string]interface{})
	if !ok {
		errMsg := "Invalid recipient format"
		return nil, errors.New(errMsg)
	}

	whatsappSendingErr := whatsapp.SendWhatsAppMessage(recipient["whatsapp"].(string), "Hello There")
	if whatsappSendingErr != nil {
		return nil, whatsappSendingErr
	}
	result := map[string]interface{}{
		"success": true,
	}
	return result, nil
}
func (w *TemporalWorkflow) InitFirebasePushActivity(props map[string]interface{}) (map[string]interface{}, error) {

	log.Println("InitFirebaseActivity", props)

	if props == nil {
		errMsg := "Payload is Empty"
		return nil, errors.New(errMsg)
	}

	firebasePushErr := firebase.SendPushNotification(props["token"].(string), "Hello There", "Hey Whatsapp")
	if firebasePushErr != nil {
		return nil, firebasePushErr
	}
	result := map[string]interface{}{
		"success": true,
	}
	return result, nil
}
