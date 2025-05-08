package handler

import (
	"context"
	pb "cra-protos/messaging_service"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"messaging_service/pkg/kafka"
	"messaging_service/pkg/logger"
)

type MessagingHandler struct {
	pb.UnimplementedMessagingServiceServer
}

func (s *MessagingHandler) PublishMessage(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResponse, error) {
	header := map[string]string{
		"Content-Type": "application/json",
	}

	headers := make([]kafka.Header, 0, len(header))

	for k, v := range header {
		headers = append(headers, kafka.Header{
			Key:   k,
			Value: []byte(v),
		})
	}

	cfg := kafka.KafkaConfig{
		Brokers:  []string{"kafka:9092"}, // Optional: make this dynamic
		Topic:    req.Topic,
		GroupID:  req.GroupId,
		MinBytes: int(req.MinBytes),
		MaxBytes: int(req.MaxBytes),
	}

	producer := kafka.NewDLQProducer(cfg)

	b, err := MapToBytes(req.Value)
	if err != nil {
		log.Fatalf("serialization failed: %v", err)
	}

	// Call DLQProducer.SendToDLQ (assumes all messages go to DLQ for now)
	err = producer.SendToDLQ(ctx, "", b, headers)
	if err != nil {
		return &pb.PublishResponse{
			Status: "failed",
		}, err
	}

	return &pb.PublishResponse{
		Status: "success",
	}, nil
}

func (s *MessagingHandler) SubscribeStream(req *pb.SubscribeRequest, stream pb.MessagingService_SubscribeStreamServer) error {
	const defaultMinBytes = 1
	const defaultMaxBytes = 1048576

	cfg := kafka.KafkaConfig{
		Brokers:  []string{"kafka:9092"}, // Optional: make this dynamic
		Topic:    req.Topic,
		GroupID:  req.GroupId,
		MinBytes: int(req.MinBytes),
		MaxBytes: int(req.MaxBytes),
	}
	logger.LogInfo("this is config data", fmt.Sprintf("Data : %v", cfg))

	if cfg.MinBytes <= 0 {
		cfg.MinBytes = defaultMinBytes
	}
	if cfg.MaxBytes <= 0 {
		cfg.MaxBytes = defaultMaxBytes
	}

	// Initialize the consumer
	consumer := kafka.NewConsumer(cfg, func(message []byte) error {
		// Assuming you have a way to extract headers from the message
		// Send the message to the stream

		msg, err := BytesToMap(message)
		if err != nil {
			log.Printf("deserialization failed: %v %s", err, message)
		}

		return stream.Send(&pb.KafkaMessage{
			Value:     msg,
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

func MapToBytes(m map[string]string) ([]byte, error) {
	return json.Marshal(m)
}
func BytesToMap(b []byte) (map[string]string, error) {
	var m map[string]string
	err := json.Unmarshal(b, &m)
	return m, err
}
