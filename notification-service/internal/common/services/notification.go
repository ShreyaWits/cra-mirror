package services

import (
	"context"
	"fmt"
	"log"
	"notification-service/internal/app"
	"notification-service/internal/common/api/dtos"
	"notification-service/internal/common/repositories"
	"notification-service/proto"
)

type NotificationServiceInterface interface {
	SaveConfig(ctx context.Context, configs []dtos.ChannelConfig) error
	GetConfigure() (*proto.GetConfigureResponse, error)
}

type NotificationService struct {
	configRepo repositories.ConfigRepositoryInterface // Repository field for saving configuration
}

// NewNotificationService initializes the NotificationService and injects the repository.
func NewNotificationService() *NotificationService {
	appInstance := app.Init()
	configRepo := appInstance.GetConfigRepo()

	return &NotificationService{
		configRepo: configRepo,
	}
}

// NewNotificationServiceWithRepo creates a new NotificationService with the provided repository.
func NewNotificationServiceWithRepo(repo repositories.ConfigRepositoryInterface) *NotificationService {
	return &NotificationService{
		configRepo: repo,
	}
}

// SaveConfig handles saving the configuration to the repository (Cassandra)
func (s *NotificationService) SaveConfig(ctx context.Context, configs []dtos.ChannelConfig) error {
	fmt.Println("Configure service post API called")
	// Here you call the method from the ConfigRepository to save configs in Cassandra
	return s.configRepo.SaveConfig(ctx, configs)
}

func (s *NotificationService) GetConfigure() (*proto.GetConfigureResponse, error) {
	// Fetch configurations from the repository
	configs, err := s.configRepo.GetConfig()
	if err != nil {
		log.Printf("Error fetching configuration: %v", err)
		return nil, err
	}

	// Map the DTOs to the response proto
	var protoConfigs []*proto.ChannelConfig
	for _, cfg := range configs {
		protoConfigs = append(protoConfigs, &proto.ChannelConfig{
			Service:  cfg.Service,
			Primary:  cfg.Primary,
			Fallback: cfg.Fallback,
		})
	}

	// Return the response
	return &proto.GetConfigureResponse{
		Status:     "success",
		StatusCode: 200,
		Message:    "Configurations fetched successfully",
		Configs:    protoConfigs,
	}, nil
}
