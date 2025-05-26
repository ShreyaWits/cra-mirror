package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	common "nps-config-service/internal/common/errors"
	"nps-config-service/internal/constants"
	"nps-config-service/internal/modules/config-manager/models"
	etcdDB "nps-config-service/pkg/etcd"
	"nps-config-service/pkg/observability"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/google/uuid"
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

func (r *ConfigRepository) recordMetrics(ctx context.Context, operation string, start time.Time, err error) {
	if r.ObservabilityStack == nil || r.ObservabilityStack.MetricsService == nil {
		return
	}

	// Record latency
	latency := float64(time.Since(start).Milliseconds())

	r.ObservabilityStack.MetricsService.RecordHistogram(ctx, constants.ConfigRepoOperationLatencyMetric, latency, map[string]string{
		"operation": operation,
		"service":   "config_repository",
	})

	// Record operation count
	r.ObservabilityStack.MetricsService.IncrementCounter(ctx, constants.ConfigRepoOperationCountMetric, 1, map[string]string{
		"operation": operation,
		"service":   "config_repository",
	})

	// Record error if any
	if err != nil {
		r.ObservabilityStack.MetricsService.IncrementCounter(ctx, constants.ConfigRepoErrorCountMetric, 1, map[string]string{
			"operation": operation,
			"service":   "config_repository",
			"error":     err.Error(),
		})
	}
}

func (r *ConfigRepository) StoreConfig(ctx context.Context, serviceName, environment string, configData map[string]interface{}) (interface{}, error) {
	start := time.Now()
	ctx, span := r.ObservabilityStack.TracerService.Start(ctx, "ConfigRepository.StoreConfig")
	defer span.End()
	defer r.recordMetrics(ctx, "store_config", start, nil)

	r.ObservabilityStack.Logger.InfoContext(ctx, "Storing config",
		"environment", environment,
		"service", serviceName)

	baseKey := fmt.Sprintf("%s/%s", environment, serviceName)
	changeHistory := []string{}

	b, err := json.Marshal(configData)
	if err != nil {
		r.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to marshal config data",
			"error", err,
			"environment", environment,
			"service", serviceName)
		return nil, fmt.Errorf("failed to marshal config body: %v", err)
	}

	if err := r.EtcdClient.PutKey(baseKey, string(b)); err != nil {
		r.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to store config",
			"error", err,
			"key", baseKey,
			"environment", environment,
			"service", serviceName)
		return nil, fmt.Errorf("failed to store config field %s: %v", baseKey, err)
	}

	r.ObservabilityStack.Logger.InfoContext(ctx, "Config stored successfully",
		"environment", environment,
		"service", serviceName,
		"changes", changeHistory)
	return configData, nil
}

// GetConfig retrieves a configuration from etcd
func (r *ConfigRepository) GetConfig(ctx context.Context, serviceName, environment string) (map[string]interface{}, error) {
	start := time.Now()
	ctx, span := r.ObservabilityStack.TracerService.Start(ctx, "ConfigRepository.GetConfig")
	defer span.End()
	defer func() {
		r.recordMetrics(ctx, "get_config", start, nil)
	}()

	r.ObservabilityStack.Logger.InfoContext(ctx, "Getting config",
		"environment", environment,
		"service", serviceName)

	baseKey := fmt.Sprintf("%s/%s", environment, serviceName)
	key, err := r.EtcdClient.GetKey(baseKey)
	if err != nil {
		r.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to get config",
			"error", err,
			"key", baseKey,
			"environment", environment,
			"service", serviceName)
		return nil, fmt.Errorf("failed to get config keys: %v", err)
	}

	result := map[string]interface{}{}
	if err := json.Unmarshal([]byte(key), &result); err != nil {
		r.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to unmarshal config data",
			"error", err,
			"environment", environment,
			"service", serviceName)
		return nil, fmt.Errorf("failed to get config keys: %v", err)
	}

	r.ObservabilityStack.Logger.InfoContext(ctx, "Config retrieved successfully",
		"environment", environment,
		"service", serviceName)
	return result, nil
}

// GetConfigValue retrieves a specific config value from etcd
func (r *ConfigRepository) GetConfigValue(ctx context.Context, serviceName, environment, key string) (interface{}, error) {
	start := time.Now()
	ctx, span := r.ObservabilityStack.TracerService.Start(ctx, "ConfigRepository.GetConfigValue")
	defer span.End()
	defer func() {
		r.recordMetrics(ctx, "get_config_value", start, nil)
	}()

	r.ObservabilityStack.Logger.InfoContext(ctx, "Getting config value",
		"environment", environment,
		"service", serviceName,
		"key", key)

	baseKey := fmt.Sprintf("%s/%s", environment, serviceName)
	baseValue, err := r.EtcdClient.GetKey(baseKey)
	if err != nil {
		r.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to get config value",
			"error", err,
			"base_key", baseKey,
			"environment", environment,
			"service", serviceName)
		return nil, fmt.Errorf("failed to get config value: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(baseValue), &result); err != nil {
		r.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to unmarshal config data",
			"error", err,
			"environment", environment,
			"service", serviceName)
		return nil, fmt.Errorf("failed to unmarshal baseKey data: %v", err)
	}

	value, ok := result[key]
	if !ok {
		r.ObservabilityStack.Logger.WarnContext(ctx, "Config key not found",
			"key", key,
			"environment", environment,
			"service", serviceName)
		return nil, fmt.Errorf("failed to get key %v from stored config", key)
	}

	r.ObservabilityStack.Logger.InfoContext(ctx, "Config value retrieved successfully",
		"environment", environment,
		"service", serviceName,
		"key", key)
	return value, nil
}

func (r *ConfigRepository) SetEtcdKey(ctx context.Context, key string, data string, ttl time.Duration) error {
	start := time.Now()
	ctx, span := r.ObservabilityStack.TracerService.Start(ctx, "ConfigRepository.SetEtcdKey")
	defer span.End()
	defer func() {
		r.recordMetrics(ctx, "set_etcd_key", start, nil)
	}()

	r.ObservabilityStack.Logger.InfoContext(ctx, "Setting etcd key",
		"key", key,
		"ttl", ttl)

	if ttl > 0 {
		leaseResp, err := r.EtcdClient.Client.Grant(ctx, int64(ttl.Seconds()))
		if err != nil {
			r.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to create lease",
				"error", err,
				"key", key,
				"ttl", ttl)
			return fmt.Errorf("lease creation failed: %w", err)
		}

		_, err = r.EtcdClient.Client.Put(ctx, key, data, clientv3.WithLease(leaseResp.ID))
		if err != nil {
			r.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to store key with lease",
				"error", err,
				"key", key)
			return fmt.Errorf("failed to store key %s: %w", key, err)
		}
	} else {
		_, err := r.EtcdClient.Client.Put(ctx, key, data)
		if err != nil {
			r.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to store key",
				"error", err,
				"key", key)
			return fmt.Errorf("failed to store key %s: %w", key, err)
		}
	}

	r.ObservabilityStack.Logger.InfoContext(ctx, "Etcd key set successfully",
		"key", key)
	return nil
}

func (r *ConfigRepository) GetEtcdKey(ctx context.Context, key string) (string, error) {
	start := time.Now()
	ctx, span := r.ObservabilityStack.TracerService.Start(ctx, "ConfigRepository.GetEtcdKey")
	defer span.End()
	defer func() {
		r.recordMetrics(ctx, "get_etcd_key", start, nil)
	}()

	r.ObservabilityStack.Logger.InfoContext(ctx, "Getting etcd key",
		"key", key)

	resp, err := r.EtcdClient.Client.Get(ctx, key)
	if err != nil {
		r.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to get key",
			"error", err,
			"key", key)
		return "", fmt.Errorf("failed to retrieve key %s: %w", key, err)
	}

	if len(resp.Kvs) == 0 {
		r.ObservabilityStack.Logger.WarnContext(ctx, "Key not found",
			"key", key)
		return "", fmt.Errorf("key %s not found", key)
	}

	r.ObservabilityStack.Logger.InfoContext(ctx, "Etcd key retrieved successfully",
		"key", key)
	return string(resp.Kvs[0].Value), nil
}

func (r *ConfigRepository) DeleteEtcdKey(ctx context.Context, key string) error {
	start := time.Now()
	ctx, span := r.ObservabilityStack.TracerService.Start(ctx, "ConfigRepository.DeleteEtcdKey")
	defer span.End()
	defer func() {
		r.recordMetrics(ctx, "delete_etcd_key", start, nil)
	}()

	r.ObservabilityStack.Logger.InfoContext(ctx, "Deleting etcd key",
		"key", key)

	_, err := r.EtcdClient.Client.Delete(ctx, key)
	if err != nil {
		r.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to delete key",
			"error", err,
			"key", key)
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}

	r.ObservabilityStack.Logger.InfoContext(ctx, "Etcd key deleted successfully",
		"key", key)
	return nil
}

func (r *ConfigRepository) CreateAdmin(ctx context.Context, admin *models.Admin) (*models.Admin, error) {
	start := time.Now()
	ctx, span := r.ObservabilityStack.TracerService.Start(ctx, "ConfigRepository.CreateAdmin")
	defer span.End()
	defer func() {
		r.recordMetrics(ctx, "create_admin", start, nil)
	}()

	r.ObservabilityStack.Logger.InfoContext(ctx, "Creating admin user",
		"username", admin.UserName)

	// Generate UUID if not set
	if admin.ID == uuid.Nil {
		admin.ID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	admin.CreatedAt = now
	admin.UpdatedAt = now

	// Check if admin already exists
	key := fmt.Sprintf("/admins/%s", admin.UserName)
	existingData, err := r.EtcdClient.GetKey(key)
	if err == nil && existingData != "" {
		r.ObservabilityStack.Logger.WarnContext(ctx, "Admin user already exists",
			"username", admin.UserName)
		return nil, common.ThrowError(fiber.StatusConflict, "ADMIN001")
	}

	// Marshal admin data to JSON
	adminData, err := json.Marshal(admin)
	if err != nil {
		r.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to marshal admin data",
			"error", err,
			"username", admin.UserName)
		return nil, err
	}

	// Store in etcd
	err = r.EtcdClient.PutKey(key, string(adminData))
	if err != nil {
		r.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to store admin in etcd",
			"error", err,
			"username", admin.UserName)
		return nil, err
	}

	r.ObservabilityStack.Logger.InfoContext(ctx, "Admin user created successfully",
		"username", admin.UserName)
	return admin, nil
}

func (r *ConfigRepository) GetAdminByCredentials(ctx context.Context, username, password string) (*models.Admin, error) {
	start := time.Now()
	ctx, span := r.ObservabilityStack.TracerService.Start(ctx, "ConfigRepository.GetAdminByCredentials")
	defer span.End()
	defer func() {
		r.recordMetrics(ctx, "get_admin_by_credentials", start, nil)
	}()

	r.ObservabilityStack.Logger.InfoContext(ctx, "Getting admin by credentials",
		"username", username)

	// Get admin data from etcd
	key := fmt.Sprintf("/admins/%s", username)
	adminData, err := r.EtcdClient.GetKey(key)
	if err != nil {
		r.ObservabilityStack.Logger.WarnContext(ctx, "Admin not found",
			"username", username)
		return nil, common.ThrowError(fiber.StatusUnauthorized, "ADMIN002")
	}

	// Unmarshal admin data
	var admin models.Admin
	if err := json.Unmarshal([]byte(adminData), &admin); err != nil {
		r.ObservabilityStack.Logger.ErrorContext(ctx, "Failed to unmarshal admin data",
			"error", err,
			"username", username)
		return nil, err
	}

	// Verify password
	if admin.Password != password {
		r.ObservabilityStack.Logger.WarnContext(ctx, "Invalid password",
			"username", username)
		return nil, common.ThrowError(fiber.StatusUnauthorized, "ADMIN002")
	}

	r.ObservabilityStack.Logger.InfoContext(ctx, "Admin retrieved successfully",
		"username", username)
	return &admin, nil
}
