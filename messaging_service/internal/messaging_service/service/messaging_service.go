package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pb "cra-protos/messaging_service"
	errors "messaging_service/pkg/errors"
	kafkapkg "messaging_service/pkg/kafka"
	"messaging_service/pkg/logger"

	kafka "github.com/segmentio/kafka-go"
)

type MessagingService interface {
	PublishMessage(ctx context.Context, cfg kafkapkg.KafkaConfig, req *pb.PublishRequest) *errors.CustomError
	ConsumeMessage(stream pb.MessagingService_SubscribeV1Server, cfg kafkapkg.KafkaConfig) *errors.CustomError
	CreateTopic(ctx context.Context, req *pb.CreateTopicRequest, cfg kafkapkg.KafkaConfig) *errors.CustomError
}

type MessagingServiceImpl struct {
	producer *kafkapkg.Producer
	admin    kafkapkg.KafkaAdmin
}

func NewMessagingService(producer *kafkapkg.Producer, admin kafkapkg.KafkaAdmin) MessagingService {
	return &MessagingServiceImpl{
		producer: producer,
		admin:    admin,
	}
}

func (s *MessagingServiceImpl) PublishMessage(ctx context.Context, cfg kafkapkg.KafkaConfig, req *pb.PublishRequest) *errors.CustomError {
	if req.Topic == "" {
		return errors.NewCustomError(errors.MSGErrInvalidTopic, fmt.Errorf("topic is required"))
	}
	if req.Value == nil {
		return errors.NewCustomError(errors.PUBErrInvalidMessage, fmt.Errorf("message content is required"))
	}

	valueBytes, err := json.Marshal(req.Value)
	if err != nil {
		return errors.NewCustomError(errors.PUBErrInvalidMessage, fmt.Errorf("failed to serialize message: %w", err))
	}

	writer := s.producer.GetWriter(cfg)
	defer func() {
		if err := s.producer.Close(writer); err != nil {
			logger.LogEvent("", "kafka_close_writer_failed", req.Topic, "error", fmt.Sprintf("Failed to close writer: %v", err))
		}
	}()

	headers := make([]kafka.Header, 0, len(req.Headers))
	for k, v := range req.Headers {
		headers = append(headers, kafka.Header{
			Key:   k,
			Value: []byte(v),
		})
	}

	key := []byte(req.Key)
	if len(key) == 0 {
		key = []byte(fmt.Sprintf("%s-%d", req.Topic, time.Now().UnixNano()))
	}

	err = s.producer.WriteWithRetry(ctx, writer, key, valueBytes, headers, 3)
	if err != nil {
		logger.LogEvent("", "kafka_publish_failed", req.Topic, "error", fmt.Sprintf("Failed to publish message: %v", err))
		return errors.NewCustomError(errors.PUBErrPublishFailed, fmt.Errorf("failed to publish message: %w", err))
	}

	return nil
}

func (s *MessagingServiceImpl) ConsumeMessage(stream pb.MessagingService_SubscribeV1Server, cfg kafkapkg.KafkaConfig) *errors.CustomError {
	msgChan := make(chan *pb.KafkaMessage, 1000)

	go func() {
		for msg := range msgChan {
			if err := stream.Send(msg); err != nil {
				logger.LogEvent("", "stream_send_failed", cfg.Topic, "error", err.Error())
				// Not returning error here, as this is a goroutine
			}
		}
	}()

	consumer := kafkapkg.NewConsumer(cfg, func(payload []byte) error {
		var value map[string]string
		if err := json.Unmarshal(payload, &value); err != nil {
			return fmt.Errorf("failed to deserialize message: %w", err)
		}

		msg := &pb.KafkaMessage{
			Value:     value,
			Timestamp: time.Now().UnixMilli(),
		}

		select {
		case msgChan <- msg:
		default:
			if err := stream.Send(msg); err != nil {
				return fmt.Errorf("failed to send message to stream: %w", err)
			}
		}
		return nil
	})

	consumer.Start(stream.Context())
	close(msgChan)
	return nil
}

func (s *MessagingServiceImpl) CreateTopic(ctx context.Context, req *pb.CreateTopicRequest, cfg kafkapkg.KafkaConfig) *errors.CustomError {
	if req.Topic == "" {
		return errors.NewCustomError(errors.MSGErrInvalidTopic, fmt.Errorf("topic name is required"))
	}

	if req.NumPartitions <= 0 {
		req.NumPartitions = 1
	}
	if req.ReplicationFactor <= 0 {
		req.ReplicationFactor = 1
	}

	if req.Config == nil {
		req.Config = &pb.TopicConfig{}
	}

	if req.Config.RetentionMs == 0 {
		req.Config.RetentionMs = 3000
	}
	if req.Config.CleanupPolicy == pb.CleanupPolicy_CLEANUP_POLICY_UNSPECIFIED {
		req.Config.CleanupPolicy = pb.CleanupPolicy_CLEANUP_POLICY_DELETE
	}

	kafkaConfig := map[string]string{
		"retention.ms":                    fmt.Sprintf("%d", req.Config.RetentionMs),
		"cleanup.policy":                  getCleanupPolicyString(req.Config.CleanupPolicy),
		"delete.retention.ms":             "1000",
		"segment.ms":                      "1000",
		"log.retention.check.interval.ms": "1000",
		"log.cleanup.interval.mins":       "1",
		"log.segment.delete.delay.ms":     "1000",
	}

	err := s.admin.CreateTopic(ctx, req.Topic, int(req.NumPartitions), int(req.ReplicationFactor), kafkaConfig)
	if err != nil {
		logger.LogEvent("kafka", "create_topic_failed", req.Topic, "error", err.Error())
		return errors.NewCustomError(errors.TOPErrCreateFailed, fmt.Errorf("failed to create topic: %w", err))
	}

	return nil
}

// Helper function to convert CleanupPolicy enum to string
func getCleanupPolicyString(policy pb.CleanupPolicy) string {
	switch policy {
	case pb.CleanupPolicy_CLEANUP_POLICY_DELETE:
		return "delete"
	case pb.CleanupPolicy_CLEANUP_POLICY_COMPACT:
		return "compact"
	default:
		return "delete"
	}
}
