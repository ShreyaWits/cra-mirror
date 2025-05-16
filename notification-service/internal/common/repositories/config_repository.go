package repositories

import (
	"context"
	"fmt"
	"log"

	"notification-service/internal/common/api/dtos"

	"github.com/gocql/gocql"
)

// Define the minimal session interface we need
type CassandraSession interface {
	Query(stmt string, values ...interface{}) QueryExecutor
	ExecuteBatch(batch *gocql.Batch) error
}

// Update the repository to use the interface
type ConfigRepository struct {
	session CassandraSession
}

type ConfigRepositoryInterface interface {
	SaveConfig(ctx context.Context, configs []dtos.ChannelConfig) error
	GetConfig() ([]dtos.ChannelConfig, error)
	GetConfigByChannel(channel string) (*dtos.ChannelConfig, error)
}

// Add this line at the top of the file to ensure the concrete type implements the interface
var _ ConfigRepositoryInterface = (*ConfigRepository)(nil)

// Update the constructor
func NewConfigRepository(session CassandraSession) (ConfigRepositoryInterface, error) {
	if session == nil {
		return nil, fmt.Errorf("Cassandra session is nil")
	}

	return &ConfigRepository{
		session: session,
	}, nil
}

// SaveConfig saves the configuration to the Cassandra database.
// Added extra nil check to ensure the repository instance is valid before using it
func (r *ConfigRepository) SaveConfig(ctx context.Context, configs []dtos.ChannelConfig) error {
	// Check if the repository is nil to avoid panic
	if r == nil || r.session == nil {
		return fmt.Errorf("repository or session is nil")
	}

	// Start a batch for inserting config data
	batch := gocql.NewBatch(gocql.LoggedBatch)

	// Iterate through the config and prepare the batch insert queries
	for _, cfg := range configs {
		// Create a query for each config item
		query := fmt.Sprintf(`INSERT INTO channel_configs (service, primary_provider, fallback_provider)
			VALUES ('%s', '%s', '%s')`, cfg.Service, cfg.Primary, cfg.Fallback)

		// Add the query to the batch using the Batch.Query() method
		batch.Query(query)
	}

	// Execute the batch of queries
	err := r.session.ExecuteBatch(batch)
	if err != nil {
		fmt.Printf("Error saving config to Cassandra: %v", err)
		return err
	}

	fmt.Printf("Successfully saved %d configs to Cassandra", len(configs))
	return nil
}

// GetConfig fetches all the configurations from the Cassandra database.
func (r *ConfigRepository) GetConfig() ([]dtos.ChannelConfig, error) {
	// Check if the repository is nil to avoid panic
	if r == nil || r.session == nil {
		return nil, fmt.Errorf("repository or session is nil")
	}

	// Prepare a query to fetch all configurations
	query := `SELECT service, primary_provider, fallback_provider FROM channel_configs`

	// Execute the query
	var rows *gocql.Iter
	rows = r.session.Query(query).Iter()

	// Slice to hold the fetched configuration data
	var configs []dtos.ChannelConfig

	// Iterate over the rows and map them to the DTOs
	for {
		var cfg dtos.ChannelConfig
		if !rows.Scan(&cfg.Service, &cfg.Primary, &cfg.Fallback) {
			break
		}
		configs = append(configs, cfg)
	}

	if len(configs) == 0 {
		log.Println("No configurations found in the database")
	}

	return configs, nil
}

// GetConfig fetches all the configurations from the Cassandra database.
func (r *ConfigRepository) GetConfigByChannel(channel string) (*dtos.ChannelConfig, error) {
	if r == nil || r.session == nil {
		return nil, fmt.Errorf("repository or session is nil")
	}

	// Update the query to use channel_configs table instead of notifications
	query := `SELECT service, primary_provider, fallback_provider FROM channel_configs WHERE service = ?`

	var config dtos.ChannelConfig
	err := r.session.Query(query, channel).Scan(&config.Service, &config.Primary, &config.Fallback)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &config, nil
}
