package repositories

import (
	commonDtos "protected_link/internal/common/api/dtos"
	apiDtos "protected_link/internal/modules/link_generation/apis/dtos"
)

// IGeneratedRepository defines the contract for the generated link repository operations
type IGeneratedRepository interface {
	// SaveGeneratedLink saves a generated link with the provided data
	SaveGeneratedLink(dto *apiDtos.GenerateUrlRequest) (*commonDtos.ApiResponseDto, error)

	// GetOriginalToken retrieves the original token from Redis using the shortcode
	GetOriginalToken(shortCode string) (string, error)

	// GetTokenData retrieves and decrypts token data for a given link
	GetTokenData(link *string) (*apiDtos.SecurePayload, error)

	// DeleteShortCode deletes a shortcode from Redis
	DeleteShortCode(shortCode string) (*commonDtos.ApiResponseDto, error)
}
