package repositories

import (
	"encoding/json"
	"fmt"
	"time"

	commonDtos "protected_link/internal/common/api/dtos"
	"protected_link/internal/common/constants"
	messageUtility "protected_link/internal/common/utils"
	configEnv "protected_link/internal/configs"
	repository "protected_link/internal/modules/cassandra/repository"
	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"
	"protected_link/internal/modules/link_generation/models"
	link_utils "protected_link/internal/modules/link_generation/utils"
	"protected_link/pkg/jwt"
	database "protected_link/pkg/redis"

	"github.com/redis/go-redis/v9"
)

const (
	operationPrefix = "🔒"
)

type GeneratedRepository struct {
	redis      *database.RedisConfig
	config     *configEnv.Config
	jwtService *jwt.JwtCreation
	cassandra  repository.ICassandraRepository
}

// NewGeneratedRepository creates a new instance of GeneratedRepository
func NewGeneratedRepository(redis *database.RedisConfig, cassandra repository.ICassandraRepository) (*GeneratedRepository, error) {
	jwtService, err := jwt.NewJwtCreation()
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT service: %w", err)
	}

	cfg, err := configEnv.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &GeneratedRepository{
		redis:      redis,
		jwtService: jwtService,
		config:     cfg,
		cassandra:  cassandra,
	}, nil
}

// SaveGeneratedLink saves a generated link with the provided data
func (g *GeneratedRepository) SaveGeneratedLink(dto *apiDtos.GenerateUrlRequest) (*commonDtos.ApiResponseDto, error) {
	const operation = "SaveGeneratedLink"
	var data string

	// Handle hybrid model type
	if dto.ModelType == "hybrid" {
		id, err := g.cassandra.SaveData(dto)
		if err != nil {
			fmt.Printf("%s %s: failed to save token in Cassandra: %v\n", operationPrefix, operation, err)
			return nil, fmt.Errorf("failed to save token in Cassandra: %w", err)
		}
		data = id.String()
		fmt.Printf("%s %s: saved token in Cassandra: %s\n", operationPrefix, operation, id)
	}

	// Parse expiration time
	expiresAt, err := link_utils.ParseExpiration(dto.ExpireIn)
	if err != nil {
		fmt.Printf("%s %s: failed to parse expiration: %v\n", operationPrefix, operation, err)
		return nil, fmt.Errorf("failed to parse expiration: %w", err)
	}

	// Prepare payload data
	var payloadData interface{}
	if data == "" {
		payloadData = dto
	} else {
		payloadData = data
	}

	// Create secure payload
	payload := apiDtos.SecurePayload{
		Data:      payloadData,
		ExpiresAt: expiresAt,
		ModelType: dto.ModelType,
	}

	// Encrypt payload
	token, err := g.jwtService.Encrypt(payload)
	if err != nil {
		fmt.Printf("%s %s: failed to encrypt DTO: %v\n", operationPrefix, operation, err)
		return nil, fmt.Errorf("failed to encrypt DTO: %w", err)
	}

	// Generate shortcode and parse duration
	shortCode := link_utils.GenerateShortCode(8)
	duration, err := time.ParseDuration(dto.ExpireIn)
	if err != nil {
		return nil, fmt.Errorf("invalid expire_in format: %w", err)
	}

	// Save to Redis with TTL
	if err := g.redis.Client.Set(g.redis.Ctx, shortCode, token, duration).Err(); err != nil {
		fmt.Printf("%s %s: failed to save token in Redis: %v\n", operationPrefix, operation, err)
		return nil, fmt.Errorf("failed to save token in Redis: %w", err)
	}

	// Construct response
	shortURL := fmt.Sprintf("%s?token=%s", g.config.RedirectionURL, shortCode)
	return &commonDtos.ApiResponseDto{
		Success: true,
		Message: messageUtility.GetMessage(string(constants.ProtectedLinkGeneratedSuccessfully)),
		Data: &models.ProtectedLinkResponse{
			URL: shortURL,
		},
	}, nil
}

// GetOriginalToken retrieves the original token from Redis using the shortcode
func (g *GeneratedRepository) GetOriginalToken(shortCode string) (string, error) {
	const operation = "GetOriginalToken"
	token, err := g.redis.Client.Get(g.redis.Ctx, shortCode).Result()
	if err != nil {
		fmt.Printf("%s %s: failed to fetch token from Redis for shortCode %s: %v\n",
			operationPrefix, operation, shortCode, err)

		if err == redis.Nil {
			return "", fmt.Errorf("%s", messageUtility.GetMessage(string(constants.RequestLinkExpiredTitle)))
		}
		return "", fmt.Errorf("failed to get token: %w", err)
	}
	return token, nil
}

// GetTokenData retrieves and decrypts token data for a given link
func (g *GeneratedRepository) GetTokenData(link *string) (*apiDtos.SecurePayload, error) {
	const operation = "GetTokenData"

	// Get encrypted token
	encryptedToken, err := g.GetOriginalToken(*link)
	if err != nil {
		return nil, fmt.Errorf("%s: error retrieving token: %w", operation, err)
	}

	// Decrypt token
	decrypted, err := g.jwtService.Decrypt(encryptedToken)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to decrypt token: %w", operation, err)
	}

	// Unmarshal payload
	var dto apiDtos.SecurePayload
	if err := json.Unmarshal(decrypted, &dto); err != nil {
		return nil, fmt.Errorf("%s: failed to unmarshal decrypted payload: %w", operation, err)
	}

	// Delete key after use
	if err := g.redis.Client.Del(g.redis.Ctx, *link).Err(); err != nil {
		fmt.Printf("%s %s: failed to delete key after use: %v\n", operationPrefix, operation, err)
	}

	return &dto, nil
}

// DeleteShortCode deletes a shortcode from Redis
func (g *GeneratedRepository) DeleteShortCode(shortCode string) (*commonDtos.ApiResponseDto, error) {
	const operation = "DeleteShortCode"

	if err := g.redis.Client.Del(g.redis.Ctx, shortCode).Err(); err != nil {
		fmt.Printf("%s %s: failed to delete shortcode from Redis: %v\n", operationPrefix, operation, err)
		return nil, fmt.Errorf("failed to delete shortcode from Redis: %w", err)
	}

	return &commonDtos.ApiResponseDto{
		Success: true,
		Message: messageUtility.GetMessage(string(constants.ProtectedLinkDeletedSuccessfully)),
		Data: &models.ProtectedLinkResponse{
			URL: shortCode,
		},
	}, nil
}
