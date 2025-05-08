package authRepository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	commonDtos "protected_link/internal/common/api/dtos"
	"protected_link/internal/common/constants"
	"protected_link/internal/common/utils"
	configEnv "protected_link/internal/configs"
	"protected_link/internal/modules/authentication/api/dtos"
	kafkaService "protected_link/internal/modules/authentication/messaging"
	"protected_link/internal/modules/authentication/models"
	repository "protected_link/internal/modules/cassandra/repository"
	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"
	kafka "protected_link/pkg/kafka"
	database "protected_link/pkg/redis"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// OTPRepository handles OTP-related operations
type OTPRepository struct {
	redisClient *database.RedisConfig
	service     *kafkaService.NotifierService
	ctx         context.Context
	cfg         *configEnv.Config
	cassendra   repository.ICassandraRepository
}

// NewOTPRepository creates a new instance of OTPRepository
func NewOTPRepository(redisClient *database.RedisConfig, casendra repository.ICassandraRepository) *OTPRepository {
	cfg, err := configEnv.LoadConfig()
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		return nil
	}

	producer, err := kafka.NewKafkaProducer([]string{cfg.Kafka_Broker_Url})
	if err != nil {
		log.Printf("Failed to create Kafka producer: %v", err)
		return nil
	}

	notifier := kafkaService.NewNotifierService(producer)

	return &OTPRepository{
		redisClient: redisClient,
		ctx:         context.Background(),
		service:     notifier,
		cfg:         cfg,
		cassendra:   casendra,
	}
}

// SendOtp sends an OTP to the user
func (r *OTPRepository) SendOtp(request apiDtos.GenerateUrlRequest, otp string, dbId string) (*commonDtos.ApiResponseDto, error) {
	if request.Data == nil {
		request.Data = make(map[string]interface{})
	}
	request.Data["otp"] = otp

	auth := dtos.AuthPayload{
		DbId: dbId,
		ID:   request.UserID,
		OTP:  otp,
	}

	var payloadRequest apiDtos.SecurePayload
	if request.ModelType == "hybrid" {
		payloadRequest = apiDtos.SecurePayload{
			Data:      auth,
			ModelType: request.ModelType,
		}
	} else {
		payloadRequest = apiDtos.SecurePayload{
			Data:      request,
			ModelType: request.ModelType,
		}
	}

	payloadBytes, err := json.Marshal(payloadRequest)
	if err != nil {
		log.Printf("Failed to marshal request: %v", err)
		return nil, fmt.Errorf("failed to marshal request for Redis: %w", err)
	}

	verificationID := uuid.New().String()
	key := fmt.Sprintf("%s-%s", verificationID, request.UserID)
	expiry := time.Minute

	if err := r.redisClient.Client.Set(r.ctx, key, payloadBytes, expiry).Err(); err != nil {
		log.Printf("Failed to save to Redis: %v", err)
		return nil, fmt.Errorf("failed to save OTP data in Redis: %w", err)
	}

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

	if err := r.service.SendNotification(payload, r.cfg.KafkaProducer); err != nil {
		log.Printf("Failed to send notification: %v", err)
		return nil, fmt.Errorf("failed to send notification: %w", err)
	}

	return &commonDtos.ApiResponseDto{
		Success: true,
		Message: utils.GetMessage(string(constants.OtpSentSuccessfully)),
		Data: &models.VerificationRequest{
			VerificationID: key,
		},
	}, nil
}

// GetOTP retrieves an OTP for a user
func (r *OTPRepository) GetOTP(userID string) (string, error) {
	return r.redisClient.Client.Get(r.ctx, userID).Result()
}

// VerifyOtp verifies an OTP for a user
func (r *OTPRepository) VerifyOtp(request *models.VerifyOTPRequest) (*commonDtos.ApiResponseDto, error) {
	val, err := r.redisClient.Client.Get(r.ctx, request.VerificationID).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("%s", utils.GetMessage(string(constants.OTPIncorrect)))
	} else if err != nil {
		return nil, fmt.Errorf("failed to fetch OTP from Redis: %w", err)
	}

	var res apiDtos.SecurePayload
	if err := json.Unmarshal([]byte(val), &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OTP data: %w", err)
	}

	if res.ModelType == "hybrid" {
		return r.verifyHybridOTP(res, request)
	}
	return r.verifyStandardOTP(res, request)
}

// verifyHybridOTP verifies OTP for hybrid model
func (r *OTPRepository) verifyHybridOTP(res apiDtos.SecurePayload, request *models.VerifyOTPRequest) (*commonDtos.ApiResponseDto, error) {
	dataMap, ok := res.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected res.Data to be map[string]interface{}, got %T", res.Data)
	}

	dataBytes, err := json.Marshal(dataMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data map: %w", err)
	}

	var payload dtos.AuthPayload
	if err := json.Unmarshal(dataBytes, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal into AuthPayload: %w", err)
	}

	if payload.ID != request.UserID {
		return &commonDtos.ApiResponseDto{
			Success: false,
			Message: utils.GetMessage(string(constants.VldUser)),
			Error:   utils.GetMessage(string(constants.InvalidUser)),
		}, nil
	}

	if payload.OTP != request.OTP {
		return &commonDtos.ApiResponseDto{
			Success: false,
			Message: utils.GetMessage(string(constants.OtpInvalid)),
			Error:   utils.GetMessage(string(constants.OtpInvalidDescription)),
		}, nil
	}

	generateUrlRequest, err := r.cassendra.GetDataByID(payload.DbId)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve token data: %w", err)
	}

	_ = r.redisClient.Client.Del(r.ctx, request.VerificationID)

	return &commonDtos.ApiResponseDto{
		Success: true,
		Message: utils.GetMessage(string(constants.OtpVerifiedSuccessfully)),
		Data:    generateUrlRequest,
	}, nil
}

// verifyStandardOTP verifies OTP for standard model
func (r *OTPRepository) verifyStandardOTP(res apiDtos.SecurePayload, request *models.VerifyOTPRequest) (*commonDtos.ApiResponseDto, error) {
	jsonData, err := json.Marshal(res.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal res.Data: %w", err)
	}

	var storedRequest apiDtos.GenerateUrlRequest
	if err := json.Unmarshal(jsonData, &storedRequest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON into GenerateUrlRequest: %w", err)
	}

	storedOtp, ok := storedRequest.Data["otp"].(string)
	if !ok {
		return nil, fmt.Errorf("%s", utils.GetMessage(string(constants.OtpInvalidDescription)))
	}

	if storedRequest.UserID != request.UserID {
		return &commonDtos.ApiResponseDto{
			Success: false,
			Message: utils.GetMessage(string(constants.VldUser)),
			Error:   utils.GetMessage(string(constants.InvalidUser)),
		}, nil
	}

	if storedOtp != request.OTP {
		return &commonDtos.ApiResponseDto{
			Success: false,
			Message: utils.GetMessage(string(constants.OtpInvalid)),
			Error:   utils.GetMessage(string(constants.OtpInvalidDescription)),
		}, nil
	}

	delete(storedRequest.Data, "otp")
	_ = r.redisClient.Client.Del(r.ctx, request.VerificationID)

	return &commonDtos.ApiResponseDto{
		Success: true,
		Message: utils.GetMessage(string(constants.OtpVerifiedSuccessfully)),
		Data:    storedRequest,
	}, nil
}
