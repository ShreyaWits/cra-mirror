package confluent_test

import (
	"fmt"
	"messaging_service/internal/messaging_service/mock"
	"messaging_service/pkg/confluent"
	"messaging_service/pkg/logger"
	"sync"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/golang/mock/gomock"
)

// TestKafkaMonitorDeliveryReports tests handling of Kafka delivery reports
func TestKafkaMonitorDeliveryReports(t *testing.T) {
	logger.InitLogger()
	topic := "test-topic"
	topicPtr := &topic

	tests := []struct {
		name   string
		events []kafka.Event
	}{
		{
			name: "successful delivery",
			events: []kafka.Event{
				&kafka.Message{
					TopicPartition: kafka.TopicPartition{
						Topic:     topicPtr,
						Partition: 1,
						Offset:    10,
						Error:     nil,
					},
				},
			},
		},
		{
			name: "failed delivery",
			events: []kafka.Event{
				&kafka.Message{
					TopicPartition: kafka.TopicPartition{
						Topic:     topicPtr,
						Partition: 0,
						Offset:    0,
						Error:     fmt.Errorf("delivery failed"),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventChan := make(chan kafka.Event, len(tt.events))
			for _, e := range tt.events {
				eventChan <- e
			}
			close(eventChan)

			// Mock the producer instead of creating a real one
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			producer := mock.NewMockKafkaProducerInterface(ctrl)
			producer.EXPECT().Events().Return(eventChan)

			// Use a WaitGroup to wait for the goroutine to finish
			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				defer wg.Done()
				confluent.KafkaMonitorDeliveryReports(producer, "test-topic")
			}()

			// Wait for the goroutine to finish processing
			wg.Wait()
		})
	}
}
func TestLogKafkaError(t *testing.T) {

	logger.InitLogger()

	confluent.LogKafkaError("test_event", "test-topic", "test message", fmt.Errorf("sample error"))

}

func TestKafkaWarning(t *testing.T) {

	logger.InitLogger()

	confluent.LogKafkaWarning("test_event", "test-topic", "test message")

}

func TestLogKafkaInfo(t *testing.T) {
	logger.InitLogger()

	confluent.LogKafkaInfo("test_event", "test-topic", "test message")
}
