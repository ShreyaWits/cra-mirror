package repositories

import (
	"encoding/json"
	"fmt"
	commonDtos "protected_link/internal/common/api/dtos"
	configEnv "protected_link/internal/configs"

	apiDtos "protected_link/internal/modules/protected_link_generation/apis/dtos"
	"protected_link/internal/modules/protected_link_generation/models"
	"protected_link/internal/modules/protected_link_generation/utils"

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
		return nil, fmt.Errorf("failed to save token in Redis: %w", err)
	}

	// Construct short URL
	shortURL := fmt.Sprintf("%s?token=%s", g.config.RedirectionURL, shortCode)

	return &commonDtos.ApiResponseDto{
		Success: true,
		Message: "Protected link generated successfully",
		Data: &models.ProtectedLinkResponse{
			URL: shortURL,
		},
	}, nil
}

func (g *GeneratedRepository) GetOriginalToken(shortCode string) (string, error) {
	token, err := g.redis.Client.Get(g.redis.Ctx, shortCode).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("⏰ link is expired or does not exist")
		}
		return "", err
	}

	// Clean up: delete the token from Redis after first use
	if err := g.redis.Client.Del(g.redis.Ctx, shortCode).Err(); err != nil {
		fmt.Println("⚠️ Failed to delete key after use:", err)
	}

	return token, nil
}

func (g *GeneratedRepository) GetTokenData(link *string) (*commonDtos.ApiResponseDto, error) {
	// Get encrypted token from Redis using the short link
	encryptedToken, err := g.GetOriginalToken(*link)
	if err != nil {
		return nil, err
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

	return &commonDtos.ApiResponseDto{
		Success: true,
		Message: "Data fetched successfully",
		Data:    dto.Data,
	}, nil
}
