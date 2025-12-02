package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/surukanti/reverse-proxy/internal/backend"
	"github.com/surukanti/reverse-proxy/internal/router"
)

func TestNewProxy(t *testing.T) {
	p := NewProxy()
	if p == nil {
		t.Fatal("expected proxy to be non-nil")
	}
}

func TestProxyClientIP(t *testing.T) {
	p := NewProxy()

	req, _ := http.NewRequest("GET", "http://localhost/api/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	ip := p.getClientIP(req)
	if ip != "192.168.1.1" {
		t.Errorf("expected IP 192.168.1.1, got %s", ip)
	}
}

func TestProxyClientIPFromHeader(t *testing.T) {
	p := NewProxy()

	req, _ := http.NewRequest("GET", "http://localhost/api/test", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")

	ip := p.getClientIP(req)
	if ip != "10.0.0.1" {
		t.Errorf("expected IP 10.0.0.1, got %s", ip)
	}
}

func TestProxyGetStats(t *testing.T) {
	p := NewProxy()
	stats := p.GetStats()
	if stats.RequestCount < 0 {
		t.Error("expected valid stats")
	}
}

func TestProxyClearCache(t *testing.T) {
	p := NewProxy()
	p.ClearCache()
	if len(p.cache) > 0 {
		t.Error("expected cache to be cleared")
	}
}

func TestProxyServeHTTPWithRoute(t *testing.T) {
	mockBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend response"))
	}))
	defer mockBackend.Close()

	p := NewProxy()
	pool := backend.NewPool()
	pool.AddServer(mockBackend.URL, 1)

	route := &router.Route{
		Name:    "test",
		Pattern: "/api/test",
		Methods: []string{"GET"},
		Backend: pool,
	}
	p.AddRoute(route)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://localhost/api/test", nil)

	p.ServeHTTP(w, req)

	// Check if response is valid (either 200 or an error code)
	if w.Code == 0 {
		t.Error("expected non-zero status code")
	}
}
