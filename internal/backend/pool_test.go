package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewPool(t *testing.T) {
	pool := NewPool()
	if pool == nil {
		t.Fatal("expected pool to be non-nil")
	}
	if len(pool.Servers) != 0 {
		t.Errorf("expected empty servers, got %d", len(pool.Servers))
	}
}

func TestAddServer(t *testing.T) {
	pool := NewPool()

	_, err := pool.AddServer("http://localhost:3000", 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(pool.Servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(pool.Servers))
	}

	server := pool.Servers[0]
	if server.URL.String() != "http://localhost:3000" {
		t.Errorf("expected URL http://localhost:3000, got %s", server.URL.String())
	}
}

func TestAddServerInvalidURL(t *testing.T) {
	pool := NewPool()

	_, err := pool.AddServer("invalid://[url", 1)
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestGetServer(t *testing.T) {
	pool := NewPool()
	pool.AddServer("http://server1:3000", 1)
	pool.AddServer("http://server2:3000", 1)

	server := pool.GetServer()
	if server == nil {
		t.Fatal("expected server to be non-nil")
	}
	if server.URL.Host != "server1:3000" && server.URL.Host != "server2:3000" {
		t.Errorf("unexpected server: %s", server.URL.Host)
	}
}

func TestGetServerEmpty(t *testing.T) {
	pool := NewPool()

	server := pool.GetServer()
	if server != nil {
		t.Fatal("expected server to be nil for empty pool")
	}
}

func TestGetServerUnhealthy(t *testing.T) {
	pool := NewPool()
	server, _ := pool.AddServer("http://server1:3000", 1)

	// Mark server as unhealthy
	pool.SetServerHealth(server, false)

	result := pool.GetServer()
	if result != nil {
		t.Fatal("expected server to be nil when all unhealthy")
	}
}

func TestSetServerHealth(t *testing.T) {
	pool := NewPool()
	server, _ := pool.AddServer("http://server1:3000", 1)

	pool.SetServerHealth(server, false)
	if pool.GetServerHealth(server) {
		t.Fatal("expected server to be unhealthy")
	}

	pool.SetServerHealth(server, true)
	if !pool.GetServerHealth(server) {
		t.Fatal("expected server to be healthy")
	}
}

func TestGetServerByIndex(t *testing.T) {
	pool := NewPool()
	server1, _ := pool.AddServer("http://server1:3000", 1)
	server2, _ := pool.AddServer("http://server2:3000", 1)

	retrieved := pool.GetServerByIndex(0)
	if retrieved != server1 {
		t.Error("unexpected server at index 0")
	}

	retrieved = pool.GetServerByIndex(1)
	if retrieved != server2 {
		t.Error("unexpected server at index 1")
	}

	retrieved = pool.GetServerByIndex(99)
	if retrieved != nil {
		t.Fatal("expected nil for out of bounds index")
	}
}

func TestHealthChecker(t *testing.T) {
	// Create mock server that responds with health check
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	pool := NewPool()
	server, _ := pool.AddServer(mockServer.URL, 1)

	// Create health checker
	hc := NewHealthChecker(pool, 100*time.Millisecond, 1*time.Second, "/health")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hc.Start(ctx)

	// Wait for health check to run
	time.Sleep(200 * time.Millisecond)

	if !pool.GetServerHealth(server) {
		t.Fatal("expected server to be healthy after health check")
	}

	hc.Stop()
}

func TestHealthCheckerUnhealthy(t *testing.T) {
	// Create mock server that returns error
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer mockServer.Close()

	pool := NewPool()
	server, _ := pool.AddServer(mockServer.URL, 1)

	hc := NewHealthChecker(pool, 100*time.Millisecond, 1*time.Second, "/health")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hc.Start(ctx)

	// Wait for health check
	time.Sleep(200 * time.Millisecond)

	if pool.GetServerHealth(server) {
		t.Fatal("expected server to be unhealthy")
	}

	hc.Stop()
}

func TestServerMetadata(t *testing.T) {
	pool := NewPool()
	server, _ := pool.AddServer("http://localhost:3000", 1)

	server.SetMetadata("key1", "value1")
	if server.GetMetadata("key1") != "value1" {
		t.Fatal("metadata not set correctly")
	}

	server.SetMetadata("key2", 42)
	if server.GetMetadata("key2") != 42 {
		t.Fatal("integer metadata not set correctly")
	}

	if server.GetMetadata("nonexistent") != nil {
		t.Fatal("expected nil for nonexistent key")
	}
}

func TestRoundRobin(t *testing.T) {
	pool := NewPool()
	_, _ = pool.AddServer("http://server1:3000", 1)
	_, _ = pool.AddServer("http://server2:3000", 1)

	// Get multiple servers and check round-robin
	servers := make([]*Server, 0)
	for i := 0; i < 4; i++ {
		servers = append(servers, pool.GetServer())
	}

	// Should alternate between server1 and server2
	if servers[0].URL.Host != "server1:3000" && servers[0].URL.Host != "server2:3000" {
		t.Error("unexpected first server")
	}
}

func TestPoolConcurrency(t *testing.T) {
	pool := NewPool()
	pool.AddServer("http://server1:3000", 1)
	pool.AddServer("http://server2:3000", 1)

	done := make(chan bool)

	// Concurrent readers
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				pool.GetServer()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
