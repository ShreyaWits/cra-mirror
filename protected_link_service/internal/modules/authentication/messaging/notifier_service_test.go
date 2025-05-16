package kafkaService

import (
	"encoding/json"
	"errors"
	"testing"

	"protected_link/internal/modules/authentication/models"

	"github.com/stretchr/testify/assert"
)

// MockKafkaProducer mocks the KafkaProducer
type MockKafkaProducer struct {
	SendMessageFunc func(topic string, message []byte) error
}

func (m *MockKafkaProducer) SendMessage(topic string, message []byte) error {
	return m.SendMessageFunc(topic, message)
}

func TestSendNotification_Success(t *testing.T) {
	mockProducer := &MockKafkaProducer{
		SendMessageFunc: func(topic string, message []byte) error {
			var payload models.MessagePayload
			err := json.Unmarshal(message, &payload)
			assert.NoError(t, err)
			assert.Equal(t, "template-123", payload.TemplateID)
			assert.Equal(t, "notifications", topic)
			return nil
		},
	}

	service := NewNotifierService(mockProducer)

	payload := models.MessagePayload{
		TemplateID: "template-123",
		Channels:   []string{"email", "sms"},
		Recipients: []models.Recipient{
			{
				UserID: "user-1",
				Email:  "user@example.com",
				Phone:  "1234567890",
				Data:   map[string]string{"name": "John"},
			},
		},
	}

	err := service.SendNotification(payload, "notifications")
	assert.NoError(t, err)
}

func TestSendNotification_ProducerError(t *testing.T) {
	mockProducer := &MockKafkaProducer{
		SendMessageFunc: func(topic string, message []byte) error {
			return errors.New("failed to send")
		},
	}

	service := NewNotifierService(mockProducer)

	payload := models.MessagePayload{
		TemplateID: "template-999",
		Channels:   []string{"sms"},
		Recipients: []models.Recipient{
			{
				UserID: "user-2",
				Phone:  "9876543210",
				Data:   map[string]string{"code": "1234"},
			},
		},
	}

	err := service.SendNotification(payload, "alerts")
	assert.Error(t, err)
}
