# Architecture Guide

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Client Requests                         │
│              (HTTP/HTTPS/WebSocket)                         │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                  Reverse Proxy (Main)                       │
│  Port: 8080 (configurable)                                  │
│  - Request parsing                                          │
│  - Route matching                                           │
└────────────────┬────────────────────────────────────────────┘
                 │
        ┌────────┴─────────┬──────────────┬──────────────┐
        │                  │              │              │
        ▼                  ▼              ▼              ▼
    ┌────────┐    ┌──────────────┐   ┌────────┐   ┌─────────┐
    │ Router │    │ Middleware   │   │ Cache  │   │ Events  │
    │        │    │ Chain        │   │ Layer  │   │ System  │
    └────┬───┘    └──────────────┘   └────────┘   └─────────┘
         │
         └─────────────────┬─────────────────────────────────┐
                           │                                 │
                    ┌──────▼──────┐              ┌──────────▼──┐
                    │  Backend    │              │ Advanced    │
                    │  Pool       │              │ Features    │
                    │  Manager    │              │             │
                    └──────┬──────┘              └─────────────┘
                           │
         ┌─────────────────┼─────────────────┬────────────────┐
         │                 │                 │                │
         ▼                 ▼                 ▼                ▼
    ┌─────────┐   ┌──────────────┐   ┌────────────┐   ┌──────────┐
    │ Health  │   │ Load         │   │ Failover   │   │ Rate     │
    │ Checker │   │ Balancing    │   │ Logic      │   │ Limiting │
    └─────────┘   └──────────────┘   └────────────┘   └──────────┘
         │
         └─────────────────┬───────────────────────────────────┐
                           │                                   │
                    ┌──────▼────────────────┐        ┌─────────▼──┐
                    │ Backend Servers      │        │ Monitoring │
                    │                      │        │ & Metrics  │
                    │ • Health checking    │        │            │
                    │ • Connection pooling │        └────────────┘
                    │ • Load distribution  │
                    └─────────────────────┘
                           │
    ┌──────────────────────┼──────────────────────┐
    │                      │                      │
    ▼                      ▼                      ▼
┌────────┐           ┌────────┐           ┌────────┐
│Service │           │Service │           │Service │
│  1     │           │  2     │           │  N     │
└────────┘           └────────┘           └────────┘
```

## Component Details

### 1. Router (`internal/router/router.go`)
- **Purpose**: Match incoming requests to routes
- **Features**:
  - Path-based matching
  - Subdomain-based matching
  - Header-based matching
  - Regex pattern matching
  - Priority-based selection

**Flow**:
```
Request → Check Methods → Check Subdomain → Check Headers → 
Check Path → Check Pattern → Return Matching Route
```

### 2. Backend Pool (`internal/backend/pool.go`)
- **Purpose**: Manage backend servers
- **Features**:
  - Round-robin load balancing
  - Health status tracking
  - Server pooling
  - Automatic failover

**Flow**:
```
Get Server → Filter Healthy → Round-robin Selection → Return Server
```

### 3. Health Checker (`internal/backend/pool.go`)
- **Purpose**: Monitor backend health
- **Features**:
  - Periodic health checks
  - Configurable intervals
  - Status updates
  - Automatic recovery detection

**Flow**:
```
Interval Tick → Check Each Server → GET /health → 
Status Code 200? → Update Status → Next Interval
```

### 4. Middleware Chain (`internal/middleware/middleware.go`)
- **Purpose**: Process requests through middleware
- **Chain Order**:
  1. Logging middleware
  2. CORS middleware
  3. Auth middleware
  4. Rate limiting (in handler)

**Features**:
- Extensible chain
- Request/response modification
- Early termination on error

### 5. Proxy Engine (`internal/proxy/proxy.go`)
- **Purpose**: Main request handling
- **Features**:
  - Route matching
  - Middleware execution
  - Backend selection
  - Request forwarding
  - Response caching
  - Event emission

**Request Flow**:
```
1. Rate Limit Check
   ├─ Exceeded? → Return 429
   └─ Allowed? → Continue

2. Middleware Chain
   ├─ Logging
   ├─ CORS
   ├─ Auth
   └─ Error? → Return Error

3. Route Matching
   ├─ Found? → Continue
   └─ Not Found? → Return 404

4. Backend Selection
   ├─ Available? → Continue
   └─ None Available? → Return 503

5. Cache Check
   ├─ Cache Hit? → Return Cached
   └─ Cache Miss? → Continue

6. Forward Request
   ├─ Forward to Backend
   ├─ Get Response
   ├─ Cache Response (if applicable)
   └─ Return to Client

7. Event Emission
   └─ Emit appropriate events
```

### 6. Advanced Features (`internal/proxy/advanced.go`)

#### A/B Testing Manager
```
User Request → Hash User ID → Calculate Variant Percent →
Select Variant A or B → Route Request → Track Metrics
```

#### Blue-Green Manager
```
Traffic Shift Start → Calculate Shift Percent →
Route Percent to New Version → Monitor Errors →
Complete Shift → Activate New Version
```

#### Circuit Breaker
```
Closed (Normal) → Failure Threshold Reached → Open (Failed) →
Wait Timeout → Half-Open (Test) → Success? → Closed or Open
```

### 7. Configuration (`internal/config/config.go`)
- **Purpose**: Load and manage configuration
- **Formats**:
  - YAML (primary)
  - JSON (alternative)

**Structure**:
```
Config
├── Server (Host, Port, TLS)
├── Backends (List of backend pools)
├── Routes (Routing rules)
└── Policies (Rate limit, CORS, Auth, Cache)
```

## Data Flow Examples

### Example 1: Microservices Gateway
```
Client Request: GET /users/123

1. Router matches /users → users-service backend
2. Backend pool selects healthy user-service instance
3. Request forwarded to: http://users-service:3001/users/123
4. Response cached (if GET)
5. Response returned to client
6. Event: "request_forwarded" emitted
```

### Example 2: Multi-Tenant Request
```
Client Request: GET / (from tenant1.app.com)

1. Router extracts subdomain: "tenant1"
2. Route matches subdomain: tenant1-route
3. Backend pool selected: tenant1-backend
4. Rate limiter checks per-tenant limit
5. Request forwarded to tenant1 service
6. Response returned
```

### Example 3: Blue-Green Deployment
```
Client Request: GET /api/data

1. Blue-Green Manager checks traffic shift (20% to green)
2. Client hash: 45 (out of 100)
3. 45 > 20 → Route to Blue version
4. Request forwarded to blue backend
5. After 100%, all requests go to green
```

### Example 4: Rate Limiting
```
Client Request: GET /api/resource

1. Rate Limiter checks client IP: 10.0.0.1
2. Bucket for 10.0.0.1: 5 tokens remaining
3. Request uses 1 token
4. 4 tokens remaining
5. Request allowed → Continue

Later:
1. Bucket: 0 tokens
2. Next request arrives
3. 429 Too Many Requests returned
```

## Concurrency Model

### Thread Safety
- **Router**: RWMutex for route list
- **Backend Pool**: Atomic operations for health status
- **Rate Limiter**: Map with per-client locks
- **Cache**: RWMutex for cache map
- **Event System**: Goroutines for async event handling

### Goroutines
1. **Health Checker**: Background goroutine checking health
2. **Event Handlers**: Goroutines for each event listener
3. **Request Forwarding**: Each request in separate goroutine
4. **Cache Cleanup**: Background goroutine (optional)

## Performance Characteristics

### Time Complexity
- Route Matching: O(n) where n = number of routes
- Backend Selection: O(1) round-robin with filtering
- Rate Limit Check: O(1) map lookup + bucket calculation
- Cache Lookup: O(1) map access

### Space Complexity
- Cache: O(c) where c = number of cached entries
- Connection Pool: O(p) where p = max connections
- Backend Servers: O(b) where b = number of backends

## Deployment Patterns

### Single Instance
```
Client → Proxy (Port 8080) → Multiple Backends
```

### Load-Balanced Proxies
```
Client → LB → Proxy 1, Proxy 2, Proxy 3 → Backends
```

### Chained Proxies
```
Client → Proxy 1 → Proxy 2 → Backend
(For additional filtering/transformation)
```

### Proxy Network
```
     ┌─→ Proxy (Users)
Client → LB
     └─→ Proxy (Orders)
     └─→ Proxy (Products)
```

## Security Architecture

```
┌──────────────┐
│   Request    │
└──────┬───────┘
       │
       ▼
┌──────────────────┐
│ Rate Limit       │
│ (IP-based)       │
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│ CORS Policy      │
│ (Origin check)   │
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│ Auth Validation  │
│ (JWT/API Key)    │
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│ Route to Backend │
│ (Trusted)        │
└──────────────────┘
```

---

See [README.md](README.md) for detailed documentation.
