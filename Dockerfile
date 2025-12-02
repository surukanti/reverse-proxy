# Build stage
FROM golang:1.24.2-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/proxy ./cmd/proxy

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS backend connections
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/bin/proxy /usr/local/bin/proxy

# Create config directory
RUN mkdir -p /etc/proxy/config

# Copy example configurations
COPY examples/ /etc/proxy/examples/

# Default configuration file
ENV CONFIG_FILE=/etc/proxy/config/config.yaml

# Expose port 9000
EXPOSE 9000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9000/health || exit 1

# Run the proxy - use shell form to expand environment variables
ENTRYPOINT ["/bin/sh", "-c"]
CMD ["proxy -config ${CONFIG_FILE}"]
