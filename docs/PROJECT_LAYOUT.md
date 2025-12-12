# Project Layout

This document describes the directory structure and organization of the reverse proxy project.

## Root Directory Structure

```
reverse-proxy/
├── cmd/                    # Application entrypoints
│   └── proxy/             # Main proxy application
│       └── main.go
├── internal/              # Private application code
│   ├── advanced/         # Advanced proxy features
│   ├── backend/          # Backend server management
│   ├── config/           # Configuration handling
│   ├── middleware/       # HTTP middleware components
│   ├── proxy/            # Core proxy logic
│   └── router/           # Routing logic
├── configs/              # Main configuration files
│   ├── config.yaml       # Default configuration
│   └── config-simple.yaml # Simplified configuration
├── examples/             # Example configurations and setups
│   ├── configs/          # Example YAML configurations
│   │   ├── 1-microservices-gateway.yaml
│   │   ├── 2-blue-green-deployment.yaml
│   │   └── ...
│   └── docker/           # Docker-based examples
│       ├── 11-basic-nginx-backends.yaml
│       ├── nginx*.conf
│       └── README-nginx-backends.md
├── tools/                # Development tools and scripts
│   └── test_rate_limit.sh # Rate limiting test utility
├── docs/                 # Documentation
│   ├── CONTRIBUTING.md   # Contribution guidelines
│   ├── DEVELOPMENT.md    # Development guide
│   └── PROJECT_LAYOUT.md # This file
├── bin/                  # Compiled binaries (generated)
├── docker-compose.yml    # Main Docker setup
├── Dockerfile           # Docker build instructions
├── Makefile             # Build automation
├── go.mod               # Go module definition
├── go.sum               # Go dependencies
├── .gitignore           # Git ignore rules
├── README.md            # Project README
├── LICENSE              # License file
└── ARCHITECTURE.md      # Architecture documentation
```

## Directory Descriptions

### `cmd/`
Contains application entrypoints. Each subdirectory represents a separate binary.

- `cmd/proxy/` - Main reverse proxy application

### `internal/`
Private application code that should not be imported by external applications.

- `internal/advanced/` - Advanced proxy features and extensions
- `internal/backend/` - Backend server pool management and health checking
- `internal/config/` - Configuration loading and validation
- `internal/middleware/` - HTTP middleware (rate limiting, auth, CORS, etc.)
- `internal/proxy/` - Core proxy logic and request handling
- `internal/router/` - Route matching and routing logic

### `configs/`
Main configuration files used by the application.

- Production-ready configurations
- Default settings
- Base configurations that can be extended

### `examples/`
Example configurations and setups for different use cases.

- `examples/configs/` - YAML configuration examples for different scenarios
- `examples/docker/` - Docker Compose setups with example backend services

### `scripts/`
Utility scripts for development, testing, and deployment.

### `docs/`
Additional documentation beyond the main README.

### Generated/Output Directories

- `bin/` - Compiled binaries (created by `make build` or `go build`)

## File Naming Conventions

- Configuration files: `*.yaml` or `*.yml`
- Go source files: `*.go`
- Test files: `*_test.go`
- Documentation: `*.md`
- Scripts: `*.sh`

## Package Organization

- Each internal package has a single responsibility
- Packages are organized by domain/feature
- Clear separation between core logic and external interfaces
- Test files are co-located with source files

## Configuration Hierarchy

1. `configs/config.yaml` - Main default configuration
2. `examples/configs/*.yaml` - Use case specific examples
3. User-provided configuration files (via `-config` flag)

## Development Workflow

- Place new features in appropriate `internal/` packages
- Add tests alongside source code
- Update documentation when changing public APIs
- Use examples to demonstrate new functionality

## Import Guidelines

- Internal packages should only import from other internal packages or standard library
- External dependencies should be minimal and well-vetted
- Prefer standard library solutions when possible
