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