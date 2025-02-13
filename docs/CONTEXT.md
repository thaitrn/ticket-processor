# Kafka Message Processing System in Go

## Project Overview
Develop a high-performance Go application to process ticket messages using Apache Kafka, implementing both producer and consumer functionality with the Sarama library.

## Requirements
- Process 100 tickets per second
- Maintain message ordering
- Handle both message production and consumption
- Ensure reliable message delivery
- Monitor system performance with Prometheus and Grafana
- Implement structured logging with zap

## Technical Stack
- Language: Go
- Kafka Client: Sarama (github.com/IBM/sarama)
- Message Broker: Apache Kafka
- Metrics: Prometheus
- Visualization: Grafana
- Logging: zap
- Containerization: Docker

## Key Features
1. Message Producer
   - Generate ticket messages
   - Ensure sequential message ordering using OrderID as partition key
   - Implement proper partitioning strategy
   - Handle back-pressure scenarios
   - Rate limiting to 100 messages/second

2. Message Consumer
   - Process messages in order
   - Implement consumer group functionality
   - Handle message acknowledgments
   - Implement error handling and retry logic
   - Track consumer lag

3. Monitoring
   - Prometheus metrics collection
   - Grafana dashboards
   - Health check endpoints
   - Structured logging

## Performance Goals
- Throughput: 100 messages/second
- Low latency processing
- Message ordering guarantee
- Fault tolerance
- Real-time metrics monitoring

## Implementation Considerations
- Use proper Kafka partitioning to maintain message order
- Implement consumer groups for scalability
- Configure appropriate batch sizes and compression
- Implement proper error handling and logging
- Monitor system performance metrics
- Use Docker for containerization
- Implement graceful shutdown

## Implementation Steps

### 1. Environment Setup
1. Install Go (1.21 or later)
2. Set up Docker and Docker Compose
3. Initialize Go project:
   ```bash
   go mod init ticket-processor
   go get github.com/IBM/sarama
   go get github.com/prometheus/client_golang/prometheus
   go get go.uber.org/zap
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
│   ├── kafka/
│   │   ├── producer.go
│   │   └── consumer.go
│   ├── metrics/
│   │   └── metrics.go
│   └── monitoring/
│       └── server.go
├── dashboards/
│   ├── kafka-overview.json
│   └── kafka-metrics.json
├── grafana-provisioning/
│   ├── dashboards/
│   │   └── kafka.yml
│   └── datasources/
│       └── prometheus.yml
├── prometheus.yml
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

### 3. Docker Configuration

#### Dockerfile
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build producer and consumer
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/producer ./cmd/producer
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/consumer ./cmd/consumer

# Create final lightweight images
FROM alpine:3.19

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /bin/producer /bin/producer
COPY --from=builder /bin/consumer /bin/consumer

# Create non-root user
RUN adduser -D appuser
USER appuser

# Set environment variables
ENV KAFKA_BROKERS=kafka:9092
ENV KAFKA_TOPIC=tickets
ENV GROUP_ID=ticket-processor-group
ENV METRICS_PORT=2112

# Default command (can be overridden)
CMD ["/bin/producer"]
```

#### Docker Compose
Create `docker-compose.yml`:
```yaml
version: '3.8'

services:
  zookeeper:
    image: confluentinc/cp-zookeeper:latest
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000
    ports:
      - "2181:2181"
    healthcheck:
      test: ["CMD", "nc", "-z", "localhost", "2181"]
      interval: 10s
      timeout: 5s
      retries: 5

  kafka:
    image: confluentinc/cp-kafka:latest
    depends_on:
      zookeeper:
        condition: service_healthy
    ports:
      - "9092:9092"
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://kafka:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS: 0
    healthcheck:
      test: ["CMD", "nc", "-z", "localhost", "9092"]
      interval: 10s
      timeout: 5s
      retries: 5

  producer:
    build:
      context: .
      dockerfile: Dockerfile
    command: ["/bin/producer"]
    depends_on:
      kafka:
        condition: service_healthy
    environment:
      KAFKA_BROKERS: kafka:9092
      KAFKA_TOPIC: tickets
      METRICS_PORT: 2112
    ports:
      - "2112:2112"

  consumer:
    build:
      context: .
      dockerfile: Dockerfile
    command: ["/bin/consumer"]
    depends_on:
      kafka:
        condition: service_healthy
    environment:
      KAFKA_BROKERS: kafka:9092
      KAFKA_TOPIC: tickets
      GROUP_ID: ticket-processor-group
      METRICS_PORT: 2113
    ports:
      - "2113:2113"

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    depends_on:
      - producer
      - consumer

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    depends_on:
      - prometheus
    volumes:
      - ./dashboards:/var/lib/grafana/dashboards
      - ./grafana-provisioning:/etc/grafana/provisioning
```

### 4. Monitoring Configuration

#### Prometheus Configuration
Create `prometheus.yml`:
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'producer'
    static_configs:
      - targets: ['producer:2112']

  - job_name: 'consumer'
    static_configs:
      - targets: ['consumer:2113']
```

#### Grafana Configuration
Create `grafana-provisioning/datasources/prometheus.yml`:
```yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
```

Create `grafana-provisioning/dashboards/kafka.yml`:
```yaml
apiVersion: 1

providers:
  - name: 'Kafka'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    editable: true
    options:
      path: /var/lib/grafana/dashboards
```

### 5. Running the Application

1. Build and start all services:
```bash
docker-compose up --build
```

2. Access monitoring:
   - Grafana: http://localhost:3000 (admin/admin)
   - Prometheus: http://localhost:9090
   - Producer metrics: http://localhost:2112/metrics
   - Consumer metrics: http://localhost:2113/metrics

3. Monitor logs:
```bash
docker-compose logs -f
```

4. Shutdown:
```bash
docker-compose down
```

### 6. Monitoring Features
- Message production/consumption rates
- Consumer lag
- Processing duration
- Error counts
- Health status
- System metrics

### 7. Next Steps
1. Add more comprehensive metrics
2. Implement retry strategies
3. Add alerting rules
4. Implement message validation
5. Add integration tests
6. Implement message schema validation
7. Add message compression
8. Implement dead letter queue
9. Add message tracing
10. Implement backpressure handling

Would you like me to provide implementation details for any of these next steps?