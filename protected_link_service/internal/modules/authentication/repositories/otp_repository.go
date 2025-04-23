package authRepository

import (
	"encoding/json"
	"fmt"
	"log"
	commonDtos "protected_link/internal/common/api/dtos"

	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"

	"context"

	"protected_link/kafka"
	kafkaService "protected_link/pkg/kafka"
	database "protected_link/pkg/redis"
	"time"

	"github.com/redis/go-redis/v9"
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
		log.Println("Request Data is nil, initializing it")
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

	// fallback default if needed

	if err := r.redisClient.Client.Set(r.ctx, key, payloadBytes, expiry).Err(); err != nil {
		return nil, fmt.Errorf("failed to save OTP data in Redis: %w", err)
	}

	print("Generatedd OTP Check Expiry	:", expiry)
	// Continue to send notification
	// payload := models.MessagePayload{
	// 	Channels: []string{"email", "sms", "whatsapp"},
	// 	Recipients: []models.Recipient{
	// 		{
	// 			UserID: request.UserID,
	// 			Email:  request.Email,
	// 			Phone:  request.Phone,
	// 			Data: map[string]string{
	// 				"username": "Jane Doe",
	// 				"otp":      otp,
	// 			},
	// 		},
	// 	},
	// }

	// err = r.service.SendNotification(payload, "notification-topic")
	// if err != nil {
	// 	log.Fatal("Failed to send notification:", err)
	// }

	return &commonDtos.ApiResponseDto{
		Success: true,
		Message: "OTP sent successfully",
		Data:    &key,
	}, nil

}

func (r *OTPRepository) GetOTP(userID string) (string, error) {
	return r.redisClient.Client.Get(r.ctx, userID).Result()
}

func (r *OTPRepository) VerifyOtp(userId string, providedOtp string) (*commonDtos.ApiResponseDto, error) {
	key := fmt.Sprintf("otp:%s", userId)

	// Get the data from Redis
	val, err := r.redisClient.Client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("OTP not found or expired")
	} else if err != nil {
		return nil, fmt.Errorf("failed to fetch OTP from Redis: %w", err)
	}

	// Unmarshal into GenerateUrlRequest DTO
	var storedRequest apiDtos.GenerateUrlRequest
	err = json.Unmarshal([]byte(val), &storedRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal OTP data: %w", err)
	}

	// Extract stored OTP
	storedOtp, ok := storedRequest.Data["otp"].(string)
	if !ok {
		return nil, fmt.Errorf("stored OTP is not valid or missing")
	}

	// Compare with provided OTP
	if storedOtp != providedOtp {
		return &commonDtos.ApiResponseDto{
			Success: false,
			Message: "Invalid OTP provided",
		}, nil
	}

	delete(storedRequest.Data, "otp")
	// OTP matched: optionally delete the OTP from Redis
	_ = r.redisClient.Client.Del(r.ctx, key)

	// Send back success with original data (optional)
	return &commonDtos.ApiResponseDto{
		Success: true,
		Message: "OTP verified successfully",
		Data:    storedRequest, // or maybe only parts of it if needed
	}, nil
}
