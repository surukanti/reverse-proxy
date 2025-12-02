#!/bin/bash
# Docker operations script for reverse proxy

set -e

PROJECT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
IMAGE_NAME="reverse-proxy"
IMAGE_TAG="latest"
CONTAINER_NAME="reverse-proxy"
PORT="9000"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
print_info() {
    echo -e "${GREEN}ℹ${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1" >&2
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Build Docker image
build_image() {
    print_info "Building Docker image: ${IMAGE_NAME}:${IMAGE_TAG}"
    docker build -t ${IMAGE_NAME}:${IMAGE_TAG} ${PROJECT_DIR}
    print_info "Image built successfully"
}

# Run container with Docker
run_container() {
    print_info "Starting container ${CONTAINER_NAME} on port ${PORT}"
    
    # Check if container already running
    if docker ps | grep -q ${CONTAINER_NAME}; then
        print_warning "Container already running. Stopping existing container..."
        docker stop ${CONTAINER_NAME}
        docker rm ${CONTAINER_NAME}
    fi
    
    docker run -d \
        --name ${CONTAINER_NAME} \
        -p ${PORT}:9000 \
        -v ${PROJECT_DIR}/config.yaml:/etc/proxy/config/config.yaml:ro \
        -v ${PROJECT_DIR}/examples:/etc/proxy/examples:ro \
        --restart unless-stopped \
        ${IMAGE_NAME}:${IMAGE_TAG}
    
    print_info "Container started successfully"
    sleep 2
    print_info "Container status:"
    docker ps | grep ${CONTAINER_NAME}
}

# Run with Docker Compose
run_compose() {
    print_info "Starting services with Docker Compose"
    cd ${PROJECT_DIR}
    docker-compose up -d
    print_info "Services started successfully"
    sleep 2
    docker-compose ps
}

# Stop container
stop_container() {
    print_info "Stopping container ${CONTAINER_NAME}"
    docker stop ${CONTAINER_NAME} 2>/dev/null || true
    docker rm ${CONTAINER_NAME} 2>/dev/null || true
    print_info "Container stopped"
}

# Stop compose services
stop_compose() {
    print_info "Stopping Docker Compose services"
    cd ${PROJECT_DIR}
    docker-compose down
    print_info "Services stopped"
}

# View logs
view_logs() {
    print_info "Viewing logs from container ${CONTAINER_NAME}"
    docker logs -f ${CONTAINER_NAME}
}

# View compose logs
view_compose_logs() {
    print_info "Viewing logs from Docker Compose services"
    cd ${PROJECT_DIR}
    docker-compose logs -f
}

# Test container
test_container() {
    print_info "Testing proxy health"
    
    # Wait for container to be ready
    sleep 2
    
    if curl -s http://localhost:${PORT}/health > /dev/null; then
        print_info "✓ Proxy is healthy"
    else
        print_error "Proxy health check failed"
        return 1
    fi
    
    # Test basic request
    print_info "Testing basic request to /get"
    response=$(curl -s http://localhost:${PORT}/get)
    if echo "$response" | grep -q "method"; then
        print_info "✓ Basic request successful"
    else
        print_error "Basic request failed"
        return 1
    fi
    
    # Test CORS
    print_info "Testing CORS preflight"
    status=$(curl -s -o /dev/null -w "%{http_code}" -X OPTIONS http://localhost:${PORT}/get)
    if [ "$status" -eq 200 ]; then
        print_info "✓ CORS preflight successful"
    else
        print_error "CORS preflight failed (status: $status)"
        return 1
    fi
    
    print_info "All tests passed!"
}

# Show stats
show_stats() {
    print_info "Container statistics for ${CONTAINER_NAME}"
    docker stats ${CONTAINER_NAME} --no-stream
}

# Show info
show_info() {
    print_info "Docker image information:"
    docker images | grep ${IMAGE_NAME}
    echo ""
    print_info "Container information:"
    docker inspect ${CONTAINER_NAME} 2>/dev/null | grep -E "\"Image\"|\"State\"|\"Port\"" || true
}

# Clean up
cleanup() {
    print_info "Cleaning up Docker resources"
    docker stop ${CONTAINER_NAME} 2>/dev/null || true
    docker rm ${CONTAINER_NAME} 2>/dev/null || true
    docker rmi ${IMAGE_NAME}:${IMAGE_TAG} 2>/dev/null || true
    print_info "Cleanup complete"
}

# Print usage
print_usage() {
    cat << EOF
Usage: $0 <command> [options]

Commands:
  build              Build Docker image
  run                Run container (standalone)
  run-compose        Run with Docker Compose
  stop               Stop standalone container
  stop-compose       Stop Docker Compose services
  logs               View standalone container logs
  logs-compose       View Docker Compose logs
  test               Test proxy health and routes
  stats              Show container statistics
  info               Show image/container information
  clean              Remove all Docker resources
  help               Show this help message

Examples:
  $0 build
  $0 run
  $0 logs
  $0 test
  $0 stop
  $0 clean

Environment:
  PORT               Port to run proxy on (default: 9000)
  IMAGE_TAG          Docker image tag (default: latest)
  CONTAINER_NAME     Container name (default: reverse-proxy)

EOF
}

# Main
main() {
    case "${1:-help}" in
        build)
            build_image
            ;;
        run)
            build_image
            run_container
            ;;
        run-compose)
            run_compose
            ;;
        stop)
            stop_container
            ;;
        stop-compose)
            stop_compose
            ;;
        logs)
            view_logs
            ;;
        logs-compose)
            view_compose_logs
            ;;
        test)
            test_container
            ;;
        stats)
            show_stats
            ;;
        info)
            show_info
            ;;
        clean)
            cleanup
            ;;
        help)
            print_usage
            ;;
        *)
            print_error "Unknown command: $1"
            print_usage
            exit 1
            ;;
    esac
}

main "$@"
