package kafkaService

import (
	"encoding/json"
	"fmt"
	"log"

	"protected_link/internal/modules/authentication/models"
	kafka "protected_link/pkg/kafka"
)

type INotifierService interface {
	SendNotification(payload models.MessagePayload, topic string) error
}

var _ INotifierService = (*NotifierService)(nil) // ensure it implements the interface

type NotifierService struct {
	Producer kafka.MessageProducer // use the interface, not *KafkaProducer
}

func NewNotifierService(producer kafka.MessageProducer) *NotifierService {
	return &NotifierService{Producer: producer}
}

func (ns *NotifierService) SendNotification(payload models.MessagePayload, topic string) error {
	fmt.Println("Sending message to Kafka topic:", topic)

	msgBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Failed to marshal message payload:", err)
		return err
	}

	err = ns.Producer.SendMessage(topic, msgBytes)
	if err != nil {
		log.Println("Failed to send message to Kafka:", err)
		return err
	}

	log.Println("Notification sent successfully to topic:", topic)
	return nil
}
