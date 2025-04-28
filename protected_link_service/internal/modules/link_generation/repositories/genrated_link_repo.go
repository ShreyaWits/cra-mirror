package repositories

import (
	"encoding/json"
	"fmt"
	commonDtos "protected_link/internal/common/api/dtos"
	"protected_link/internal/common/constants"
	configEnv "protected_link/internal/configs"

	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"
	"protected_link/internal/modules/link_generation/models"
	"protected_link/internal/modules/link_generation/utils"

	messageUtility "protected_link/internal/common/utils"
	"protected_link/pkg/jwt"
	database "protected_link/pkg/redis"
	"time"

	"github.com/redis/go-redis/v9"
)

type GeneratedRepository struct {
	redis      *database.RedisConfig
	config     *configEnv.Config
	jwtService *jwt.JwtCreation
}

func NewGeneratedRepository(redis *database.RedisConfig) *GeneratedRepository {
	jwtService, _ := jwt.NewJwtCreation()
	cfg, _ := configEnv.LoadConfig()

	return &GeneratedRepository{
		redis:      redis,
		jwtService: jwtService,
		config:     cfg,
	}
}

func (g *GeneratedRepository) SaveGeneratedLink(dto *apiDtos.GenerateUrlRequest) (*commonDtos.ApiResponseDto, error) {
	// Encrypt the DTO to generate a token
	token, err := g.jwtService.Encrypt(dto)

	if err != nil {
		fmt.Println("🔒 Encrypting DTO:", err)
		return nil, fmt.Errorf("failed to encrypt DTO: %w", err)
	}

	// Generate a unique shortcode
	shortCode := utils.GenerateShortCode(8)

	// Parse expiration string like "30s", "1h"
	duration, err := time.ParseDuration(dto.ExpireIn)
	if err != nil {
		return nil, fmt.Errorf("invalid expire_in format: %w", err)
	}

	// Save the token to Redis with TTL
	err = g.redis.Client.Set(g.redis.Ctx, shortCode, token, duration).Err()
	if err != nil {
		fmt.Println("🔒 Saving token in Redis:", err)
		return nil, fmt.Errorf("failed to save token in Redis: %w", err)
	}

	// Construct short URL
	shortURL := fmt.Sprintf("%s?token=%s", g.config.RedirectionURL, shortCode)

	return &commonDtos.ApiResponseDto{
		Success: true,
		Message: messageUtility.GetMessage(string(constants.ProtectedLinkGeneratedSuccessfully)),
		Data: &models.ProtectedLinkResponse{
			URL: shortURL,
		},
	}, nil
}

func (g *GeneratedRepository) GetOriginalToken(shortCode string) (string, error) {
	token, err := g.redis.Client.Get(g.redis.Ctx, shortCode).Result()
	if err != nil {

		println("🔒 Fetching token from Redis for shortCode:", err)
		if err == redis.Nil {
			println("🔒 Fetching token from Redis for shortCode: sadadsadsads", err)
			return "", fmt.Errorf("%s", messageUtility.GetMessage(string(constants.RequestLinkExpiredTitle)))
		}
		return "", err
	}

	// Clean up: delete the token from Redis after first use
	// Defer deletion after return

	return token, nil
}

func (g *GeneratedRepository) GetTokenData(link *string) (*commonDtos.ApiResponseDto, error) {
	// Get encrypted token from Redis using the short link

	encryptedToken, err := g.GetOriginalToken(*link)
	if err != nil {

		return nil, fmt.Errorf("error retrieving token: %w", err)
	}

	// Decrypt the token
	decrypted, err := g.jwtService.Decrypt(encryptedToken)
	if err != nil {
		return nil, fmt.Errorf("❌ failed to decrypt token: %w", err)
	}

	// Unmarshal into expected DTO
	var dto apiDtos.SecurePayload
	if err := json.Unmarshal(decrypted, &dto); err != nil {
		return nil, fmt.Errorf("failed to unmarshal decrypted payload: %w", err)
	}

	///Check if OTP is required
	if delErr := g.redis.Client.Del(g.redis.Ctx, *link).Err(); delErr != nil {
		fmt.Println("⚠️ Failed to delete key after use:", delErr)
	}

	return &commonDtos.ApiResponseDto{
		Success: true,
		Message: messageUtility.GetMessage(string(constants.DataFetchedSuccessfully)),
		Data:    dto.Data,
	}, nil
}

func (g *GeneratedRepository) DeleteShortCode(shortCode string) (*commonDtos.ApiResponseDto, error) {
	err := g.redis.Client.Del(g.redis.Ctx, shortCode).Err()
	if err != nil {
		return nil, fmt.Errorf("❌ failed to delete shortcode from Redis: %w", err)
	}

	return &commonDtos.ApiResponseDto{
		Success: true,
		Message: messageUtility.GetMessage(string(constants.ProtectedLinkDeletedSuccessfully)),
		Data: &models.ProtectedLinkResponse{
			URL: shortCode,
		},
	}, nil
}
