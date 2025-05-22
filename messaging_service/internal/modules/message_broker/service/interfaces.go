package service

import (
	"context"
	"messaging_service/internal/config"
)

// ConfigManagerServiceInterface defines the methods that a ConfigManagerService must implement
type ConfigManagerServiceInterface interface {
	GetFromApiConfiguration(ctx context.Context) (*config.Config, error)
	SetDataToCache(ctx context.Context, key string, cfg *config.Config) error
	GetDataToCache(ctx context.Context, key string) (*config.Config, error)
}
