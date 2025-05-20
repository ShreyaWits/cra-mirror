package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	common "nps-config-service/internal/common/errors"
	"nps-config-service/internal/configs/db"
	"nps-config-service/internal/modules/config-manager/models"
	etcdDB "nps-config-service/pkg/etcd"
	"nps-config-service/pkg/observability"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type ConfigRepository struct {
	EtcdClient         *etcdDB.EtcdClientImpl
	ObservabilityStack *observability.ObservabilityStack
}

type IConfigRepo interface {
	StoreConfig(ctx context.Context, serviceName, environment string, configData map[string]interface{}) (interface{}, error)
	GetConfig(ctx context.Context, serviceName, environment string) (map[string]interface{}, error)
	GetConfigValue(ctx context.Context, serviceName, environment, key string) (interface{}, error)
	SetEtcdKey(ctx context.Context, key string, data string, ttl time.Duration) error
	GetEtcdKey(ctx context.Context, key string) (string, error)
	DeleteEtcdKey(ctx context.Context, key string) error
	CreateAdmin(ctx context.Context, admin *models.Admin) (*models.Admin, error)
	GetAdminByCredentials(ctx context.Context, username, password string) (*models.Admin, error)
}

func NewConfigRepository(etcdClient *etcdDB.EtcdClientImpl, observabilityStack *observability.ObservabilityStack) IConfigRepo {
	return &ConfigRepository{EtcdClient: etcdClient, ObservabilityStack: observabilityStack}
}

func (r *ConfigRepository) StoreConfig(ctx context.Context, serviceName, environment string, configData map[string]interface{}) (interface{}, error) {
	ctx, span := r.ObservabilityStack.TracerService.Start(ctx, "StoreConfig")
	defer span.End()
	// Store each config field as a separate key
	baseKey := fmt.Sprintf("%s/%s", environment, serviceName)

	log.Printf("Storing config with base key: %s", baseKey)
	changeHistory := []string{}
	// Store config data fields
	b, err := json.Marshal(configData)
	if err != nil {
		fmt.Errorf("failed to marshal config body: %v", err)
	}

	if err := r.EtcdClient.PutKey(baseKey, string(b)); err != nil {
		log.Printf("Error storing key %s: %v", baseKey, err)
		return nil, fmt.Errorf("failed to store config field %s: %v", baseKey, err)
	}

	// for key, value := range configData {
	// 	configKey := fmt.Sprintf("%s/%s", baseKey, key)
	// 	valueStr := fmt.Sprintf("%v", value)
	// 	log.Printf("Storing key: %s, value: %s", configKey, valueStr)
	// 	existingVal, err := r.GetConfigValue(serviceName, environment, key)
	// 	if err != nil {
	// 		changeHistory = append(changeHistory, fmt.Sprintf("added: %s", key))
	// 	} else if existingVal != valueStr {
	// 		changeHistory = append(changeHistory, fmt.Sprintf("updated: %s", key))
	// 	}
	// }

	log.Println("Config data stored successfully. Change history: ", changeHistory)
	// Store metadata, code is commented out for now, will use it later when we have to store metadata
	// now := time.Now()
	// type ConfigMetadata struct {
	// 	LastModifiedBy string    `json:"last_modified_by"`
	// 	ChangeHistory  []string  `json:"change_history"`
	// 	LastModifiedAt time.Time `json:"last_modified_at"`
	// }
	// result := models.ConfigMetadata{}
	// for key, value := range metadata {
	// 	metadataKey := fmt.Sprintf("/metadata/%s/%s", baseKey, key)
	// 	log.Printf("Storing metadata key: %s, value: %s", metadataKey, value)

	// 	err := r.EtcdClient.PutKey(metadataKey, value)
	// 	if err != nil {
	// 		log.Printf("Error storing metadata key %s: %v", metadataKey, err)
	// 		return nil, fmt.Errorf("failed to store metadata field %s: %v", key, err)
	// 	}
	// }

	log.Printf("Successfully stored all config and metadata for %s", baseKey)
	return configData, nil
}

// GetConfig retrieves a configuration from etcd
func (r *ConfigRepository) GetConfig(ctx context.Context, serviceName, environment string) (map[string]interface{}, error) {

	baseKey := fmt.Sprintf("%s/%s", environment, serviceName)
	log.Printf("Getting all config for base key: %s", baseKey)

	// Get all keys under the base key
	key, err := r.EtcdClient.GetKey(baseKey)
	if err != nil {
		log.Printf("Error getting all keys: %v", err)
		return nil, fmt.Errorf("failed to get config keys: %v", err)
	}

	// Create result map
	type ConfigMetadata struct {
		LastModifiedBy string    `json:"last_modified_by"`
		ChangeHistory  []string  `json:"change_history"`
		LastModifiedAt time.Time `json:"last_modified_at"`
	}
	result := map[string]interface{}{}

	// // Process each key-value pair
	// for key, value := range keys {
	// 	// Skip metadata fields
	// 	result[key] = value
	// 	log.Printf("Retrieved key: %s, value: %v", key, value)
	// }

	if err := json.Unmarshal([]byte(key), &result); err != nil {
		return nil, fmt.Errorf("failed to get config keys: %v", err)
	}

	log.Printf("Successfully retrieved all config for %s", baseKey)
	return result, nil
}

// GetConfigValue retrieves a specific config value from etcd
func (r *ConfigRepository) GetConfigValue(ctx context.Context, serviceName, environment, key string) (interface{}, error) {
	// app.InitEtcdDB()
	// defer app.Client.Close()

	// Construct the full key
	baseKey := fmt.Sprintf("%s/%s", environment, serviceName)
	log.Printf("Getting config value for key: %s", key)

	// Get the value
	baseValue, err := r.EtcdClient.GetKey(baseKey)
	if err != nil {
		log.Printf("Error getting key %s: %v", baseKey, err)
		return nil, fmt.Errorf("failed to get config value: %v", err)
	}

	// Try to convert the value to appropriate type
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(baseValue), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal baseKey data: %v", err)
	}

	// if value == "true" || value == "false" {
	// 	result = value == "true"
	// } else if intValue, err := strconv.Atoi(value); err == nil {
	// 	result = intValue
	// } else if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
	// 	result = floatValue
	// } else {
	// 	result = value
	// }
	value, ok := result[key]
	if !ok {
		return nil, fmt.Errorf("failed to get key %v from stored config", key)
	}

	log.Printf("Successfully retrieved value for key %s: %v", key, value)
	return value, nil
}

func (r *ConfigRepository) SetEtcdKey(ctx context.Context, key string, data string, ttl time.Duration) error {
	if ttl > 0 {
		// Create a lease
		leaseResp, err := r.EtcdClient.Client.Grant(ctx, int64(ttl.Seconds()))
		if err != nil {
			log.Printf("Failed to create lease: %v", err)
			return fmt.Errorf("lease creation failed: %w", err)
		}

		_, err = r.EtcdClient.Client.Put(ctx, key, data, clientv3.WithLease(leaseResp.ID))
		if err != nil {
			log.Printf("Error storing key with lease %s: %v", key, err)
			return fmt.Errorf("failed to store key %s: %w", key, err)
		}
	} else {
		_, err := r.EtcdClient.Client.Put(ctx, key, data)
		if err != nil {
			log.Printf("Error storing key %s: %v", key, err)
			return fmt.Errorf("failed to store key %s: %w", key, err)
		}
	}

	return nil
}

func (r *ConfigRepository) GetEtcdKey(ctx context.Context, key string) (string, error) {

	resp, err := r.EtcdClient.Client.Get(ctx, key)
	if err != nil {
		log.Printf("Failed to get key %s: %v", key, err)
		return "", fmt.Errorf("failed to retrieve key %s: %w", key, err)
	}

	if len(resp.Kvs) == 0 {
		return "", fmt.Errorf("key %s not found", key)
	}

	return string(resp.Kvs[0].Value), nil
}
func (r *ConfigRepository) DeleteEtcdKey(ctx context.Context, key string) error {
	_, err := r.EtcdClient.Client.Delete(ctx, key)
	if err != nil {
		log.Printf("Failed to delete key %s: %v", key, err)
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}
	return nil
}

func (r *ConfigRepository) CreateAdmin(ctx context.Context, admin *models.Admin) (*models.Admin, error) {
	// Check if user with the same email already exists
	var existingUser models.Admin
	if err := db.DB.Where("user_name = ?", admin.UserName).First(&existingUser).Error; err == nil {
		return nil, common.ThrowError(fiber.StatusConflict, "ADMIN001") // User already exists
	}

	// Create new user using GORM
	if err := db.DB.Create(admin).Error; err != nil {
		return nil, err
	}

	return admin, nil
}

func (r *ConfigRepository) GetAdminByCredentials(ctx context.Context, username, password string) (*models.Admin, error) {
	var admin models.Admin

	// Find user by username and password
	err := db.DB.Where("user_name = ? AND password = ?", username, password).First(&admin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ThrowError(fiber.StatusUnauthorized, "ADMIN002") // Admin not found
		}
		return nil, err
	}

	return &admin, nil
}
