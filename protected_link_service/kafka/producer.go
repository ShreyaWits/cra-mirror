package kafka

import (
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

// KafkaProducer struct that holds a producer
type KafkaProducer struct {
	Producer sarama.SyncProducer
}

// NewKafkaProducer initializes a new Kafka producer with enhanced error handling
func NewKafkaProducer(brokers []string) (*KafkaProducer, error) {
	// Configuring the producer with retries and other settings
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	// Initialize Kafka producer
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		// Logging the error with context for troubleshooting
		log.Printf("❌ Failed to initialize Kafka producer: %v", err)
		// Returning the error so the caller can handle it
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	// Returning the initialized KafkaProducer
	return &KafkaProducer{Producer: producer}, nil
}

// SendMessage sends a message to the specified Kafka topic with enhanced error handling
func (kp *KafkaProducer) SendMessage(topic string, message []byte) error {
	if kp.Producer == nil {
		err := fmt.Errorf("producer is not initialized")
		log.Printf("❌ Kafka producer is not initialized: %v", err)
		return err
	}

	// Creating the message to send to Kafka
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message),
	}

	// Sending the message and checking for errors
	partition, offset, err := kp.Producer.SendMessage(msg)
	if err != nil {
		// Logging the error with relevant information
		log.Printf("❌ Failed to send message to Kafka topic %s: %v", topic, err)
		return fmt.Errorf("failed to send message to Kafka topic %s: %w", topic, err)
	}

	// Logging the successful message send with partition and offset details
	log.Printf("✅ Sent message to topic %s [partition: %d, offset: %d]", topic, partition, offset)
	return nil
}

// Close closes the Kafka producer with enhanced error handling
func (kp *KafkaProducer) Close() error {
	if kp.Producer == nil {
		err := fmt.Errorf("producer is not initialized, cannot close")
		log.Printf("❌ Kafka producer is not initialized: %v", err)
		return err
	}

	// Closing the producer and checking for errors
	err := kp.Producer.Close()
	if err != nil {
		log.Printf("❌ Failed to close Kafka producer: %v", err)
		return fmt.Errorf("failed to close Kafka producer: %w", err)
	}

	// Logging successful closure of producer
	log.Println("✅ Kafka producer closed successfully")
	return nil
}
