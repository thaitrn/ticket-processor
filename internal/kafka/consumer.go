package kafka

import (
	"encoding/json"
	"sync"
	"time"

	"ticket-processor/internal/metrics"
	"ticket-processor/internal/models"

	"github.com/IBM/sarama"
)

type Consumer struct {
	consumer sarama.ConsumerGroup
	topic    string
}

func NewConsumer(brokers []string, groupID string, topic string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	// Enable consumer group heartbeat
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	// Set session timeout
	config.Consumer.Group.Session.Timeout = 10 * time.Second
	// Enable return errors
	config.Consumer.Return.Errors = true

	group, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumer: group,
		topic:    topic,
	}, nil
}

type Handler struct {
	ready   chan bool
	metrics *metrics.Metrics
	wg      *sync.WaitGroup
}

func NewHandler(metrics *metrics.Metrics, wg *sync.WaitGroup) *Handler {
	return &Handler{
		ready:   make(chan bool),
		metrics: metrics,
		wg:      wg,
	}
}

func (h *Handler) Setup(sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

func (h *Handler) Cleanup(sarama.ConsumerGroupSession) error {
	h.wg.Done()
	return nil
}

func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		startTime := time.Now()

		var ticket models.Ticket
		if err := json.Unmarshal(message.Value, &ticket); err != nil {
			h.metrics.ErrorsCount.Inc()
			continue
		}

		// Process ticket here
		// Using OrderID as partition key to maintain ordering
		h.metrics.MessagesConsumed.Inc()
		h.metrics.ProcessingDuration.Observe(time.Since(startTime).Seconds())

		// Calculate and update consumer lag
		lag := time.Since(ticket.Timestamp).Seconds()
		h.metrics.ConsumerLag.Set(lag)

		session.MarkMessage(message, "")
	}
	return nil
}
