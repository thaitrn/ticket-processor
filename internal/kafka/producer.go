package kafka

import (
	"encoding/json"
	"strconv"

	"ticket-processor/internal/metrics"
	"ticket-processor/internal/models"

	"github.com/IBM/sarama"
)

type Producer struct {
	producer sarama.SyncProducer
	topic    string
	metrics  *metrics.Metrics
}

func NewProducer(brokers []string, topic string) (*Producer, error) {
	config := sarama.NewConfig()
	// Ensure synchronous producer for reliability
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	// Enable idempotent producer to prevent duplicates
	config.Producer.Idempotent = true
	// Use consistent partitioning for ordering
	config.Producer.Partitioner = sarama.NewHashPartitioner

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Producer{
		producer: producer,
		topic:    topic,
	}, nil
}

func (p *Producer) SendTicket(ticket *models.Ticket) error {
	data, err := json.Marshal(ticket)
	if err != nil {
		return err
	}

	// Use OrderID as the partition key to maintain message ordering
	msg := &sarama.ProducerMessage{
		Topic:     p.topic,
		Key:       sarama.StringEncoder(strconv.FormatInt(ticket.OrderID, 10)),
		Value:     sarama.ByteEncoder(data),
		Timestamp: ticket.Timestamp,
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		p.metrics.ErrorsCount.Inc()
		return err
	}

	p.metrics.MessagesProduced.Inc()
	return nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}
