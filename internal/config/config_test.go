package config

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestLoadFromYAML(t *testing.T) {
	yaml := `
server:
  host: localhost
  port: "8080"
  tls: false
routes:
  - name: api
    pattern: /api
    methods:
      - GET
      - POST
    backend_id: backend1
backends:
  - id: backend1
    servers:
      - http://localhost:3000
    load_balancing: round_robin
`

	tmpfile, err := ioutil.TempFile("", "config*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(yaml); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	cfg, err := LoadFromYAML(tmpfile.Name())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg == nil {
		t.Fatal("expected config to be non-nil")
	}

	if cfg.Server.Host != "localhost" {
		t.Errorf("expected host localhost, got %s", cfg.Server.Host)
	}

	if cfg.Server.Port != "8080" {
		t.Errorf("expected port 8080, got %s", cfg.Server.Port)
	}

	if len(cfg.Routes) != 1 {
		t.Errorf("expected 1 route, got %d", len(cfg.Routes))
	}

	if len(cfg.Backends) != 1 {
		t.Errorf("expected 1 backend, got %d", len(cfg.Backends))
	}
}

func TestLoadFromYAMLNotFound(t *testing.T) {
	cfg, err := LoadFromYAML("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	if cfg != nil {
		t.Fatal("expected config to be nil on error")
	}
}

func TestLoadFromJSON(t *testing.T) {
	jsonStr := `{
  "server": {
    "host": "localhost",
    "port": "8080",
    "tls": false
  },
  "routes": [
    {
      "name": "api",
      "pattern": "/api",
      "methods": ["GET"],
      "backend_id": "backend1"
    }
  ],
  "backends": [
    {
      "id": "backend1",
      "servers": ["http://localhost:3000"]
    }
  ]
}`

	tmpfile, err := ioutil.TempFile("", "config*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(jsonStr); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	cfg, err := LoadFromJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg == nil {
		t.Fatal("expected config to be non-nil")
	}

	if cfg.Server.Host != "localhost" {
		t.Errorf("expected host, got %s", cfg.Server.Host)
	}
}

func TestLoadFromJSONNotFound(t *testing.T) {
	cfg, err := LoadFromJSON("/nonexistent/path.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	if cfg != nil {
		t.Fatal("expected config to be nil on error")
	}
}

func TestServerConfig(t *testing.T) {
	srv := &ServerConfig{
		Host: "localhost",
		Port: "8080",
		TLS:  false,
	}

	if srv.Host != "localhost" {
		t.Errorf("expected host, got %s", srv.Host)
	}
	if srv.Port != "8080" {
		t.Errorf("expected port 8080, got %s", srv.Port)
	}
	if srv.TLS {
		t.Error("expected TLS to be false")
	}
}

func TestRouteConfig(t *testing.T) {
	route := &RouteConfig{
		Name:      "api",
		Pattern:   "/api",
		Methods:   []string{"GET", "POST"},
		BackendID: "backend1",
		Priority:  10,
	}

	if route.Name != "api" {
		t.Errorf("expected name api, got %s", route.Name)
	}
	if route.Pattern != "/api" {
		t.Errorf("expected pattern /api, got %s", route.Pattern)
	}
	if len(route.Methods) != 2 {
		t.Errorf("expected 2 methods, got %d", len(route.Methods))
	}
	if route.Priority != 10 {
		t.Errorf("expected priority 10, got %d", route.Priority)
	}
}

func TestBackendConfig(t *testing.T) {
	backend := &BackendConfig{
		ID:            "backend1",
		LoadBalancing: "round_robin",
	}
	backend.Servers = append(backend.Servers, "http://localhost:3000")

	if backend.ID != "backend1" {
		t.Errorf("expected ID backend1, got %s", backend.ID)
	}
	if len(backend.Servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(backend.Servers))
	}
}

func TestHealthConfig(t *testing.T) {
	health := &HealthConfig{
		Enabled:  true,
		Interval: "10s",
		Timeout:  "5s",
		Path:     "/health",
	}

	if !health.Enabled {
		t.Error("expected health check to be enabled")
	}
	if health.Interval != "10s" {
		t.Errorf("expected interval 10s, got %s", health.Interval)
	}
	if health.Path != "/health" {
		t.Errorf("expected path /health, got %s", health.Path)
	}
}

func TestRateLimitPolicy(t *testing.T) {
	policy := &RateLimitPolicy{
		Enabled:     true,
		MaxRequests: 1000,
		Window:      "1s",
	}

	if !policy.Enabled {
		t.Error("expected rate limit to be enabled")
	}
	if policy.MaxRequests != 1000 {
		t.Errorf("expected 1000 requests, got %d", policy.MaxRequests)
	}
}

func TestCORSPolicy(t *testing.T) {
	policy := &CORSPolicy{
		Enabled:        true,
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:3001"},
	}

	if !policy.Enabled {
		t.Error("expected CORS to be enabled")
	}
	if len(policy.AllowedOrigins) != 2 {
		t.Errorf("expected 2 origins, got %d", len(policy.AllowedOrigins))
	}
}

func TestAuthPolicy(t *testing.T) {
	policy := &AuthPolicy{
		Enabled: true,
		Type:    "bearer",
		Secret:  "my-secret-key",
	}

	if !policy.Enabled {
		t.Error("expected auth to be enabled")
	}
	if policy.Type != "bearer" {
		t.Errorf("expected bearer type, got %s", policy.Type)
	}
}

func TestCachePolicy(t *testing.T) {
	policy := &CachePolicy{
		Enabled: true,
		TTL:     "3600s",
		Methods: []string{"GET", "HEAD"},
	}

	if !policy.Enabled {
		t.Error("expected cache to be enabled")
	}
	if policy.TTL != "3600s" {
		t.Errorf("expected TTL 3600s, got %s", policy.TTL)
	}
	if len(policy.Methods) != 2 {
		t.Errorf("expected 2 methods, got %d", len(policy.Methods))
	}
}

func TestLoadFromYAMLMultipleBackends(t *testing.T) {
	yaml := `
server:
  host: localhost
  port: "8080"
backends:
  - id: backend1
    servers:
      - http://localhost:3000
  - id: backend2
    servers:
      - http://localhost:3001
      - http://localhost:3002
`

	tmpfile, err := ioutil.TempFile("", "config*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(yaml); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	cfg, err := LoadFromYAML(tmpfile.Name())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(cfg.Backends) != 2 {
		t.Errorf("expected 2 backends, got %d", len(cfg.Backends))
	}

	if len(cfg.Backends[1].Servers) != 2 {
		t.Errorf("expected 2 servers in backend2, got %d", len(cfg.Backends[1].Servers))
	}
}

func TestLoadFromYAMLWithPolicies(t *testing.T) {
	yaml := `
server:
  host: localhost
  port: "8080"
policies:
  rate_limit:
    enabled: true
    max_requests: 1000
    window: "1s"
  cors:
    enabled: true
    allowed_origins:
      - http://localhost:3000
  cache:
    enabled: true
    ttl: "3600s"
    methods: [GET, HEAD]
`

	tmpfile, err := ioutil.TempFile("", "config*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(yaml); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	cfg, err := LoadFromYAML(tmpfile.Name())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !cfg.Policies.RateLimit.Enabled {
		t.Error("expected rate limit to be enabled")
	}

	if !cfg.Policies.CORS.Enabled {
		t.Error("expected CORS to be enabled")
	}

	if !cfg.Policies.Cache.Enabled {
		t.Error("expected cache to be enabled")
	}
}
