package authRepository

import (
	"encoding/json"
	"fmt"
	"log"
	commonDtos "protected_link/internal/common/api/dtos"
	"protected_link/internal/common/constants"
	"protected_link/internal/common/utils"
	configEnv "protected_link/internal/configs"

	"protected_link/internal/modules/authentication/models"
	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"

	"context"

	"protected_link/kafka"
	kafkaService "protected_link/pkg/kafka"
	database "protected_link/pkg/redis"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type OTPRepository struct {
	redisClient *database.RedisConfig
	service     *kafkaService.NotifierService
	ctx         context.Context
	cfg         *configEnv.Config
}

func NewOTPRepository(redisClient *database.RedisConfig) *OTPRepository {

	cfg, _ := configEnv.LoadConfig()

	producer, err := kafka.NewKafkaProducer([]string{cfg.Kafka_Broker_Url})
	if err != nil {
		log.Fatal("Error creating Kafka producer:", err)
	}
	//defer producer.Close()

	notifier := kafkaService.NewNotifierService(producer)

	return &OTPRepository{
		redisClient: redisClient,
		ctx:         context.Background(),
		service:     notifier,
		cfg:         cfg,
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
		fmt.Println("Error marshalling request:", err)
		return nil, fmt.Errorf("failed to marshal request for Redis: %v", err)
	}

	// Save in Redis

	verificationID := uuid.New().String()
	key := fmt.Sprintf("%s-%s", verificationID, request.UserID)

	expiry := time.Minute

	// fallback default if needed

	if err := r.redisClient.Client.Set(r.ctx, key, payloadBytes, expiry).Err(); err != nil {
		fmt.Println("Error saving to Redis:", err)
		return nil, fmt.Errorf("failed to save OTP data in Redis: %w", err)
	}

	fmt.Println("Stored request Sending otp:", otp)

	//Continue to send notification
	payload := models.MessagePayload{
		Channels:   []string{request.ChannelType},
		TemplateID: "otp_verification",
		Recipients: []models.Recipient{
			{
				UserID: request.UserID,
				Email:  request.Email,
				Phone:  request.Phone,
				Data: map[string]string{
					"username": request.Name,
					"otp":      otp,
				},
			},
		},
	}

	err = r.service.SendNotification(payload, r.cfg.KafkaProducer)

	if err != nil {
		log.Fatal("Failed to send notification:", err)
	}

	return &commonDtos.ApiResponseDto{
		Success: true,
		Message: utils.GetMessage(string(constants.OtpSentSuccessfully)),
		Data: &models.VerificationRequest{
			VerificationID: key,
		},
	}, nil

}

func (r *OTPRepository) GetOTP(userID string) (string, error) {
	return r.redisClient.Client.Get(r.ctx, userID).Result()
}

func (r *OTPRepository) VerifyOtp(request *models.VerifyOTPRequest) (*commonDtos.ApiResponseDto, error) {

	verificationId := request.VerficationId
	userId := request.UserID
	otp := request.OTP

	// Get the data from Redis
	val, err := r.redisClient.Client.Get(r.ctx, verificationId).Result()

	if err == redis.Nil {
		return nil, fmt.Errorf("%s", utils.GetMessage(string(constants.OTPIncorrect)))
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
		return nil, fmt.Errorf("%s", utils.GetMessage(string(constants.OtpInvalidDescription)))
	}

	if storedRequest.UserID != userId {
		return &commonDtos.ApiResponseDto{
			Success: false,
			Message: utils.GetMessage(string(constants.VldUser)),
			Error:   utils.GetMessage(string(constants.InvalidUser)),
		}, nil

	}

	// Compare with provided OTP
	if storedOtp != otp {
		return &commonDtos.ApiResponseDto{
			Success: false,
			Message: utils.GetMessage(string(constants.OtpInvalid)),
			Error:   utils.GetMessage(string(constants.OtpInvalidDescription)),
		}, nil
	}

	delete(storedRequest.Data, "otp")
	// OTP matched: optionally delete the OTP from Redis
	_ = r.redisClient.Client.Del(r.ctx, verificationId)

	// Send back success with original data (optional)
	return &commonDtos.ApiResponseDto{
		Success: true,
		Message: utils.GetMessage(string(constants.OtpVerifiedSuccessfully)),
		Data:    storedRequest, // or maybe only parts of it if needed
	}, nil
}
