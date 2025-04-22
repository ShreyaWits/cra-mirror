package authRepository

import (
	"encoding/json"
	"fmt"
	"log"
	commonDtos "protected_link/internal/common/api/dtos"
	"protected_link/internal/modules/authentication/models"
	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"

	"context"

	"protected_link/kafka"
	kafkaService "protected_link/pkg/kafka"
	database "protected_link/pkg/redis"
	"time"
)

type OTPRepository struct {
	redisClient *database.RedisConfig
	service     *kafkaService.NotifierService
	ctx         context.Context
}

func NewOTPRepository(redisClient *database.RedisConfig) *OTPRepository {

	producer, err := kafka.NewKafkaProducer([]string{"kafka:9092"})
	if err != nil {
		log.Fatal("Error creating Kafka producer:", err)
	}
	defer producer.Close()

	notifier := kafkaService.NewNotifierService(producer)

	return &OTPRepository{
		redisClient: redisClient,
		ctx:         context.Background(),
		service:     notifier,
	}
}

func (r *OTPRepository) SendOtp(request apiDtos.GenerateUrlRequest, otp string) (*commonDtos.ApiResponseDto, error) {

	// Add the OTP to the request Data field
	if request.Data == nil {
		request.Data = make(map[string]interface{})
	}
	request.Data["otp"] = otp

	// Marshal the entire request to JSON
	payloadBytes, err := json.Marshal(request)

	if err != nil {
		return nil, fmt.Errorf("failed to marshal request for Redis: %w", err)
	}

	// Save in Redis
	key := fmt.Sprintf("otp:%s", request.UserID)

	expiry := time.Minute
	print("Generated OTP Check Expiry	:", expiry)
	print("Generated OTP Check Request	:", payloadBytes)

	// fallback default if needed

	if err := r.redisClient.Client.Set(r.ctx, key, payloadBytes, expiry).Err(); err != nil {
		return nil, fmt.Errorf("failed to save OTP data in Redis: %w", err)
	}

	print("Generatedd OTP Check Expiry	:", expiry)
	// Continue to send notification
	payload := models.MessagePayload{
		Channels: []string{"email", "sms", "whatsapp"},
		Recipients: []models.Recipient{
			{
				UserID: request.UserID,
				Email:  request.Email,
				Phone:  request.Phone,
				Data: map[string]string{
					"username": "Jane Doe",
					"otp":      otp,
				},
			},
		},
	}
	print("Generated OTP Check Expiry	:", expiry)

	err = r.service.SendNotification(payload, "notification-topic")
	if err != nil {
		log.Fatal("Failed to send notification:", err)
	}

	return &commonDtos.ApiResponseDto{
		Success: true,
		Message: "OTP sent successfully",
		Data:    &key,
	}, nil

}

func (r *OTPRepository) GetOTP(userID string) (string, error) {
	return r.redisClient.Client.Get(r.ctx, userID).Result()
}
