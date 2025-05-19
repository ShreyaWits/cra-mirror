package confluent_test

import (
	"context"
	"fmt"
	"messaging_service/internal/modules/message_broker/mock"
	"messaging_service/pkg/confluent"
	"messaging_service/pkg/observability"
	"sync"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/golang/mock/gomock"
)

var obs = &observability.ObservabilityStack{}

// TestKafkaMonitorDeliveryReports tests handling of Kafka delivery reports
func TestKafkaMonitorDeliveryReports(t *testing.T) {

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
				confluent.KafkaMonitorDeliveryReports(producer, "test-topic", obs)
			}()

			// Wait for the goroutine to finish processing
			wg.Wait()
		})
	}
}
func TestLogKafkaError(t *testing.T) {

	confluent.LogKafkaError(context.Background(), "test_event", "test-topic", "test message", fmt.Errorf("sample error"), obs)

}

func TestKafkaWarning(t *testing.T) {

	confluent.LogKafkaWarning(context.Background(), "test_event", "test-topic", "test message", obs)

}

func TestLogKafkaInfo(t *testing.T) {

	confluent.LogKafkaInfo(context.Background(), "test_event", "test-topic", "test message", obs)
}
