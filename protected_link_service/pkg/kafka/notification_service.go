package kafkaService

import (
	"encoding/json"
	"log"

	"protected_link/internal/modules/authentication/models"
	"protected_link/kafka"
)

type NotifierService struct {
	Producer *kafka.KafkaProducer
}

func NewNotifierService(producer *kafka.KafkaProducer) *NotifierService {
	return &NotifierService{Producer: producer}
}

func (ns *NotifierService) SendNotification(payload models.MessagePayload, topic string) error {
	msgBytes, err := json.Marshal(payload)
	if err != nil {
		log.Println("Failed to marshal payload:", err)
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
