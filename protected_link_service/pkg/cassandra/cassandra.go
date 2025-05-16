package cassandra

import (
	"fmt"
	"log"
	configEnv "protected_link/internal/configs"
	cassandra "protected_link/migrations"
	"time"

	"github.com/gocql/gocql"
)

type CassandraConfig struct {
	Session *gocql.Session
}

func NewCassandraConfig(cfg *configEnv.Config) (*CassandraConfig, error) {
	maxRetries := 3
	retryDelay := 2 * time.Second
	var session *gocql.Session
	var err error

	cassandraAddress := fmt.Sprintf("%s:%s", cfg.CASSANDRA_HOST, cfg.CASSANDRA_PORT)

	log.Println("🔧 Cassandra Config:")
	log.Println("   Host:", cassandraAddress)
	log.Println("   Keyspace:", cfg.CASSANDRA_KEYSPACE)
	log.Println("   Username:", cfg.CASSANDRA_USERNAME)

	cluster := gocql.NewCluster(cassandraAddress)
	cluster.Keyspace = cfg.CASSANDRA_KEYSPACE
	cluster.Consistency = gocql.Quorum
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: cfg.CASSANDRA_USERNAME,
		Password: cfg.CASSANDRA_PASSWORD,
	}
	cluster.ProtoVersion = 4 // Manually setting protocol version can help avoid negotiation issues
	cluster.Timeout = 10 * time.Second

	for i := 1; i <= maxRetries; i++ {
		log.Printf("🔄 Attempting to connect to Cassandra (Attempt %d/%d)...", i, maxRetries)
		session, err = cluster.CreateSession()
		if err == nil {
			log.Println("✅ Successfully connected to Cassandra")

			// Apply migrations after successful connection
			if err := cassandra.ApplyMigrations(session, "./migrations"); err != nil {
				log.Printf("❌ Failed to apply migrations: %v", err)
				return nil, err
			}

			return &CassandraConfig{
				Session: session,
			}, nil
		}

		log.Printf("❌ Cassandra connection failed (Attempt %d/%d): %v", i, maxRetries, err)
		if i < maxRetries {
			log.Printf("⏳ Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	return nil, fmt.Errorf("failed to connect to Cassandra after %d retries: %w", maxRetries, err)
}
func (c *CassandraConfig) Close() {
	if c.Session != nil {
		c.Session.Close()
		log.Println("✅ Cassandra connection closed")
	}
}
