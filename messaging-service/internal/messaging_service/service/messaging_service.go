package service

import (
	"context"
	"time"

	"messaging-service/pkg/kafka"
	pb "messaging-service/protos/messaging_service"
)

type MessagingServer struct {
	pb.UnimplementedMessagingServiceServer
}

func (s *MessagingServer) PublishMessage(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResponse, error) {
	headers := make([]kafka.Header, 0, len(req.Headers))
	for k, v := range req.Headers {
		headers = append(headers, kafka.Header{
			Key:   k,
			Value: []byte(v),
		})
	}

	// Generate timestamp for message ID
	timestamp := time.Now()

	cfg := kafka.KafkaConfig{
		Brokers:  req.Brokers, // Optional: make this dynamic
		Topic:    req.Topic,
		GroupID:  req.GroupId,
		Mode:     kafka.CompetingConsumer,
		MinBytes: int(req.MinBytes),
		MaxBytes: int(req.MaxBytes),
	}

	producer := kafka.NewDLQProducer(cfg)

	// Call DLQProducer.SendToDLQ (assumes all messages go to DLQ for now)
	err := producer.SendToDLQ(ctx, "", req.Value, headers)
	if err != nil {
		return &pb.PublishResponse{
			Status:    "failed",
			MessageId: "",
		}, err
	}

	return &pb.PublishResponse{
		Status:    "success",
		MessageId: timestamp.Format(time.RFC3339Nano),
	}, nil
}

func (s *MessagingServer) SubscribeStream(req *pb.SubscribeRequest, stream pb.MessagingService_SubscribeStreamServer) error {
	const defaultMinBytes = 10 * 1024
	const defaultMaxBytes = 10 * 1024 * 1024

	cfg := kafka.KafkaConfig{
		Brokers:  req.Brokers, // Optional: make this dynamic
		Topic:    req.Topic,
		GroupID:  req.GroupId,
		Mode:     kafka.CompetingConsumer,
		MinBytes: int(req.MinBytes),
		MaxBytes: int(req.MaxBytes),
	}

	if cfg.MinBytes <= 0 {
		cfg.MinBytes = defaultMinBytes
	}
	if cfg.MaxBytes <= 0 {
		cfg.MaxBytes = defaultMaxBytes
	}

	// Initialize the consumer
	consumer := kafka.NewConsumer(cfg, func(message []byte) error {
		// Here you can define how to handle the incoming message
		// For example, you can send it to the gRPC stream
		headers := make(map[string]string)
		// Assuming you have a way to extract headers from the message
		// Send the message to the stream
		return stream.Send(&pb.KafkaMessage{
			Key:       "", // Set the key if available
			Value:     message,
			Headers:   headers,
			Timestamp: time.Now().UnixMilli(), // Set the current timestamp
		})
	})

	// Start the consumer
	go func() {
		consumer.Start(stream.Context()) // No need to check for a return value
		// Handle any additional logic if needed
	}()

	// Keep the stream open
	<-stream.Context().Done()
	return nil
}
