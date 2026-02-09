# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git make

# Set GOTOOLCHAIN to auto to allow downloading newer Go versions
ENV GOTOOLCHAIN=auto

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o victorialogs-aggregator ./cmd/vl-aggregator

# Final stage
FROM alpine:3.21

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -u 1000 appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/victorialogs-aggregator .

# Create config directory
RUN mkdir -p /etc/victorialogs-aggregator && \
    chown -R appuser:appuser /app /etc/victorialogs-aggregator

USER appuser

EXPOSE 8085

ENV CONFIG_PATH=/etc/victorialogs-aggregator/config.yaml

CMD ["./victorialogs-aggregator"]
