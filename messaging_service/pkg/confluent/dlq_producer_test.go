package confluent_test

import (
	"testing"
)

func TestNewDLQProducer(t *testing.T) {
	// Skip this test as it requires real Kafka connections and proper observability stack
	t.Skip("Skipping test that requires real Kafka connections")
}

func TestSendToDLQ(t *testing.T) {
	// Skip this test as it requires proper observability stack
	t.Skip("Skipping test that requires real observability stack")
}

func TestDLQClose(t *testing.T) {
	// Skip this test as it requires proper observability stack
	t.Skip("Skipping test that requires real observability stack")
}
