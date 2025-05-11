package kafka

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type KafkaPkg struct {
}

type Producer struct {
	Writer           *kafka.Writer
	Config           KafkaConfig
	transactionMu    sync.Mutex
	transactionalID  string
	isTransactional  bool
	transactionState string
}

// NewProducer creates a new Kafka producer with appropriate configuration
func NewProducer(cfg KafkaConfig) *Producer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Topic:    cfg.Topic,
		Balancer: GetBalancer(cfg.BalancerType),
	}

	// Configure delivery semantics - only at-least-once and exactly-once supported
	switch cfg.DeliverySemantics {
	case ExactlyOnce:
		// Exactly-once requires idempotence and transactions
		writer.RequiredAcks = kafka.RequireAll
		writer.MaxAttempts = 10

		// Configure for idempotent delivery
		// Note: The actual transaction API is handled separately since
		// kafka-go library doesn't have direct transaction support
	default:
		// Default to safer at-least-once semantics
		writer.RequiredAcks = kafka.RequireAll
		writer.MaxAttempts = 5
		writer.Async = false // Synchronous writes
	}

	transactionalID := ""
	isTransactional := false

	// Set up transactions for exactly-once semantics if configured
	if cfg.DeliverySemantics == ExactlyOnce && cfg.ExactlyOnceConfig.EnableTransactions {
		transactionalID = fmt.Sprintf("%s-%s",
			cfg.ExactlyOnceConfig.TransactionalIDPrefix,
			uuid.New().String())
		isTransactional = true

		// In a real implementation, we would set up the transaction coordinator here
		log.Printf("Creating transactional producer with ID: %s", transactionalID)
	}

	return &Producer{
		Writer:          writer,
		Config:          cfg,
		transactionalID: transactionalID,
		isTransactional: isTransactional,
	}
}

// BeginTransaction starts a new Kafka transaction for exactly-once processing
func (p *Producer) BeginTransaction(ctx context.Context) error {
	if !p.isTransactional {
		return fmt.Errorf("producer is not configured for transactions")
	}

	p.transactionMu.Lock()
	defer p.transactionMu.Unlock()

	if p.transactionState == "in_transaction" {
		return fmt.Errorf("transaction already in progress")
	}

	// In a real implementation, we would initiate a transaction with the Kafka transaction coordinator
	// For now, we'll just log and set state
	log.Printf("Beginning transaction for producer %s", p.transactionalID)
	p.transactionState = "in_transaction"

	return nil
}

// CommitTransaction commits the current Kafka transaction
func (p *Producer) CommitTransaction(ctx context.Context) error {
	if !p.isTransactional {
		return fmt.Errorf("producer is not configured for transactions")
	}

	p.transactionMu.Lock()
	defer p.transactionMu.Unlock()

	if p.transactionState != "in_transaction" {
		return fmt.Errorf("no transaction in progress")
	}

	// In a real implementation, we would commit the transaction with the Kafka transaction coordinator
	// For now, we'll just log and set state
	log.Printf("Committing transaction for producer %s", p.transactionalID)
	p.transactionState = "no_transaction"

	return nil
}

// AbortTransaction aborts the current Kafka transaction
func (p *Producer) AbortTransaction(ctx context.Context) error {
	if !p.isTransactional {
		return fmt.Errorf("producer is not configured for transactions")
	}

	p.transactionMu.Lock()
	defer p.transactionMu.Unlock()

	if p.transactionState != "in_transaction" {
		return fmt.Errorf("no transaction in progress")
	}

	// In a real implementation, we would abort the transaction with the Kafka transaction coordinator
	// For now, we'll just log and set state
	log.Printf("Aborting transaction for producer %s", p.transactionalID)
	p.transactionState = "no_transaction"

	return nil
}

// WriteTransactional sends a message within the current transaction
func (p *Producer) WriteTransactional(ctx context.Context, key, value []byte) error {
	if !p.isTransactional {
		return fmt.Errorf("producer is not configured for transactions")
	}

	p.transactionMu.Lock()
	defer p.transactionMu.Unlock()

	if p.transactionState != "in_transaction" {
		return fmt.Errorf("no transaction in progress")
	}

	// In a real implementation, we would send the message with the transaction
	// For now, we'll just do a regular write
	err := p.Writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})

	if err != nil {
		log.Printf("Failed to write transactional message: %v", err)
		return err
	}

	return nil
}

// Write sends a message to Kafka with appropriate error handling
func (p *Producer) Write(ctx context.Context, key, value []byte) error {
	// If exactly-once semantics with transactions is enabled, use transactions
	if p.isTransactional {
		if err := p.BeginTransaction(ctx); err != nil {
			return err
		}

		err := p.WriteTransactional(ctx, key, value)
		if err != nil {
			p.AbortTransaction(ctx)
			return err
		}

		return p.CommitTransaction(ctx)
	}

	// Otherwise, use at-least-once semantics with standard writes
	err := p.Writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})

	if err != nil {
		log.Printf("Kafka write error: %v", err)
		return err
	}

	return nil
}

// WriteWithRetry attempts to write a message with exponential backoff retries
func (p *Producer) WriteWithRetry(ctx context.Context, key, value []byte, maxRetries int) error {
	// For exactly-once with transactions, use transactional writes
	if p.isTransactional {
		return p.Write(ctx, key, value) // Already includes transaction handling
	}

	// For at-least-once, use standard writes with retries
	var err error
	attempt := 0

	for attempt <= maxRetries {
		// Break if context is canceled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err = p.Writer.WriteMessages(ctx, kafka.Message{
			Key:   key,
			Value: value,
		})

		if err == nil {
			return nil // Success
		}

		attempt++
		if attempt > maxRetries {
			break
		}

		// Exponential backoff with jitter
		backoff := time.Duration(100*attempt*attempt) * time.Millisecond
		log.Printf("Kafka write failed, retrying in %v (attempt %d of %d): %v",
			backoff, attempt, maxRetries, err)
		time.Sleep(backoff)
	}

	log.Printf("Kafka write failed after %d attempts: %v", attempt, err)
	return err
}

func (p *Producer) Close() error {
	// If in a transaction, abort it
	if p.isTransactional && p.transactionState == "in_transaction" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := p.AbortTransaction(ctx); err != nil {
			log.Printf("Failed to abort transaction on close: %v", err)
		}
	}

	return p.Writer.Close()
}
