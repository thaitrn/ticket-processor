# Kafka Message Processing System in Go

## Project Overview
Develop a high-performance Go application to process ticket messages using Apache Kafka, implementing both producer and consumer functionality with the Sarama library.

## Requirements
- Process 100 tickets per second
- Maintain message ordering
- Handle both message production and consumption
- Ensure reliable message delivery

## Technical Stack
- Language: Go
- Kafka Client: Sarama (github.com/IBM/sarama)
- Message Broker: Apache Kafka

## Key Features
1. Message Producer
   - Generate ticket messages
   - Ensure sequential message ordering
   - Implement proper partitioning strategy
   - Handle back-pressure scenarios

2. Message Consumer
   - Process messages in order
   - Implement consumer group functionality
   - Handle message acknowledgments
   - Implement error handling and retry logic

## Performance Goals
- Throughput: 100 messages/second
- Low latency processing
- Message ordering guarantee
- Fault tolerance

## Implementation Considerations
- Use proper Kafka partitioning to maintain message order
- Implement consumer groups for scalability
- Configure appropriate batch sizes and compression
- Implement proper error handling and logging
- Monitor system performance metrics

## Implementation Steps

### 1. Environment Setup
1. Install Go (1.20 or later)
2. Set up Kafka locally:
   ```bash
   docker-compose up -d
   ```
3. Initialize Go project:
   ```bash
   go mod init ticket-processor
   go get github.com/IBM/sarama
   go get github.com/prometheus/client_golang/prometheus
   go get github.com/cenkalti/backoff/v4
   go get github.com/stretchr/testify
   ```

### 2. Project Structure
```
ticket-processor/
├── cmd/
│   ├── producer/
│   │   └── main.go
│   └── consumer/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── models/
│   │   └── ticket.go
│   └── kafka/
│       ├── producer.go
│       └── consumer.go
├── docker-compose.yml
└── go.mod
```

### 3. Create Docker Environment
Create `docker-compose.yml`:
```yaml
version: '3'
services:
  zookeeper:
    image: confluentinc/cp-zookeeper:latest
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000
    ports:
      - "2181:2181"

  kafka:
    image: confluentinc/cp-kafka:latest
    depends_on:
      - zookeeper
    ports:
      - "9092:9092"
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
```

### 4. Implement Core Components

#### 4.1. Create Ticket Model
Create `internal/models/ticket.go`:
```go
package models

import (
    "time"
)

type Ticket struct {
    ID        string    `json:"id"`
    Timestamp time.Time `json:"timestamp"`
    OrderID   int64     `json:"order_id"`
    Data      string    `json:"data"`
}
```

#### 4.2. Create Configuration
Create `internal/config/config.go`:
```go
package config

type Config struct {
    KafkaBrokers []string
    Topic        string
    GroupID      string
    BatchSize    int
    BatchTimeout time.Duration
}

func NewConfig() *Config {
    return &Config{
        KafkaBrokers: []string{"localhost:9092"},
        Topic:        "tickets",
        GroupID:      "ticket-processor-group",
        BatchSize:    100,
        BatchTimeout: time.Second * 1,
    }
}
```

#### 4.3. Implement Producer
Create `internal/kafka/producer.go`:
```go
package kafka

import (
    "encoding/json"
    "github.com/IBM/sarama"
    "ticket-processor/internal/models"
)

type Producer struct {
    producer sarama.SyncProducer
    topic    string
}

func NewProducer(brokers []string, topic string) (*Producer, error) {
    config := sarama.NewConfig()
    config.Producer.RequiredAcks = sarama.WaitForAll
    config.Producer.Retry.Max = 5
    config.Producer.Return.Successes = true
    
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

    msg := &sarama.ProducerMessage{
        Topic: p.topic,
        Key:   sarama.StringEncoder(ticket.ID),
        Value: sarama.ByteEncoder(data),
    }

    _, _, err = p.producer.SendMessage(msg)
    return err
}
```

#### 4.4. Implement Consumer
Create `internal/kafka/consumer.go`:
```go
package kafka

import (
    "encoding/json"
    "github.com/IBM/sarama"
    "ticket-processor/internal/models"
)

type Consumer struct {
    consumer sarama.ConsumerGroup
    topic    string
}

func NewConsumer(brokers []string, groupID string, topic string) (*Consumer, error) {
    config := sarama.NewConfig()
    config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
    config.Consumer.Offsets.Initial = sarama.OffsetOldest

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
    ready chan bool
}

func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
    for message := range claim.Messages() {
        var ticket models.Ticket
        if err := json.Unmarshal(message.Value, &ticket); err != nil {
            // Handle error
            continue
        }

        // Process ticket here
        
        session.MarkMessage(message, "")
    }
    return nil
}
```

### 5. Next Implementation Tasks

1. Create main producer application (`cmd/producer/main.go`)
2. Create main consumer application (`cmd/consumer/main.go`)
3. Add logging and metrics
4. Implement error handling and retries
5. Add tests
6. Performance optimization

Would you like me to provide the implementation for any of these next tasks?