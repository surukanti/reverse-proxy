# Basic Nginx Backend Example

This example demonstrates how to set up the reverse proxy with simple nginx backend services for testing and development.

## Files

- `11-basic-nginx-backends.yaml` - Docker Compose configuration with reverse proxy and two nginx backend services
- `nginx1.conf` - Nginx configuration for backend-api-1
- `nginx2.conf` - Nginx configuration for backend-api-2

## Usage

1. Start the services:
   ```bash
   cd examples
   docker-compose -f 11-basic-nginx-backends.yaml up --build -d
   ```

2. Test the reverse proxy:
   ```bash
   curl http://localhost:9000/
   ```

3. Test load balancing:
   ```bash
   curl http://localhost:9000/get
   ```

The reverse proxy will load balance requests between backend-api-1 and backend-api-2.

## Health Checks

Both backend services expose health check endpoints:
- `/get` - Returns "OK from backend-X"
- `/health` - Returns "OK from backend-X"

The reverse proxy performs health checks every 30 seconds on the `/get` endpoint.
