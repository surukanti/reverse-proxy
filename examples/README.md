# Examples

This directory contains example configurations and setups demonstrating various use cases for the reverse proxy.

## Directory Structure

```
examples/
├── configs/              # YAML configuration examples
│   ├── 1-microservices-gateway.yaml
│   ├── 2-blue-green-deployment.yaml
│   ├── 3-multi-tenant-saas.yaml
│   ├── 4-dev-environment-proxy.yaml
│   ├── 5-legacy-migration.yaml
│   ├── 6-cdn-origin-shield.yaml
│   ├── 7-auth-proxy.yaml
│   ├── 8-internal-tools.yaml
│   ├── 9-ab-testing.yaml
│   └── 10-protocol-translation.yaml
└── docker/               # Docker-based examples
    ├── 11-basic-nginx-backends.yaml
    ├── nginx.conf
    ├── nginx1.conf
    ├── nginx2.conf
    └── README-nginx-backends.md
```

## Configuration Examples

Each numbered YAML file demonstrates a specific use case:

### 1. Microservices API Gateway
Routes requests to multiple microservices based on URL paths.

```bash
./bin/proxy -config examples/configs/1-microservices-gateway.yaml
```

### 2. Blue-Green Deployment
Gradually shifts traffic between blue and green deployments.

```bash
./bin/proxy -config examples/configs/2-blue-green-deployment.yaml
```

### 3. Multi-Tenant SaaS
Routes based on subdomain for multi-tenant applications.

```bash
./bin/proxy -config examples/configs/3-multi-tenant-saas.yaml
```

### 4. Development Environment Proxy
Local development routing setup.

```bash
./bin/proxy -config examples/configs/4-dev-environment-proxy.yaml
```

### 5. Legacy Migration
Gradual migration from monolithic to microservices architecture.

```bash
./bin/proxy -config examples/configs/5-legacy-migration.yaml
```

### 6. CDN Origin Shield
Origin protection and request consolidation for CDN setups.

```bash
./bin/proxy -config examples/configs/6-cdn-origin-shield.yaml
```

### 7. Authentication Proxy
Adds authentication to services without built-in auth.

```bash
./bin/proxy -config examples/configs/7-auth-proxy.yaml
```

### 8. Internal Tools
Consolidates multiple internal tools behind a single URL.

```bash
./bin/proxy -config examples/configs/8-internal-tools.yaml
```

### 9. A/B Testing
Routes users to different variants for A/B testing.

```bash
./bin/proxy -config examples/configs/9-ab-testing.yaml
```

### 10. Protocol Translation
Converts between different HTTP protocols and versions.

```bash
./bin/proxy -config examples/configs/10-protocol-translation.yaml
```

## Docker Examples

### 11. Basic Nginx Backends
Complete Docker setup with reverse proxy and nginx backend services.

```bash
cd examples/docker
docker-compose -f 11-basic-nginx-backends.yaml up --build
```

This example includes:
- Reverse proxy on port 9000
- Two nginx backend services on ports 3000 and 3001
- Health checking and load balancing
- Rate limiting demonstration

See [docker/README-nginx-backends.md](docker/README-nginx-backends.md) for detailed instructions.

## Using Examples

### 1. Standalone Binary

```bash
# Build the proxy
make build

# Run with example config
./bin/proxy -config examples/configs/1-microservices-gateway.yaml
```

### 2. Docker Container

```bash
# Build Docker image
make build-docker

# Run with example config
docker run -d \
  -p 9000:9000 \
  -v $(pwd)/examples:/etc/proxy/examples:ro \
  -v $(pwd)/configs/config.yaml:/etc/proxy/config/config.yaml:ro \
  reverse-proxy:latest
```

### 3. Docker Compose

For the full docker example:

```bash
cd examples/docker
docker-compose -f 11-basic-nginx-backends.yaml up --build
```

## Configuration Structure

All example configurations follow this structure:

```yaml
server:
  host: "0.0.0.0"
  port: "9000"
  tls: false

backends:
  - id: "backend-name"
    servers:
      - "http://service1:port"
      - "http://service2:port"
    health_check:
      enabled: true
      path: "/health"

routes:
  - name: "route-name"
    pattern: "/api/.*"
    backend_id: "backend-name"

policies:
  rate_limit:
    enabled: true
    max_requests: 1000
    window: "1m"
  cors:
    enabled: true
  auth:
    enabled: false
```

## Customization

### Modifying Examples

1. Copy an example config to `configs/`
2. Modify the configuration for your needs
3. Test with `./bin/proxy -config configs/your-config.yaml`

### Adding New Examples

1. Create a new YAML file in `examples/configs/`
2. Follow the naming convention: `N-description.yaml`
3. Update this README with the new example
4. Test thoroughly before committing

### Docker Examples

1. Create Docker Compose files in `examples/docker/`
2. Include all necessary configuration files
3. Provide clear documentation in a README
4. Test the complete setup

## Testing Examples

### Health Checks

Test that backends are healthy:

```bash
curl http://localhost:9000/health
```

### Load Balancing

Verify requests are distributed:

```bash
for i in {1..10}; do curl -s http://localhost:9000/api/test; done
```

### Rate Limiting

Test rate limiting behavior:

```bash
# Run the rate limiting test
./tools/test_rate_limit.sh
```

## Troubleshooting

### Common Issues

1. **Port conflicts**: Ensure target ports are available
2. **Backend connectivity**: Verify backend services are running
3. **Configuration syntax**: Validate YAML syntax
4. **Path matching**: Check route patterns and priorities

### Debugging

Enable debug logging:

```bash
# Add logging configuration
policies:
  logging:
    level: "debug"
```

### Logs

View proxy logs:

```bash
# Standalone
tail -f /tmp/reverse-proxy.log

# Docker
make logs-docker

# Docker Compose
make logs-compose
```

## Contributing

When adding new examples:

1. Follow existing naming conventions
2. Include comprehensive documentation
3. Test all functionality
4. Update this README
5. Ensure examples work with current codebase

See [../docs/CONTRIBUTING.md](../docs/CONTRIBUTING.md) for detailed contribution guidelines.
