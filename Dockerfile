# Multi-stage build for EVM TVL Aggregator
FROM golang:1.21-alpine AS builder

# Install necessary packages
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the applications
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o aggregator ./cmd/aggregator
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o indexer ./cmd/indexer
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o tui ./cmd/tui

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appuser && \
    adduser -S -D -H -u 1001 -s /sbin/nologin -G appuser appuser

WORKDIR /root/

# Copy binaries from builder
COPY --from=builder /app/api .
COPY --from=builder /app/aggregator .
COPY --from=builder /app/indexer .
COPY --from=builder /app/tui .

# Copy configuration files
COPY --from=builder /app/.env.example .env

# Change ownership
RUN chown -R appuser:appuser /root/

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Default command (can be overridden)
CMD ["./api"]