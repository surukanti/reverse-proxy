package proxy

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"
	"time"

	"github.com/surukanti/reverse-proxy/internal/backend"
	"github.com/surukanti/reverse-proxy/internal/middleware"
	"github.com/surukanti/reverse-proxy/internal/router"
)

// Proxy represents the reverse proxy
type Proxy struct {
	router        *router.Router
	middlewares   *middleware.Chain
	rateLimiter   *middleware.RateLimiter
	transport     *http.Transport
	mu            sync.RWMutex
	requestCount  int64
	errorCount    int64
	cache         map[string]*CacheEntry
	cacheMu       sync.RWMutex
	eventHandlers map[string][]func(Event)
}

// CacheEntry represents a cached response
type CacheEntry struct {
	Status  int
	Headers http.Header
	Body    []byte
	Expires time.Time
}

// Event represents a proxy event
type Event struct {
	Type      string
	Timestamp time.Time
	Request   *http.Request
	Response  *http.Response
	Error     error
}

// NewProxy creates a new reverse proxy
func NewProxy() *Proxy {
	return &Proxy{
		router:        router.NewRouter(),
		middlewares:   middleware.NewChain(),
		rateLimiter:   middleware.NewRateLimiter(1000, time.Minute),
		transport:     &http.Transport{},
		cache:         make(map[string]*CacheEntry),
		eventHandlers: make(map[string][]func(Event)),
	}
}

// Router returns the router
func (p *Proxy) Router() *router.Router {
	return p.router
}

// AddRoute adds a new route
func (p *Proxy) AddRoute(route *router.Route) error {
	return p.router.AddRoute(route)
}

// AddMiddleware adds a middleware handler
func (p *Proxy) AddMiddleware(handler middleware.Handler) *Proxy {
	p.middlewares.Add(handler)
	return p
}

// SetRateLimit sets the rate limit
func (p *Proxy) SetRateLimit(maxRequests int, window time.Duration) {
	p.rateLimiter = middleware.NewRateLimiter(maxRequests, window)
}

// ServeHTTP implements http.Handler
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check rate limit
	if !p.rateLimiter.Handle(r.RemoteAddr) {
		p.emitEvent(Event{
			Type:      "rate_limit_exceeded",
			Timestamp: time.Now(),
			Request:   r,
		})
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	// Execute middleware chain
	if err := p.middlewares.Execute(w, r); err != nil {
		p.emitEvent(Event{
			Type:      "middleware_error",
			Timestamp: time.Now(),
			Request:   r,
			Error:     err,
		})
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// Find matching route
	route := p.router.Match(r)
	if route == nil {
		p.emitEvent(Event{
			Type:      "no_route_found",
			Timestamp: time.Now(),
			Request:   r,
		})
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Get backend server
	server := route.Backend.GetServer()
	if server == nil {
		p.emitEvent(Event{
			Type:      "no_backend_available",
			Timestamp: time.Now(),
			Request:   r,
		})
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	// Check cache
	cacheKey := p.getCacheKey(r, server)
	if cached, ok := p.cache[cacheKey]; ok && cached.Expires.After(time.Now()) {
		p.serveCached(w, cached)
		p.emitEvent(Event{
			Type:      "cache_hit",
			Timestamp: time.Now(),
			Request:   r,
		})
		return
	}

	// Forward request
	p.forwardRequest(w, r, server)
}

// forwardRequest forwards the request to the backend server
func (p *Proxy) forwardRequest(w http.ResponseWriter, r *http.Request, server *backend.Server) {
	// Validate server URL
	if server == nil || server.URL == nil {
		p.emitEvent(Event{
			Type:      "proxy_error",
			Timestamp: time.Now(),
			Request:   r,
			Error:     fmt.Errorf("invalid server or server URL is nil"),
		})
		http.Error(w, "Bad Gateway: invalid server URL", http.StatusBadGateway)
		return
	}


	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(server.URL)

	// Set custom transport
	proxy.Transport = p.transport

	// Custom error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		p.emitEvent(Event{
			Type:      "proxy_error",
			Timestamp: time.Now(),
			Request:   r,
			Error:     err,
		})
		http.Error(w, fmt.Sprintf("Bad Gateway: %v", err), http.StatusBadGateway)
	}

	// Modify request - use the default Director from NewSingleHostReverseProxy and add our headers
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Header.Set("X-Forwarded-For", p.getClientIP(r))
		req.Header.Set("X-Forwarded-Proto", r.Header.Get("X-Forwarded-Proto"))
		if req.Header.Get("X-Forwarded-Proto") == "" {
			req.Header.Set("X-Forwarded-Proto", "http")
		}
		req.Header.Set("X-Real-IP", r.RemoteAddr)
	}

	p.emitEvent(Event{
		Type:      "request_forwarded",
		Timestamp: time.Now(),
		Request:   r,
	})

	proxy.ServeHTTP(w, r)
}

// serveCached serves a cached response
func (p *Proxy) serveCached(w http.ResponseWriter, entry *CacheEntry) {
	for key, values := range entry.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.Header().Set("X-Cache", "HIT")
	w.WriteHeader(entry.Status)
	w.Write(entry.Body)
}

// getCacheKey generates a cache key
func (p *Proxy) getCacheKey(r *http.Request, server *backend.Server) string {
	return r.Method + ":" + r.URL.Path + ":" + server.URL.String()
}

// CacheResponse caches a response
func (p *Proxy) CacheResponse(r *http.Request, server *backend.Server, status int, headers http.Header, body []byte, ttl time.Duration) {
	key := p.getCacheKey(r, server)
	p.cacheMu.Lock()
	defer p.cacheMu.Unlock()

	p.cache[key] = &CacheEntry{
		Status:  status,
		Headers: headers,
		Body:    body,
		Expires: time.Now().Add(ttl),
	}
}

// ClearCache clears the cache
func (p *Proxy) ClearCache() {
	p.cacheMu.Lock()
	defer p.cacheMu.Unlock()
	p.cache = make(map[string]*CacheEntry)
}

// On registers an event handler
func (p *Proxy) On(eventType string, handler func(Event)) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.eventHandlers[eventType]; !ok {
		p.eventHandlers[eventType] = make([]func(Event), 0)
	}

	p.eventHandlers[eventType] = append(p.eventHandlers[eventType], handler)
}

// emitEvent emits an event
func (p *Proxy) emitEvent(event Event) {
	p.mu.RLock()
	handlers := p.eventHandlers[event.Type]
	p.mu.RUnlock()

	for _, handler := range handlers {
		go handler(event)
	}
}

// getClientIP extracts the client IP from the request
func (p *Proxy) getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}

// Stats represents proxy statistics
type Stats struct {
	RequestCount int64
	ErrorCount   int64
	CacheSize    int
}

// GetStats returns proxy statistics
func (p *Proxy) GetStats() Stats {
	p.cacheMu.RLock()
	cacheSize := len(p.cache)
	p.cacheMu.RUnlock()

	return Stats{
		RequestCount: p.requestCount,
		ErrorCount:   p.errorCount,
		CacheSize:    cacheSize,
	}
}
