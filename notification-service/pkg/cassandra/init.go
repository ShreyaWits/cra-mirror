package cassandra

import (
	"fmt"
	"log"
	"notification-service/internal/common/migrations"
	"time"

	"github.com/gocql/gocql"
)

func newCassandraDBWithFatalf(
	host string, port int, keyspace, username, password string,
	newCluster func(string) *gocql.ClusterConfig,
	createSession func(*gocql.ClusterConfig) (*gocql.Session, error),
	runMigrations func(*gocql.Session),
	fatal func(format string, v ...any),
) *gocql.Session {
	cluster := newCluster(host)
	cluster.Port = port
	cluster.ProtoVersion = 4
	cluster.Keyspace = keyspace
	cluster.ConnectTimeout = time.Second * 10
	cluster.Consistency = gocql.Quorum

	session, err := createSession(cluster)
	if err != nil {
		fatal("Error creating Cassandra session: %v", err)
	}
	fmt.Println("Cassandra connection established")

	runMigrations(session)
	log.Println("Cassandra migrations completed")

	return session
}

func newCassandraDB(
	host string, port int, keyspace, username, password string,
	newCluster func(string) *gocql.ClusterConfig,
	createSession func(*gocql.ClusterConfig) (*gocql.Session, error),
	runMigrations func(*gocql.Session),
) *gocql.Session {
	return newCassandraDBWithFatalf(host, port, keyspace, username, password, newCluster, createSession, runMigrations, log.Fatalf)
}

func NewCassandraDB(host string, port int, keyspace, username, password string) *gocql.Session {
	return newCassandraDB(
		host, port, keyspace, username, password,
		func(h string) *gocql.ClusterConfig { return gocql.NewCluster(h) },
		func(cluster *gocql.ClusterConfig) (*gocql.Session, error) { return cluster.CreateSession() },
		migrations.RunMigrations,
	)
}
