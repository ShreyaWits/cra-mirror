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
	runMigrations func(*gocql.Session),
	fatal func(format string, v ...any),
) *gocql.Session {
	// Step 1: Try connecting with the keyspace
	cluster := newCluster(host)
	cluster.Port = port
	cluster.ProtoVersion = 4
	cluster.Keyspace = keyspace
	cluster.ConnectTimeout = time.Second * 10
	cluster.Consistency = gocql.Quorum

	session, err := cluster.CreateSession()
	if err != nil {
		fmt.Printf("Error creating Cassandra session: %s\n", err.Error())

		// Step 2: Create keyspace using a session without a keyspace
		tempCluster := newCluster(host)
		tempCluster.Port = port
		tempCluster.ProtoVersion = 4
		tempCluster.ConnectTimeout = time.Second * 10
		tempCluster.Consistency = gocql.Quorum

		tempSession, tempErr := tempCluster.CreateSession()
		if tempErr != nil {
			fatal("Failed to create temp session for keyspace creation: %v", tempErr)
		}
		defer tempSession.Close()

		createKeyspaceCQL := fmt.Sprintf(`
			CREATE KEYSPACE IF NOT EXISTS %s 
			WITH replication = {
				'class': 'SimpleStrategy',
				'replication_factor': '1'
			}`, keyspace)

		if execErr := tempSession.Query(createKeyspaceCQL).Exec(); execErr != nil {
			fatal("Failed to create keyspace '%s': %v", keyspace, execErr)
		}

		// Step 3: Try again to create session with keyspace
		session, err = cluster.CreateSession()
		if err != nil {
			fatal("Failed to create Cassandra session after keyspace creation: %v", err)
		}
	}

	fmt.Println("Cassandra connection established")

	runMigrations(session)
	log.Println("Cassandra migrations completed")

	return session
}

func newCassandraDB(
	host string, port int, keyspace, username, password string,
	newCluster func(string) *gocql.ClusterConfig,
	runMigrations func(*gocql.Session),
) *gocql.Session {
	return newCassandraDBWithFatalf(host, port, keyspace, username, password, newCluster, runMigrations, log.Fatalf)
}

func NewCassandraDB(host string, port int, keyspace, username, password string) *gocql.Session {
	return newCassandraDB(
		host, port, keyspace, username, password,
		func(h string) *gocql.ClusterConfig { return gocql.NewCluster(h) },
		migrations.RunMigrations,
	)
}
