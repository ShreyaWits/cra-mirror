package repositories

import (
	"fmt"
	commonDtos "protected_link/internal/common/api/dtos"
	configEnv "protected_link/internal/configs"
	"protected_link/pkg/jwt"
	database "protected_link/pkg/redis"
	"time"

	apiDtos "protected_link/internal/module/apis/dtos"
	"protected_link/internal/module/models"
	"protected_link/internal/module/utils"

	"github.com/redis/go-redis/v9"
)

type GeneratedRepository struct {
	redis  *database.RedisConfig
	config *configEnv.Config
}

// NewGeneratedRepository is a constructor function, not a type
func NewGeneratedRepository(redis *database.RedisConfig) *GeneratedRepository {
	return &GeneratedRepository{
		redis: redis,
	}
}

func (g *GeneratedRepository) SaveGeneratedLink(dto *apiDtos.GenerateUrlRequest) (*commonDtos.ApiResponseDto, error) {
	// Initialize the encryption service
	service, err := jwt.NewJwtCreation()
	if err != nil {
		return nil, fmt.Errorf("failed to init JwtCreation: %w", err)
	}

	// Encrypt the DTO and create token
	token, err := service.Encrypt(dto)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt DTO: %w", err)
	}

	shortCode := utils.GenerateShortCode(8)

	err = g.redis.Client.Set(g.redis.Ctx, shortCode, token, 24*time.Hour).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to save token in Redis: %w", err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to store shortCode in Redis: %w", err)
	}

	shortURL := fmt.Sprintf("%s/%s", g.config.RedirectionURL, shortCode)

	decrpyCode, err := g.GetOriginalToken(shortCode)
	time.Sleep(30 * time.Second)
	// 🔍 For testing: Immediately decrypt it to verify
	decrypted, err := service.Decrypt(decrpyCode)
	if err != nil {
		fmt.Println("❌ Failed to decrypt right after encrypt:", err)
	} else {
		fmt.Println("✅ Decryption test successful! Decrypted data:", string(decrypted))
	}

	// Construct the final URL
	//redirectURL := fmt.Sprintf("%s?token=%s", "https://example.com/auth", token)

	// Return wrapped in the expected response type
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
			return "", fmt.Errorf("short link not found")
		}
		return "", err
	}
	return token, nil
}
