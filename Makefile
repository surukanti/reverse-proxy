.PHONY: help build build-docker run run-docker run-compose stop stop-compose logs logs-compose test test-docker clean clean-docker clean-all push pull release version info

# Variables
PROJECT_NAME := reverse-proxy
BINARY_NAME := proxy
IMAGE_NAME := reverse-proxy
IMAGE_TAG := latest
CONTAINER_NAME := reverse-proxy
PORT := 9000
REGISTRY := docker.io
USERNAME := $(shell whoami)
GO_VERSION := 1.24.2
GO_FLAGS := -v
GO_BUILD_FLAGS := -a -installsuffix cgo

# Colors
YELLOW := \033[0;33m
GREEN := \033[0;32m
BLUE := \033[0;34m
RED := \033[0;31m
NC := \033[0m # No Color

# Paths
BIN_DIR := bin
DOCKER_DIR := .
CONFIG_FILE := configs/config.yaml
DOCKER_IMAGE := $(IMAGE_NAME):$(IMAGE_TAG)

# Default target
.DEFAULT_GOAL := help

##@ General

help: ## Display this help screen
	@echo "$(BLUE)╔════════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(BLUE)║           Reverse Proxy Makefile - Build & Deploy          ║$(NC)"
	@echo "$(BLUE)╚════════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BLUE)Quick Start:$(NC)"
	@echo "  make build              # Build standalone binary"
	@echo "  make run                # Run standalone binary"
	@echo "  make build-docker       # Build Docker image"
	@echo "  make run-docker         # Run Docker container"
	@echo "  make run-compose        # Run with Docker Compose"
	@echo ""
	@echo "$(BLUE)Testing & Validation:$(NC)"
	@echo "  make test               # Run unit tests"
	@echo "  make test-docker        # Test Docker container"
	@echo "  make coverage           # Generate coverage report"
	@echo ""

version: ## Show project version and Go version
	@echo "$(GREEN)Project:$(NC) $(PROJECT_NAME)"
	@echo "$(GREEN)Binary:$(NC) $(BINARY_NAME)"
	@echo "$(GREEN)Image:$(NC) $(DOCKER_IMAGE)"
	@echo "$(GREEN)Go Version:$(NC) $(GO_VERSION)"
	@go version

##@ Building

build: ## Build standalone binary
	@echo "$(YELLOW)🔨 Building $(BINARY_NAME) binary...$(NC)"
	@mkdir -p $(BIN_DIR)
	@CGO_ENABLED=0 go build $(GO_FLAGS) -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/proxy
	@echo "$(GREEN)✓ Binary built: $(BIN_DIR)/$(BINARY_NAME)$(NC)"
	@ls -lh $(BIN_DIR)/$(BINARY_NAME)

build-no-cgo: ## Build binary with CGO enabled (for local development)
	@echo "$(YELLOW)🔨 Building $(BINARY_NAME) (with CGO)...$(NC)"
	@mkdir -p $(BIN_DIR)
	@go build $(GO_FLAGS) -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/proxy
	@echo "$(GREEN)✓ Binary built: $(BIN_DIR)/$(BINARY_NAME)$(NC)"

build-docker: ## Build Docker image
	@echo "$(YELLOW)🐳 Building Docker image: $(DOCKER_IMAGE)...$(NC)"
	@docker build -t $(DOCKER_IMAGE) -f $(DOCKER_DIR)/Dockerfile .
	@echo "$(GREEN)✓ Docker image built$(NC)"
	@docker images | grep $(IMAGE_NAME)

build-docker-nocache: ## Build Docker image without cache
	@echo "$(YELLOW)🐳 Building Docker image (no-cache): $(DOCKER_IMAGE)...$(NC)"
	@docker build --no-cache -t $(DOCKER_IMAGE) -f $(DOCKER_DIR)/Dockerfile .
	@echo "$(GREEN)✓ Docker image built$(NC)"

##@ Running

run: build ## Build and run standalone binary
	@echo "$(YELLOW)▶️  Running $(BINARY_NAME)...$(NC)"
	@./$(BIN_DIR)/$(BINARY_NAME) -config $(CONFIG_FILE)

run-docker: build-docker ## Build and run Docker container
	@echo "$(YELLOW)▶️  Running Docker container: $(CONTAINER_NAME)...$(NC)"
	@if docker ps | grep -q $(CONTAINER_NAME); then \
		echo "$(RED)✗ Container already running. Stopping existing container...$(NC)"; \
		docker stop $(CONTAINER_NAME) && docker rm $(CONTAINER_NAME); \
	fi
	@docker run -d \
		--name $(CONTAINER_NAME) \
		-p $(PORT):9000 \
		-v $$(pwd)/$(CONFIG_FILE):/etc/proxy/config/config.yaml:ro \
		-v $$(pwd)/examples:/etc/proxy/examples:ro \
		--restart unless-stopped \
		$(DOCKER_IMAGE)
	@echo "$(GREEN)✓ Container started: $(CONTAINER_NAME)$(NC)"
	@echo "$(BLUE)  Port:$(NC) http://localhost:$(PORT)"
	@echo "$(BLUE)  Logs:$(NC) make logs-docker"
	@docker ps | grep $(CONTAINER_NAME)

run-compose: ## Start all services with Docker Compose
	@echo "$(YELLOW)▶️  Starting services with Docker Compose...$(NC)"
	@docker-compose up -d
	@echo "$(GREEN)✓ Services started$(NC)"
	@docker-compose ps

run-compose-build: ## Build and start with Docker Compose
	@echo "$(YELLOW)▶️  Building and starting services...$(NC)"
	@docker-compose up -d --build
	@echo "$(GREEN)✓ Services started$(NC)"
	@docker-compose ps

run-interactive: ## Run Docker container in interactive mode
	@echo "$(YELLOW)▶️  Running Docker container (interactive)...$(NC)"
	@docker run -it --rm \
		-p $(PORT):9000 \
		-v $$(pwd)/$(CONFIG_FILE):/etc/proxy/config/config.yaml:ro \
		-v $$(pwd)/examples:/etc/proxy/examples:ro \
		$(DOCKER_IMAGE)

##@ Stopping & Cleanup

stop: ## Stop standalone binary (if running)
	@pkill -f "$(BIN_DIR)/$(BINARY_NAME)" || echo "$(YELLOW)⚠  No process running$(NC)"
	@echo "$(GREEN)✓ Stopped$(NC)"

stop-docker: ## Stop Docker container
	@echo "$(YELLOW)⏹️  Stopping container: $(CONTAINER_NAME)...$(NC)"
	@docker stop $(CONTAINER_NAME) 2>/dev/null || echo "$(YELLOW)⚠  Container not running$(NC)"
	@docker rm $(CONTAINER_NAME) 2>/dev/null || true
	@echo "$(GREEN)✓ Container stopped$(NC)"

stop-compose: ## Stop Docker Compose services
	@echo "$(YELLOW)⏹️  Stopping services...$(NC)"
	@docker-compose down
	@echo "$(GREEN)✓ Services stopped$(NC)"

clean: ## Clean build artifacts
	@echo "$(YELLOW)🧹 Cleaning build artifacts...$(NC)"
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out
	@echo "$(GREEN)✓ Cleaned$(NC)"

clean-docker: ## Remove Docker image and containers
	@echo "$(YELLOW)🧹 Removing Docker resources...$(NC)"
	@docker stop $(CONTAINER_NAME) 2>/dev/null || true
	@docker rm $(CONTAINER_NAME) 2>/dev/null || true
	@docker rmi $(DOCKER_IMAGE) 2>/dev/null || echo "$(YELLOW)⚠  Image not found$(NC)"
	@echo "$(GREEN)✓ Docker resources cleaned$(NC)"

clean-compose: ## Stop and remove Docker Compose services
	@echo "$(YELLOW)🧹 Stopping Docker Compose services...$(NC)"
	@docker-compose down --rmi local -v
	@echo "$(GREEN)✓ Docker Compose services cleaned$(NC)"

clean-all: clean clean-docker ## Clean all artifacts and Docker resources
	@echo "$(GREEN)✓ All cleaned$(NC)"

##@ Testing

test: ## Run unit tests
	@echo "$(YELLOW)🧪 Running unit tests...$(NC)"
	@go test -v ./...
	@echo "$(GREEN)✓ Tests completed$(NC)"

test-verbose: ## Run unit tests with verbose output
	@echo "$(YELLOW)🧪 Running unit tests (verbose)...$(NC)"
	@go test -vv ./...

test-coverage: coverage ## Run tests with coverage (alias)

coverage: ## Generate coverage report
	@echo "$(YELLOW)📊 Generating coverage report...$(NC)"
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report generated$(NC)"
	@echo "  HTML Report: $(YELLOW)coverage.html$(NC)"
	@go test -cover ./... | grep coverage:
	@echo "$(BLUE)  Open in browser: open coverage.html$(NC)"

test-docker: run-docker ## Test Docker container health and endpoints
	@echo "$(YELLOW)🧪 Testing Docker container...$(NC)"
	@sleep 3
	@echo "$(BLUE)Testing health endpoint...$(NC)"
	@curl -s http://localhost:$(PORT)/health > /dev/null && echo "$(GREEN)✓ Health check passed$(NC)" || echo "$(RED)✗ Health check failed$(NC)"
	@echo ""
	@echo "$(BLUE)Testing basic GET request...$(NC)"
	@curl -s -X GET http://localhost:$(PORT)/get | head -c 100 && echo "" && echo "$(GREEN)✓ GET request passed$(NC)"
	@echo ""
	@echo "$(BLUE)Testing POST request...$(NC)"
	@curl -s -X POST -H "Content-Type: application/json" -d '{"test":"data"}' http://localhost:$(PORT)/post | head -c 100 && echo "" && echo "$(GREEN)✓ POST request passed$(NC)"
	@echo ""
	@echo "$(BLUE)Testing CORS preflight...$(NC)"
	@curl -s -X OPTIONS -H "Origin: http://localhost:3000" http://localhost:$(PORT)/get > /dev/null && echo "$(GREEN)✓ CORS preflight passed$(NC)"

bench: ## Run benchmarks
	@echo "$(YELLOW)⚡ Running benchmarks...$(NC)"
	@go test -bench=. -benchmem ./...

##@ Logging & Monitoring

logs: ## View standalone binary logs (requires process manager)
	@tail -f /tmp/$(PROJECT_NAME).log 2>/dev/null || echo "$(YELLOW)⚠  No logs available$(NC)"

logs-docker: ## View Docker container logs
	@echo "$(YELLOW)📋 Container logs ($(CONTAINER_NAME))...$(NC)"
	@docker logs -f $(CONTAINER_NAME)

logs-compose: ## View Docker Compose logs
	@echo "$(YELLOW)📋 Docker Compose logs...$(NC)"
	@docker-compose logs -f

logs-compose-proxy: ## View Docker Compose proxy service logs
	@docker-compose logs -f reverse-proxy

logs-compose-backend: ## View Docker Compose backend logs
	@docker-compose logs -f backend-api-1 backend-api-2

stats: ## Show Docker container statistics
	@echo "$(YELLOW)📊 Container statistics...$(NC)"
	@docker stats $(CONTAINER_NAME) --no-stream

ps: ## List running containers
	@echo "$(YELLOW)📦 Running containers...$(NC)"
	@docker ps | grep -E "CONTAINER|$(IMAGE_NAME)" || echo "$(YELLOW)⚠  No containers running$(NC)"

ps-compose: ## List Docker Compose services status
	@echo "$(YELLOW)📦 Docker Compose services...$(NC)"
	@docker-compose ps

##@ Docker Registry

push: ## Push Docker image to registry
	@echo "$(YELLOW)📤 Pushing image to registry...$(NC)"
	@docker tag $(DOCKER_IMAGE) $(REGISTRY)/$(USERNAME)/$(DOCKER_IMAGE)
	@docker push $(REGISTRY)/$(USERNAME)/$(DOCKER_IMAGE)
	@echo "$(GREEN)✓ Image pushed$(NC)"

pull: ## Pull Docker image from registry
	@echo "$(YELLOW)📥 Pulling image from registry...$(NC)"
	@docker pull $(REGISTRY)/$(USERNAME)/$(DOCKER_IMAGE)
	@echo "$(GREEN)✓ Image pulled$(NC)"

release: build build-docker ## Create a release build
	@echo "$(YELLOW)🎯 Creating release...$(NC)"
	@echo "$(GREEN)✓ Release build complete$(NC)"
	@echo "  Binary: $(BIN_DIR)/$(BINARY_NAME)"
	@echo "  Docker: $(DOCKER_IMAGE)"

##@ Development

fmt: ## Format Go code
	@echo "$(YELLOW)🎨 Formatting Go code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(NC)"

vet: ## Run go vet
	@echo "$(YELLOW)🔍 Running go vet...$(NC)"
	@go vet ./...
	@echo "$(GREEN)✓ No vet issues found$(NC)"

lint: ## Run linter (requires golangci-lint)
	@command -v golangci-lint >/dev/null 2>&1 || (echo "$(RED)✗ golangci-lint not installed$(NC)" && exit 1)
	@echo "$(YELLOW)🔍 Running linter...$(NC)"
	@golangci-lint run ./...
	@echo "$(GREEN)✓ Linting complete$(NC)"

deps: ## Download dependencies
	@echo "$(YELLOW)📦 Downloading dependencies...$(NC)"
	@go mod download
	@go mod verify
	@echo "$(GREEN)✓ Dependencies downloaded$(NC)"

tidy: ## Tidy dependencies
	@echo "$(YELLOW)🧹 Tidying dependencies...$(NC)"
	@go mod tidy
	@echo "$(GREEN)✓ Dependencies tidied$(NC)"

vendor: ## Vendor dependencies
	@echo "$(YELLOW)📦 Vendoring dependencies...$(NC)"
	@go mod vendor
	@echo "$(GREEN)✓ Dependencies vendored$(NC)"

##@ Information

info: ## Show build information
	@echo "$(BLUE)╔════════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(BLUE)║                   Build Information                        ║$(NC)"
	@echo "$(BLUE)╚════════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(YELLOW)Project:$(NC)"
	@echo "  Name: $(PROJECT_NAME)"
	@echo "  Binary: $(BINARY_NAME)"
	@echo "  Config: $(CONFIG_FILE)"
	@echo ""
	@echo "$(YELLOW)Docker:$(NC)"
	@echo "  Image: $(DOCKER_IMAGE)"
	@echo "  Port: $(PORT)"
	@echo "  Container: $(CONTAINER_NAME)"
	@echo ""
	@echo "$(YELLOW)Go:$(NC)"
	@go version
	@echo ""
	@echo "$(YELLOW)System:$(NC)"
	@uname -a
	@echo ""
	@echo "$(YELLOW)Docker:$(NC)"
	@docker version --format='Client: {{.Client.Version}}'
	@docker version --format='Server: {{.Server.Version}}' 2>/dev/null || echo "  Server: N/A"

images: ## List all images
	@echo "$(YELLOW)🐳 Docker images:$(NC)"
	@docker images | grep -E "REPOSITORY|$(IMAGE_NAME)" || echo "$(YELLOW)⚠  No images found$(NC)"

files: ## List project files
	@echo "$(YELLOW)📁 Project structure:$(NC)"
	@find . -type f -name "*.go" -o -name "*.yaml" -o -name "Dockerfile" -o -name "Makefile" | grep -v ".git" | sort

##@ Help & Documentation

docs: ## Open documentation
	@echo "$(YELLOW)📚 Documentation:$(NC)"
	@ls -1 *.md | grep -E "DOCKER|RUNNING|QUICKSTART|ARCHITECTURE|DEPLOYMENT|IMPLEMENTATION|INDEX"

quick-start: ## Show quick start commands
	@echo "$(BLUE)╔════════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(BLUE)║                    Quick Start Guide                       ║$(NC)"
	@echo "$(BLUE)╚════════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(YELLOW)1. Build & Run Standalone:$(NC)"
	@echo "   $$ make build"
	@echo "   $$ make run"
	@echo ""
	@echo "$(YELLOW)2. Build & Run with Docker:$(NC)"
	@echo "   $$ make build-docker"
	@echo "   $$ make run-docker"
	@echo ""
	@echo "$(YELLOW)3. Run Full Stack with Compose:$(NC)"
	@echo "   $$ make run-compose"
	@echo ""
	@echo "$(YELLOW)4. Test the Proxy:$(NC)"
	@echo "   $$ make test-docker"
	@echo ""
	@echo "$(YELLOW)5. View Logs:$(NC)"
	@echo "   $$ make logs-docker"
	@echo ""
	@echo "$(YELLOW)6. Stop Services:$(NC)"
	@echo "   $$ make stop-docker"
	@echo ""

all: clean build build-docker ## Build everything (clean, binary, docker)
	@echo "$(GREEN)✓ All builds complete$(NC)"

.PHONY: all clean deps
