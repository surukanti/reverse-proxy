# Reverse Proxy - Advanced Go Implementation

A production-ready, feature-rich reverse proxy for Go supporting 10 different use cases with advanced networking capabilities.

## 🎯 Features

### Core Functionality
- **Path-based routing**: Route requests based on URL paths
- **Subdomain-based routing**: Route based on subdomain (for multi-tenant applications)
- **Header-based routing**: Route based on request headers
- **Regex pattern matching**: Complex route matching with regular expressions
- **Priority-based routing**: Control route matching order with priorities

### Load Balancing & Health
- **Round-robin load balancing**: Distribute requests evenly across backends
- **Health checking**: Automatic health status monitoring of backend servers
- **Failover**: Automatically skip unhealthy servers
- **Connection pooling**: Efficient HTTP connection management

### Middleware & Security
- **Request logging**: Detailed HTTP request/response logging
- **Rate limiting**: Token bucket algorithm for rate limiting
- **CORS support**: Configure CORS headers and policies
- **Authentication proxy**: JWT, API Key, and Basic auth support
- **Middleware chain**: Extensible middleware system

### Advanced Features
- **Response caching**: Cache responses with configurable TTL
- **A/B testing**: Route users to different variants consistently
- **Blue-Green deployments**: Gradual traffic shifting between versions
- **Circuit breaker**: Handle backend failures gracefully
- **Per-tenant rate limiting**: Separate limits for different tenants
- **Event system**: Hook into proxy events for monitoring

## 📋 10 Use Cases Implemented

1. **Microservices API Gateway** - Route requests to different services
2. **Blue-Green Deployments** - Gradually shift traffic between versions
3. **Multi-Tenant SaaS** - Route by subdomain or headers
4. **Development Proxy** - Local development routing
5. **Legacy Migration** - Gradual monolith to microservices migration
6. **CDN Origin Shield** - Origin protection and consolidation
7. **Authentication Proxy** - Add auth to services without built-in auth
8. **Internal Tools** - Consolidate multiple tools behind single URL
9. **A/B Testing** - Route users to different variants
10. **Protocol Translation** - Convert between HTTP versions and protocols

## 🚀 Quick Start

### Building

```bash
# Build the proxy binary
go build -o bin/proxy ./cmd/proxy
```

### Running

```bash
# Run with use case config
./bin/proxy -config examples/1-microservices-gateway.yaml
./bin/proxy -config examples/2-blue-green-deployment.yaml
# ... and so on
```

## 📁 Example Configurations

All 10 use cases have complete configuration files in `examples/`:

- `1-microservices-gateway.yaml` - API Gateway pattern
- `2-blue-green-deployment.yaml` - Deployment strategy
- `3-multi-tenant-saas.yaml` - SaaS routing
- `4-dev-environment-proxy.yaml` - Development proxy
- `5-legacy-migration.yaml` - Gradual migration
- `6-cdn-origin-shield.yaml` - CDN shield
- `7-auth-proxy.yaml` - Authentication proxy
- `8-internal-tools.yaml` - Tool consolidation
- `9-ab-testing.yaml` - A/B testing
- `10-protocol-translation.yaml` - Protocol conversion

## 🔧 Key Configuration Features

- **Path-based routing** with prefix matching
- **Subdomain-based routing** for multi-tenant apps
- **Rate limiting** with token bucket algorithm
- **CORS policies** per route
- **Authentication** (JWT, API Key, Basic)
- **Response caching** with TTL
- **Health checking** for backends
- **Load balancing** across multiple servers

## 📁 Project Structure

```
internal/
├── proxy/          # Core proxy engine, A/B testing, blue-green deployments
├── router/         # Request routing logic
├── backend/        # Backend pool management and health checks
├── middleware/     # Logging, rate limiting, auth, CORS
└── config/         # Configuration loading (YAML/JSON)

cmd/proxy/          # Main entry point
examples/           # 10 configuration examples for each use case
```

## 🔐 Security Features

- JWT/API Key authentication
- CORS policy enforcement
- Rate limiting per client
- TLS/HTTPS support
- Secure header injection
- Per-tenant isolation

## 📊 Monitoring

Event system for:
- Request forwarding
- Cache hits
- Errors and failures
- Rate limit exceeded
- Backend health changes

## 📝 License

MIT License
