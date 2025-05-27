package interfaces

import (
	"context"

	"thirdparty_service/internal/modules/config/dto"
)

// ConfigManagerServiceInterface defines the interface for the config manager service.
// This is defined in a separate package to avoid import cycles.
type ConfigManagerServiceInterface interface {
	SetDataToCache(ctx context.Context, key string, cfg *dto.ConfigResponse) error
	GetDataToCache(ctx context.Context, key string) (*dto.ConfigResponse, error)
}