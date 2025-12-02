package middleware

import (
	"net/http"
	"time"
)

type Handler func(http.ResponseWriter, *http.Request) error

type Chain struct {
	handlers []Handler
}

func NewChain() *Chain {
	return &Chain{
		handlers: make([]Handler, 0),
	}
}

func (c *Chain) Add(handler Handler) *Chain {
	c.handlers = append(c.handlers, handler)
	return c
}

func (c *Chain) Execute(w http.ResponseWriter, r *http.Request) error {
	for _, handler := range c.handlers {
		if err := handler(w, r); err != nil {
			return err
		}
	}
	return nil
}

type LoggingMiddleware struct {
	logger func(string)
}

func NewLoggingMiddleware(logger func(string)) *LoggingMiddleware {
	return &LoggingMiddleware{logger: logger}
}

func (lm *LoggingMiddleware) Handle(w http.ResponseWriter, r *http.Request) error {
	start := time.Now()
	lm.logger(r.Method + " " + r.URL.Path + " from " + r.RemoteAddr)

	go func() {
		time.Sleep(100 * time.Millisecond)
		duration := time.Since(start)
		lm.logger("Request completed in " + duration.String())
	}()

	return nil
}

type RateLimiter struct {
	maxRequests int
	window      time.Duration
	buckets     map[string]*bucket
}

type bucket struct {
	tokens    float64
	lastReset time.Time
}

func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		maxRequests: maxRequests,
		window:      window,
		buckets:     make(map[string]*bucket),
	}
}

func (rl *RateLimiter) Handle(identifier string) bool {
	now := time.Now()
	b, exists := rl.buckets[identifier]

	if !exists {
		rl.buckets[identifier] = &bucket{
			tokens:    float64(rl.maxRequests),
			lastReset: now,
		}
		return true
	}

	elapsed := now.Sub(b.lastReset).Seconds()
	refillRate := float64(rl.maxRequests) / rl.window.Seconds()
	b.tokens = minFloat(float64(rl.maxRequests), b.tokens+refillRate*elapsed)
	b.lastReset = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}

	return false
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

type AuthMiddleware struct {
	validator func(token string) bool
}

func NewAuthMiddleware(validator func(string) bool) *AuthMiddleware {
	return &AuthMiddleware{
		validator: validator,
	}
}

func (am *AuthMiddleware) Handle(w http.ResponseWriter, r *http.Request) error {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return ErrUnauthorized
	}

	if !am.validator(token) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return ErrForbidden
	}

	return nil
}

type CORSMiddleware struct {
	allowedOrigins []string
}

func NewCORSMiddleware(allowedOrigins []string) *CORSMiddleware {
	return &CORSMiddleware{
		allowedOrigins: allowedOrigins,
	}
}

func (cm *CORSMiddleware) Handle(w http.ResponseWriter, r *http.Request) error {
	origin := r.Header.Get("Origin")
	allowed := false

	for _, o := range cm.allowedOrigins {
		if o == "*" || o == origin {
			allowed = true
			break
		}
	}

	if allowed {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	}

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return nil
	}

	return nil
}
