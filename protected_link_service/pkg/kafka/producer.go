package kafkaService

import (
	"log"

	"github.com/IBM/sarama"
)

// MessageProducer is the interface to abstract Kafka producer behavior
type MessageProducer interface {
	SendMessage(topic string, message []byte) error
}

type KafkaProducer struct {
	Producer sarama.SyncProducer
}

func NewKafkaProducer(brokers []string) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &KafkaProducer{Producer: producer}, nil
}

func (kp *KafkaProducer) SendMessage(topic string, message []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message),
	}

	if kp.Producer != nil {
		partition, offset, err := kp.Producer.SendMessage(msg)
		if err != nil {
			return err
		}
		log.Printf("Sent to topic %s [partition: %d, offset: %d]", topic, partition, offset)
	}
	return nil
}

func (kp *KafkaProducer) Close() error {
	return kp.Producer.Close()
}
