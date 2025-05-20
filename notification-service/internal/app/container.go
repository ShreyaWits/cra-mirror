package app

import (
	"log"
	"notification-service/internal/common/repositories"
	config "notification-service/internal/configs"
	"notification-service/pkg/cassandra"
	"notification-service/pkg/di"
	"notification-service/pkg/kafka"
	"notification-service/pkg/redis"
	"notification-service/pkg/temporal"
	"strconv"

	"github.com/gocql/gocql"
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

var (
	initContainer             = di.InitContainer
	newCassandraDB            = cassandra.NewCassandraDB
	newRedisService           = repositories.InitRedisRepo
	initRedisService          = redis.NewRedisService
	newConfigRepository       = repositories.NewConfigRepository
	newNotificationRepository = repositories.NewNotificationRepository
	initKafkaPublisher        = kafka.InitKafkaPublisher
	initTemporal              = temporal.InitTemporal
)

var container *di.Container

func InitDependency() {
	//init dig
	if container == nil {
		appContainer := initContainer()
		container = appContainer
	}

	CASSANDRA_HOST := config.GetEnv("CASSANDRA_HOST", "localhost")
	CASSANDRA_KEYSPACE := config.GetEnv("CASSANDRA_KEYSPACE", "notifications")
	CASSANDRA_USERNAME := config.GetEnv("CASSANDRA_USERNAME", "cassandra")
	CASSANDRA_PASSWORD := config.GetEnv("CASSANDRA_PASSWORD", "cassandra")
	CASSANDRA_PORT, err := strconv.Atoi(config.GetEnv("CASSANDRA_PORT", "9042"))
	REDIS_HOST := config.GetEnv("REDIS_HOST", "localhost")
	REDIS_PORT := config.GetEnv("REDIS_PORT", "6379")
	TEMPORAL_SERVER_URL := config.GetEnv("TEMPORAL_SERVER_URL", "localhost:7233")
	TEMPORAL_QUEUE := config.GetEnv("TEMPORAL_QUEUE", "notification-service-queue")
	KAFKA_BROKER := config.GetEnv("KAFKA_BROKERS", "kafka:29092")
	if err != nil {
		panic(err)
	}

	log.Println("cassandra host", CASSANDRA_HOST)
	log.Println("cassandra port", CASSANDRA_PORT)
	log.Println("cassandra keyspace", CASSANDRA_KEYSPACE)
	log.Println("cassandra username", CASSANDRA_USERNAME)
	log.Println("cassandra password", CASSANDRA_PASSWORD)
	log.Println("kafka broker", KAFKA_BROKER)

	session := newCassandraDB(CASSANDRA_HOST, CASSANDRA_PORT, CASSANDRA_KEYSPACE, CASSANDRA_USERNAME, CASSANDRA_PASSWORD)

	wrappedSession := &cassandraSessionWrapper{session}

	configRepo, err := newConfigRepository(wrappedSession)
	if err != nil {
		panic(err)
	}

	redisRepo, err := initRedisService(REDIS_HOST, REDIS_PORT, "", "")
	if err != nil {
		panic(err)
	}

	notificationRedisRepo := newRedisService(redisRepo)

	notificationRepo := newNotificationRepository(wrappedSession)
	kafkaProducer := initKafkaPublisher(KAFKA_BROKER)

	temporalClient, err := initTemporal(TEMPORAL_SERVER_URL, TEMPORAL_QUEUE, "notification-service", notificationRepo, kafkaProducer)
	if err != nil {
		panic(err)
	}

	container.RegisterService(func() *gocql.Session {
		return session
	})
	container.RegisterService(func() repositories.RedisRepositoryInterface {
		return notificationRedisRepo
	})
	container.RegisterService(func() repositories.ConfigRepositoryInterface {
		return configRepo
	})
	container.RegisterService(func() *temporal.TemporalClient {
		return temporalClient
	})
	container.RegisterService(func() repositories.NotificationRepositoryInterface {
		return notificationRepo
	})
	container.RegisterService(func() *kafka.KafkaPublisher {
		return kafkaProducer
	})

}

type DiService struct {
	Container *di.Container
}

func Init() *DiService {
	return &DiService{Container: container}
}

func (c *DiService) GetCassandraSession() *gocql.Session {
	var dbSession *gocql.Session
	c.Container.InvokeService(func(session *gocql.Session) {
		dbSession = session
	})
	return dbSession
}

func (c *DiService) GetRedisService() repositories.RedisRepositoryInterface {
	var redisService repositories.RedisRepositoryInterface
	c.Container.InvokeService(func(service repositories.RedisRepositoryInterface) {
		redisService = service
	})
	return redisService
}

func (c *DiService) GetConfigRepo() repositories.ConfigRepositoryInterface {
	var configRepo repositories.ConfigRepositoryInterface
	c.Container.InvokeService(func(repo repositories.ConfigRepositoryInterface) {
		configRepo = repo
	})
	return configRepo
}
func (c *DiService) GetNotificationRepo() repositories.NotificationRepositoryInterface {
	var notificationRepo repositories.NotificationRepositoryInterface
	c.Container.InvokeService(func(repo repositories.NotificationRepositoryInterface) {
		notificationRepo = repo
	})
	return notificationRepo
}

func (c *DiService) GetTemporalClient() *temporal.TemporalClient {
	var temporalClient *temporal.TemporalClient
	c.Container.InvokeService(func(client *temporal.TemporalClient) {
		temporalClient = client
	})
	return temporalClient
}
func (c *DiService) GetKafka() *kafka.KafkaPublisher {
	var kafkaClient *kafka.KafkaPublisher
	c.Container.InvokeService(func(client *kafka.KafkaPublisher) {
		kafkaClient = client
	})
	return kafkaClient
}

func GetContainer() *di.Container {
	return container
}

// SetContainer allows tests to set the global DI container
func SetContainer(c *di.Container) {
	container = c
}
