package backend

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

// Server represents a backend server
type Server struct {
	URL      *url.URL
	Weight   int32
	Healthy  int32 // 1 = healthy, 0 = unhealthy
	mu       sync.RWMutex
	metadata map[string]interface{}
}

// Pool manages multiple backend servers
type Pool struct {
	Servers    []*Server
	current    uint32
	mu         sync.RWMutex
	healthChan chan *Server
}

// NewPool creates a new backend pool
func NewPool() *Pool {
	return &Pool{
		Servers:    make([]*Server, 0),
		healthChan: make(chan *Server, 100),
	}
}

// AddServer adds a backend server to the pool
func (p *Pool) AddServer(rawURL string, weight int32) (*Server, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	server := &Server{
		URL:      u,
		Weight:   weight,
		Healthy:  1,
		metadata: make(map[string]interface{}),
	}

	p.mu.Lock()
	p.Servers = append(p.Servers, server)
	p.mu.Unlock()

	return server, nil
}

// GetServer returns a healthy backend server using round-robin
func (p *Pool) GetServer() *Server {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.Servers) == 0 {
		return nil
	}

	// Get healthy servers
	healthyServers := p.getHealthyServers()
	if len(healthyServers) == 0 {
		return nil
	}

	// Simple round-robin
	idx := atomic.AddUint32(&p.current, 1) % uint32(len(healthyServers))
	return healthyServers[idx]
}

// GetServerByIndex returns a specific server by index
func (p *Pool) GetServerByIndex(index int) *Server {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if index < 0 || index >= len(p.Servers) {
		return nil
	}

	return p.Servers[index]
}

// getHealthyServers returns only healthy servers (must be called with read lock)
func (p *Pool) getHealthyServers() []*Server {
	healthy := make([]*Server, 0)
	for _, server := range p.Servers {
		if atomic.LoadInt32(&server.Healthy) == 1 {
			healthy = append(healthy, server)
		}
	}
	return healthy
}

// SetServerHealth sets the health status of a server
func (p *Pool) SetServerHealth(server *Server, healthy bool) {
	val := int32(1)
	if !healthy {
		val = 0
	}
	atomic.StoreInt32(&server.Healthy, val)
}

// GetServerHealth returns the health status of a server
func (p *Pool) GetServerHealth(server *Server) bool {
	return atomic.LoadInt32(&server.Healthy) == 1
}

// HealthChecker periodically checks backend health
type HealthChecker struct {
	pool     *Pool
	interval time.Duration
	timeout  time.Duration
	path     string
	stopCh   chan struct{}
	client   *http.Client
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(pool *Pool, interval, timeout time.Duration, path string) *HealthChecker {
	if path == "" {
		path = "/health"
	}
	return &HealthChecker{
		pool:     pool,
		interval: interval,
		timeout:  timeout,
		path:     path,
		stopCh:   make(chan struct{}),
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Start begins health checking
func (hc *HealthChecker) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(hc.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-hc.stopCh:
				return
			case <-ticker.C:
				hc.checkHealth()
			}
		}
	}()
}

// Stop stops the health checker
func (hc *HealthChecker) Stop() {
	close(hc.stopCh)
}

// checkHealth checks the health of all servers
func (hc *HealthChecker) checkHealth() {
	hc.pool.mu.RLock()
	servers := make([]*Server, len(hc.pool.Servers))
	copy(servers, hc.pool.Servers)
	hc.pool.mu.RUnlock()

	for _, server := range servers {
		go hc.checkServer(server)
	}
}

// checkServer checks the health of a single server
func (hc *HealthChecker) checkServer(server *Server) {
	healthURL := server.URL.Scheme + "://" + server.URL.Host + hc.path

	ctx, cancel := context.WithTimeout(context.Background(), hc.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		hc.pool.SetServerHealth(server, false)
		return
	}

	resp, err := hc.client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		hc.pool.SetServerHealth(server, false)
		if resp != nil {
			resp.Body.Close()
		}
		return
	}
	resp.Body.Close()

	hc.pool.SetServerHealth(server, true)
}

// GetMetadata retrieves metadata for a server
func (s *Server) GetMetadata(key string) interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.metadata[key]
}

// SetMetadata sets metadata for a server
func (s *Server) SetMetadata(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metadata[key] = value
}
