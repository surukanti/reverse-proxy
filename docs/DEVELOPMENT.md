# Development Guide

This guide provides information for developers working on the reverse proxy project.

## Quick Start

### Building and Running

```bash
# Clone and setup
git clone https://github.com/surukanti/reverse-proxy.git
cd reverse-proxy

# Install dependencies
go mod download

# Build
make build
# or
go build -o bin/proxy ./cmd/proxy

# Run with default config
./bin/proxy

# Run with specific config
./bin/proxy -config configs/config.yaml

# Run with example config
./bin/proxy -config examples/configs/1-microservices-gateway.yaml
```

### Docker Development

```bash
# Build and run with Docker
docker-compose up --build

# Run specific example
cd examples/docker
docker-compose -f 11-basic-nginx-backends.yaml up --build
```

## Architecture Overview

### Core Components

- **Proxy**: Main request handling and routing logic
- **Router**: URL matching and route resolution
- **Backend**: Server pool management and load balancing
- **Middleware**: Request/response processing pipeline
- **Config**: Configuration loading and validation

### Request Flow

```
Client Request → Middleware Chain → Router → Backend Selection → Proxy → Response
```

## CI/CD Pipelines

### GitHub Actions Workflows

The project uses automated CI/CD pipelines to ensure code quality and reliability:

#### Go CI Workflow (`go.yml`)
- **Triggers**: Push and PR to `main` branch
- **Environment**: Ubuntu latest with Go 1.24.2
- **Steps**:
  1. Checkout code
  2. Setup Go environment
  3. Build project (`go build -v ./...`)
  4. Run tests (`go test -v ./...`)

#### SLSA3 Release Workflow (`go-ossf-slsa3-publish.yml`)
- **Triggers**: Release creation or manual dispatch
- **Purpose**: Secure software supply chain
- **Features**:
  - Generates provenance attestations
  - SLSA Level 3 compliance
  - OpenSSF framework integration

### Local Development with Make

The project includes a comprehensive Makefile for development tasks:

```bash
# Core development
make build          # Build binary
make test           # Run tests
make coverage       # Generate coverage report
make fmt            # Format code
make vet            # Run go vet
make lint           # Run golangci-lint

# Docker operations
make build-docker   # Build Docker image
make run-docker     # Run container
make test-docker    # Test container endpoints
make clean-docker   # Remove containers/images

# Full stack
make run-compose    # Start all services
make logs-compose   # View all logs
make clean-compose  # Stop and clean services

# Utilities
make help           # Show all targets
make version        # Show versions
make info           # Show build info
```

### Testing Strategy

#### Unit Tests
- Run automatically on every push/PR via GitHub Actions
- Comprehensive coverage of all packages
- Includes race detection for concurrent code
- Generates coverage reports

#### Integration Tests
- Docker-based testing with `make test-docker`
- Tests actual HTTP endpoints and health checks
- Validates CORS, rate limiting, and routing

#### Benchmarking
```bash
make bench          # Run performance benchmarks
go test -bench=. -benchmem ./internal/proxy
```

## Development Tasks

### Adding New Middleware

1. Create middleware in `internal/middleware/`
2. Implement `Handler` interface
3. Add to middleware chain in main.go
4. Add configuration support
5. Write tests

Example:
```go
type CustomMiddleware struct {
    // config fields
}

func (cm *CustomMiddleware) Handle(w http.ResponseWriter, r *http.Request) error {
    // middleware logic
    return nil
}
```

### Adding New Routing Features

1. Extend router in `internal/router/`
2. Add route matching logic
3. Update route configuration
4. Add tests and examples

### Adding Backend Features

1. Extend backend pool in `internal/backend/`
2. Implement load balancing algorithms
3. Add health check logic
4. Update configuration

## Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/proxy

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests

```bash
# Test with Docker backends
cd examples/docker
docker-compose -f 11-basic-nginx-backends.yaml up -d
./test_rate_limit.sh
```

### Benchmarking

```bash
# Run benchmarks
go test -bench=. ./...

# Profile performance
go test -cpuprofile=cpu.prof -memprofile=mem.prof ./internal/proxy
go tool pprof cpu.prof
```

## Debugging

### Logging

The proxy uses standard Go logging. Enable debug logging:

```go
import "log"

log.SetFlags(log.LstdFlags | log.Lshortfile)
log.SetLevel(log.DebugLevel) // if using a logging library
```

### Tracing Requests

Add request IDs and trace headers:

```go
func traceMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = generateID()
        }

        log.Printf("[%s] %s %s", requestID, r.Method, r.URL.Path)
        w.Header().Set("X-Request-ID", requestID)

        next.ServeHTTP(w, r)
    })
}
```

### Health Checks

```bash
# Check proxy health
curl http://localhost:9000/health

# Check backend health
curl http://localhost:3000/health
curl http://localhost:3001/health
```

## Performance Optimization

### Profiling

```bash
# CPU profiling
go build -o bin/proxy ./cmd/proxy
./bin/proxy &
pid=$!
go tool pprof http://localhost:6060/debug/pprof/profile
kill $pid

# Memory profiling
go tool pprof http://localhost:6060/debug/pprof/heap
```

### Benchmarking Tips

- Use `testing.B` for benchmarks
- Reset timers for setup code
- Use `b.ReportAllocs()` to measure allocations
- Compare performance across changes

### Optimization Areas

1. **Connection Pooling**: Reuse HTTP connections
2. **Caching**: Cache frequently accessed data
3. **Concurrent Processing**: Handle requests concurrently
4. **Memory Management**: Minimize allocations
5. **I/O Optimization**: Use buffered I/O where appropriate

## Configuration

### Adding New Config Options

1. Update structs in `internal/config/config.go`
2. Add validation logic
3. Update default configurations
4. Document new options

### Environment Variables

The proxy supports environment variable overrides:

```go
configFile := os.Getenv("CONFIG_FILE")
if configFile == "" {
    configFile = "configs/config.yaml"
}
```

## Deployment

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o proxy ./cmd/proxy

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/proxy .
COPY --from=builder /app/configs ./configs
CMD ["./proxy"]
```

### Kubernetes

Example deployment in `examples/kubernetes/`

### Systemd

```ini
[Unit]
Description=Reverse Proxy
After=network.target

[Service]
Type=simple
User=proxy
WorkingDirectory=/opt/reverse-proxy
ExecStart=/opt/reverse-proxy/bin/proxy -config /etc/proxy/config.yaml
Restart=always

[Install]
WantedBy=multi-user.target
```

## Monitoring

### Metrics

Add Prometheus metrics:

```go
import "github.com/prometheus/client_golang/prometheus"

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "proxy_requests_total",
			Help: "Total number of proxy requests",
		},
		[]string{"method", "status"},
	)
)

func init() {
	prometheus.MustRegister(requestsTotal)
}
```

### Health Endpoints

```go
func healthHandler(w http.ResponseWriter, r *http.Request) {
    // Check dependencies
    if err := checkDatabase(); err != nil {
        http.Error(w, "Database unhealthy", 503)
        return
    }

    w.WriteHeader(200)
    w.Write([]byte("OK"))
}
```

## Troubleshooting

### Common Issues

1. **Port already in use**: Check for running processes on target port
2. **Connection refused**: Ensure backend services are running
3. **Configuration errors**: Validate YAML syntax
4. **Rate limiting**: Check client IP detection logic

### Debug Commands

```bash
# Check running processes
ps aux | grep proxy

# Check network connections
netstat -tlnp | grep :9000

# View logs
docker-compose logs -f

# Test connectivity
curl -v http://localhost:9000/health
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed contribution guidelines.
