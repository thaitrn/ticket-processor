package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"ticket-processor/internal/config"
	"ticket-processor/internal/kafka"

	"github.com/IBM/sarama"
)

type ConsumerHandler struct {
	ready chan bool
	wg    *sync.WaitGroup
}

func (h *ConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

func (h *ConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	h.wg.Done()
	return nil
}

func (h *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		log.Printf("Message received: topic=%s partition=%d offset=%d\n",
			message.Topic, message.Partition, message.Offset)

		// Process the message
		// TODO: Add your business logic here

		session.MarkMessage(message, "")
	}
	return nil
}

func main() {
	// Initialize configuration
	cfg := config.NewConfig()

	// Create consumer
	consumer, err := kafka.NewConsumer(cfg.KafkaBrokers, cfg.GroupID, cfg.Topic)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Start consuming
	wg := &sync.WaitGroup{}
	wg.Add(1)

	handler := &ConsumerHandler{
		ready: make(chan bool),
		wg:    wg,
	}

	go func() {
		for {
			if err := consumer.consumer.Consume(ctx, []string{cfg.Topic}, handler); err != nil {
				log.Printf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
			handler.ready = make(chan bool)
		}
	}()

	<-handler.ready
	log.Println("Consumer is ready")

	// Wait for shutdown signal
	<-signals
	log.Println("Shutting down...")

	cancel()
	wg.Wait()
}
