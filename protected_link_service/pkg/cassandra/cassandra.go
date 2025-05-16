package cassandra

import (
	"fmt"
	"log"
	"time"

	configEnv "protected_link/internal/configs"

	"github.com/gocql/gocql"
)

// --- INTERFACES ---

type SessionInterface interface {
	Query(string, ...interface{}) QueryInterface
	Close()
}

type QueryInterface interface {
	Consistency(gocql.Consistency) QueryInterface
	Scan(...interface{}) error
	Exec() error
}

// --- WRAPPER TYPES FOR REAL IMPLEMENTATION ---

type RealSession struct {
	*gocql.Session
}

func (r *RealSession) Query(stmt string, values ...interface{}) QueryInterface {
	return &RealQuery{r.Session.Query(stmt, values...)}
}

type RealQuery struct {
	*gocql.Query
}

func (r *RealQuery) Consistency(c gocql.Consistency) QueryInterface {
	r.Query = r.Query.Consistency(c)
	return r
}

func (r *RealQuery) Scan(dest ...interface{}) error {
	return r.Query.Scan(dest...)
}

func (r *RealQuery) Exec() error {
	return r.Query.Exec()
}

// --- CASSANDRA CONFIG STRUCT ---

type CassandraConfig struct {
	Session SessionInterface
}

// --- FACTORY METHODS ---

func NewCassandraConfig(cfg *configEnv.Config) (*CassandraConfig, error) {
	maxRetries := 3
	retryDelay := 2 * time.Second

	cluster := gocql.NewCluster(cfg.CASSANDRA_HOST)
	cluster.Keyspace = cfg.CASSANDRA_KEYSPACE
	cluster.Consistency = gocql.Quorum
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: cfg.CASSANDRA_USERNAME,
		Password: cfg.CASSANDRA_PASSWORD,
	}
	cluster.Timeout = 5 * time.Second

	var session *gocql.Session
	var err error

	for i := 1; i <= maxRetries; i++ {
		log.Printf("🔄 Attempting to connect to Cassandra (Attempt %d/%d)...", i, maxRetries)
		session, err = cluster.CreateSession()
		if err == nil {
			log.Println("✅ Successfully connected to Cassandra")
			return &CassandraConfig{
				Session: &RealSession{session},
			}, nil
		}
		log.Printf("❌ Cassandra connection failed (Attempt %d/%d): %v", i, maxRetries, err)
		time.Sleep(retryDelay)
	}

	return nil, fmt.Errorf("failed to connect to Cassandra after %d retries: %w", maxRetries, err)
}

func NewMockCassandraConfig(session SessionInterface) *CassandraConfig {
	return &CassandraConfig{
		Session: session,
	}
}

func (c *CassandraConfig) Close() {
	if c.Session != nil {
		c.Session.Close()
		log.Println("✅ Cassandra connection closed")
	}
}
