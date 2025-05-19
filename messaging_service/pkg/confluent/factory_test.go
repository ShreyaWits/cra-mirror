package confluent_test

import (
	"context"
	"messaging_service/internal/config"
	"messaging_service/pkg/confluent"
	"messaging_service/pkg/observability"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFactoryMethods(t *testing.T) {
	tests := []struct {
		name          string
		method        func(confluent.KafkaFactory) (interface{}, error)
		expectedError error
	}{
		{
			name: "CreateProducer",
			method: func(f confluent.KafkaFactory) (interface{}, error) {
				cfg := confluent.NewDefaultKafkaConfig([]string{"localhost:9092"}, "test-topic")
				cfg.SkipTransactionInit = true
				return f.CreateProducer(context.Background(), cfg)
			},
			expectedError: nil,
		},
		{
			name: "CreateConsumer",
			method: func(f confluent.KafkaFactory) (interface{}, error) {
				cfg := confluent.NewDefaultKafkaConfig([]string{"localhost:9092"}, "test-topic")
				groupID := "test-group"
				handler := func(data []byte) error { return nil }
				return f.CreateConsumer(context.Background(), cfg, groupID, handler)
			},
			expectedError: nil,
		},
		{
			name: "CreateAdmin",
			method: func(f confluent.KafkaFactory) (interface{}, error) {
				cfg := config.GetMockConfig()

				return f.CreateAdmin(cfg)
			},
			expectedError: nil,
		},
		{
			name: "CreateDLQProducer",
			method: func(f confluent.KafkaFactory) (interface{}, error) {
				cfg := confluent.NewDefaultKafkaConfig([]string{"localhost:9092"}, "dlq-topic")
				return f.CreateDLQProducer(cfg)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := confluent.NewFactory(&observability.ObservabilityStack{})
			result, err := tt.method(factory)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}
