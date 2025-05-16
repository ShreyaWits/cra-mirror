package main

import (
	"log"
	"notification-service/internal/common/repositories"
	config "notification-service/internal/configs"
	"notification-service/pkg/cassandra"
	"notification-service/pkg/kafka"
	"notification-service/pkg/logger"
	"notification-service/pkg/temporal"
	"os"
	"strconv"

	"github.com/gocql/gocql"
	"go.temporal.io/sdk/worker"
)

type cassandraSessionWrapper struct {
	*gocql.Session
}

func (w *cassandraSessionWrapper) Query(stmt string, values ...interface{}) repositories.QueryExecutor {
	// Delegate to the underlying *gocql.Session instance
	query := w.Session.Query(stmt, values...)
	return &queryExecutorWrapper{query}
}

type queryExecutorWrapper struct {
	*gocql.Query
}

func (w *queryExecutorWrapper) Exec() error {
	// Delegate to the underlying *gocql.Query instance
	return w.Query.Exec()
}

type TemporalWorker interface {
	RegisterActivities([]any)
	RegisterWorkflows([]any)
	Run()
}

type KafkaProducer interface{}
type CassandraSession interface{}

func runWorkerWithDeps(
	temporalInit func(string, worker.Options) (TemporalWorker, error),
	cassandraInit func(string, int, string, string, string) CassandraSession,
	kafkaInit func(string) KafkaProducer,
) error {
	TemporalUrl := os.Getenv("TEMPORAL_SERVER_URL")
	LOKI_URL := config.GetEnv("LOKI_URL", "http://localhost:5000")
	logger.InitLogger(LOKI_URL)

	temporalServer, err := temporalInit(TemporalUrl, worker.Options{})
	if err != nil {
		return err
	}

	TIMEOUT := 1000
	RETRY_COUNT := 3

	CASSANDRA_HOST := config.GetEnv("CASSANDRA_HOST", "localhost")
	CASSANDRA_KEYSPACE := config.GetEnv("CASSANDRA_KEYSPACE", "notifications")
	CASSANDRA_USERNAME := config.GetEnv("CASSANDRA_USERNAME", "cassandra")
	CASSANDRA_PASSWORD := config.GetEnv("CASSANDRA_PASSWORD", "cassandra")
	CASSANDRA_PORT, err := strconv.Atoi(config.GetEnv("CASSANDRA_PORT", "9042"))
	KAFKA_BROKER := config.GetEnv("KAFKA_BROKERS", "kafka:29092")

	if err != nil {
		return err
	}

	kafkaProducer := kafkaInit(KAFKA_BROKER)
	dbSession := cassandraInit(CASSANDRA_HOST, CASSANDRA_PORT, CASSANDRA_KEYSPACE, CASSANDRA_USERNAME, CASSANDRA_PASSWORD)
	wrappedSession := &cassandraSessionWrapper{dbSession.(*gocql.Session)}
	notificationRepo := repositories.NewNotificationRepository(wrappedSession)

	temporalWorkflow := temporal.NewTemporalWorkflow(&TIMEOUT, &RETRY_COUNT, notificationRepo, kafkaProducer.(*kafka.KafkaPublisher))

	temporalServer.RegisterActivities([]any{
		temporalWorkflow.InitSMTPActivity,
		temporalWorkflow.InitSendGridActivity,
		temporalWorkflow.InitTwilioSMSActivity,
		temporalWorkflow.InitTwilioWhatsappActivity,
		temporalWorkflow.InitFirebasePushActivity})

	temporalServer.RegisterWorkflows([]any{
		temporalWorkflow.ExecuteEmailWorkflow,
		temporalWorkflow.ExecuteSMSWorkflow,
		temporalWorkflow.ExecuteWhatsappWorkflow,
		temporalWorkflow.ExecutePushNotificationWorkflow})

	temporalServer.Run()
	return nil
}

func runWorker() error {
	return runWorkerWithDeps(
		func(url string, opts worker.Options) (TemporalWorker, error) {
			return temporal.InitTemporalWorker(url, opts)
		},
		func(host string, port int, keyspace, username, password string) CassandraSession {
			return cassandra.NewCassandraDB(host, port, keyspace, username, password)
		},
		func(broker string) KafkaProducer {
			return kafka.InitKafkaPublisher(broker)
		},
	)
}

func main() {
	if err := runWorker(); err != nil {
		log.Fatal(err)
	}
}
