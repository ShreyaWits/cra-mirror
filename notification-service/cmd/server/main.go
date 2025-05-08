package main

import (
	"context"
	"notification-service/internal/app"
	"notification-service/internal/common/api/handlers"
	"notification-service/internal/common/services"
	config "notification-service/internal/configs"
	"notification-service/internal/constant"
	service "notification-service/internal/service/kafka"

	"notification-service/pkg/grpc"
	"notification-service/pkg/kafka"
	"notification-service/pkg/logger"
	"notification-service/pkg/tracer"
	"notification-service/proto"
	"os"
	"os/signal"
	"syscall"
	"time"

	segmentiokafka "github.com/segmentio/kafka-go"
	grpc_package "google.golang.org/grpc"
)

// Interfaces for testability

type GRPCServer interface {
	Start() error
	Stop()
	GetServer() *grpc_package.Server
}

type KafkaConsumer interface {
	Close()
	Consume(topic string, handler interface{})
	Start(ctx context.Context)
}

type Logger interface {
	Println(...interface{})
	Info(...interface{})
	Errorf(string, ...interface{})
	Fatalf(string, ...interface{})
}

// Adapter for GRPCServerInstance

type GRPCServerAdapter struct {
	inner *grpc.GRPCServerInstance
}

func (a *GRPCServerAdapter) Start() error                    { return a.inner.Start() }
func (a *GRPCServerAdapter) Stop()                           { a.inner.Stop() }
func (a *GRPCServerAdapter) GetServer() *grpc_package.Server { return a.inner.GetServer() }

// Adapter for KafkaConsumer

type KafkaConsumerAdapter struct {
	inner *kafka.KafkaConsumer
}

func (a *KafkaConsumerAdapter) Close() { a.inner.Close() }
func (a *KafkaConsumerAdapter) Consume(topic string, handler interface{}) {
	if handler == nil {
		a.inner.Consume(topic, nil)
		return
	}
	if h, ok := handler.(func(msg interface{})); ok {
		a.inner.Consume(topic, func(msg segmentiokafka.Message) { h(msg) })
		return
	}
	if h, ok := handler.(func(msg segmentiokafka.Message)); ok {
		a.inner.Consume(topic, h)
		return
	}
	// fallback: do nothing or panic
}
func (a *KafkaConsumerAdapter) Start(ctx context.Context) { a.inner.Start(ctx) }

func runServer(
	tracerInit func(string) (func(context.Context) error, error),
	getEnv func(string, string) string,
	initDependency func(),
	grpcServerFactory func(string) GRPCServer,
	kafkaConsumerFactory func([]string, string, []string) (KafkaConsumer, error),
	loggerLog Logger,
	signalNotify func(chan<- os.Signal, ...os.Signal),
	sleep func(time.Duration),
) error {

	shutdown, err := tracerInit("notification-service")
	if err != nil {
		loggerLog.Fatalf("Cannot initialize tracer: %v", err)
		return err
	}
	defer func() {
		if shutdown != nil {
			_ = shutdown(context.Background())
		}
	}()
	initDependency()
	grpcPort := getEnv("GRPC_PORT", ":50051")
	loggerLog.Println("GRPC_PORT", grpcPort)

	grpcInstance := grpcServerFactory(grpcPort)

	KAFKA_SERVER_URL := getEnv("KAFKA_SERVER_URL", "kafka:29092")
	KAFKA_GROUP_ID := getEnv("KAFKA_GROUP_ID", "notification-service-group")
	loggerLog.Println("KAFKA_SERVER_URL", KAFKA_SERVER_URL)
	loggerLog.Println("KAFKA_GROUP_ID", KAFKA_GROUP_ID)

	consumer, err := kafkaConsumerFactory(
		[]string{KAFKA_SERVER_URL},
		KAFKA_GROUP_ID,
		[]string{string(constant.NOTIFICATION_TOPIC)},
	)

	if err != nil {
		loggerLog.Fatalf("Failed to init Kafka consumer: %v", err)
		return err
	}
	defer consumer.Close()

	app := app.Init()

	notificationService := services.NewNotificationService()

	// Set up your gRPC service handler
	notificationHandler := handlers.NewGRPCServer(notificationService)

	// Register the handler with the gRPC server
	proto.RegisterNotificationServiceServer(grpcInstance.GetServer(), notificationHandler)

	handlers := service.NewQueueHandlers(app.GetRedisService(), app.GetConfigRepo(), app.GetTemporalClient(), app.GetNotificationRepo(), app.GetKafka())

	// Create a context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	stopChan := make(chan os.Signal, 1)
	signalNotify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// Start gRPC server in a separate goroutine
	go func() {
		loggerLog.Info("gRPC server started on port", grpcPort)
		if err := grpcInstance.Start(); err != nil {
			loggerLog.Errorf("Failed to start gRPC server: %v", err)
			cancel()
		}
		defer grpcInstance.Stop()
	}()

	// Start Kafka consumer in a separate goroutine
	go func() {
		// Start consuming messages from Kafka
		consumer.Consume(string(constant.NOTIFICATION_TOPIC), handlers.HandleSendNotification)
		consumer.Start(ctx)
	}()

	// Wait for a termination signal (SIGINT or SIGTERM)
	select {
	case <-stopChan:
		loggerLog.Info("Received shutdown signal, shutting down...")
		sleep(2 * time.Second)
		cancel()
	}

	loggerLog.Info("Shutdown completed.")
	return nil
}

func grpcServerAdapter(port string) GRPCServer {
	return &GRPCServerAdapter{inner: grpc.NewGRPCServer(port)}
}

func kafkaConsumerAdapter(brokers []string, groupID string, topics []string) (KafkaConsumer, error) {
	logger.Log.Println("Initializing Kafka consumer...", brokers, groupID, topics)
	kc, err := kafka.InitKafkaConsumer(brokers, groupID, topics)
	if err != nil {
		return nil, err
	}
	return &KafkaConsumerAdapter{inner: kc}, nil
}

func main() {
	LOKI_URL := config.GetEnv("LOKI_URL", "http://localhost:5000")
	logger.InitLogger(LOKI_URL)
	runServer(
		tracer.InitTracer,
		config.GetEnv,
		app.InitDependency,
		grpcServerAdapter,
		kafkaConsumerAdapter,
		logger.Log,
		signal.Notify,
		time.Sleep,
	)
}
