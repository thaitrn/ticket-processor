package kafka

import (
	"testing"
	"ticket-processor/internal/models"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProducer_SendTicket(t *testing.T) {
	// Create test config
	brokers := []string{"localhost:9092"}
	topic := "test-topic"

	// Create producer
	producer, err := NewProducer(brokers, topic)
	assert.NoError(t, err)
	defer producer.Close()

	// Create test ticket
	ticket := &models.Ticket{
		ID:        "test-id",
		Timestamp: time.Now(),
		OrderID:   1,
		Data:      "test-data",
	}

	// Send ticket
	err = producer.SendTicket(ticket)
	assert.NoError(t, err)
}
