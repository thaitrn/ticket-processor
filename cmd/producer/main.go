package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"ticket-processor/internal/config"
	"ticket-processor/internal/kafka"
	"ticket-processor/internal/metrics"
	"ticket-processor/internal/models"
	"ticket-processor/internal/monitoring"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Initialize configuration
	cfg := config.NewConfig()

	// Initialize metrics
	metrics := metrics.NewMetrics("ticket_processor")

	// Start monitoring server
	monitoringServer := monitoring.NewServer(logger, "2112")
	go func() {
		if err := monitoringServer.Start(); err != nil {
			logger.Fatal("Failed to start monitoring server", zap.Error(err))
		}
	}()

	// Create producer
	producer, err := kafka.NewProducer(cfg.KafkaBrokers, cfg.Topic)
	if err != nil {
		logger.Fatal("Failed to create producer", zap.Error(err))
	}
	defer producer.Close()

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Create rate limiter for 100 messages per second
	ticker := time.NewTicker(time.Second / 100)
	defer ticker.Stop()

	// Generate and send tickets
	ticketCounter := int64(0)

	logger.Info("Starting ticket producer")

	for {
		select {
		case <-signals:
			logger.Info("Shutting down producer...")
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			ticket := &models.Ticket{
				ID:        uuid.New().String(),
				Timestamp: time.Now(),
				OrderID:   ticketCounter,
				Data:      fmt.Sprintf("Ticket data %d", ticketCounter),
			}

			if err := producer.SendTicket(ticket); err != nil {
				logger.Error("Failed to send ticket",
					zap.Error(err),
					zap.String("ticket_id", ticket.ID),
					zap.Int64("order_id", ticket.OrderID))
				continue
			}

			logger.Info("Sent ticket",
				zap.String("ticket_id", ticket.ID),
				zap.Int64("order_id", ticket.OrderID))
			ticketCounter++
		}
	}
}
