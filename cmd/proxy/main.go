package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/surukanti/reverse-proxy/internal/backend"
	"github.com/surukanti/reverse-proxy/internal/config"
	"github.com/surukanti/reverse-proxy/internal/middleware"
	"github.com/surukanti/reverse-proxy/internal/proxy"
	"github.com/surukanti/reverse-proxy/internal/router"
)

func main() {
	configFile := flag.String("config", "config.yaml", "Configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadFromYAML(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Config loaded successfully")
	log.Printf("  Backends: %d", len(cfg.Backends))
	log.Printf("  Routes: %d", len(cfg.Routes))

	// Create proxy
	p := proxy.NewProxy()

	// Setup backends
	backends := make(map[string]*backend.Pool)
	for _, backendCfg := range cfg.Backends {
		log.Printf("Setting up backend: %s", backendCfg.ID)
		pool := backend.NewPool()
		for i, serverURL := range backendCfg.Servers {
			log.Printf("  Adding server %d: %s", i, serverURL)
			server, err := pool.AddServer(serverURL, 1)
			if err != nil {
				log.Printf("  Failed to add server: %v", err)
				continue
			}
			log.Printf("  Server added successfully: URL=%v", server.URL)
		}
		log.Printf("Backend %s has %d servers", backendCfg.ID, len(pool.Servers))

		// Setup health checking
		if backendCfg.HealthCheck.Enabled {
			interval := 30 * time.Second
			timeout := 5 * time.Second
			hc := backend.NewHealthChecker(pool, interval, timeout, backendCfg.HealthCheck.Path)
			hc.Start(context.Background())
		}

		backends[backendCfg.ID] = pool
	}

	// Setup routes
	for _, routeCfg := range cfg.Routes {
		pool, ok := backends[routeCfg.BackendID]
		if !ok {
			log.Printf("Backend %s not found for route %s", routeCfg.BackendID, routeCfg.Name)
			continue
		}

		route := &router.Route{
			Name:       routeCfg.Name,
			Pattern:    routeCfg.Pattern,
			PathPrefix: routeCfg.PathPrefix,
			Subdomain:  routeCfg.Subdomain,
			Headers:    routeCfg.Headers,
			Methods:    routeCfg.Methods,
			Backend:    pool,
			Priority:   routeCfg.Priority,
		}

		err := p.AddRoute(route)
		if err != nil {
			log.Printf("Failed to add route: %v", err)
		}
	}

	// Setup middleware
	if cfg.Policies.CORS.Enabled {
		corsMiddleware := middleware.NewCORSMiddleware(cfg.Policies.CORS.AllowedOrigins)
		p.AddMiddleware(corsMiddleware.Handle)
	}

	if cfg.Policies.Auth.Enabled {
		authMiddleware := middleware.NewAuthMiddleware(func(token string) bool {
			return token != ""
		})
		p.AddMiddleware(authMiddleware.Handle)
	}

	// Setup logging
	loggingMiddleware := middleware.NewLoggingMiddleware(func(msg string) {
		log.Println(msg)
	})
	p.AddMiddleware(loggingMiddleware.Handle)

	// Setup rate limiting
	if cfg.Policies.RateLimit.Enabled {
		p.SetRateLimit(cfg.Policies.RateLimit.MaxRequests, 1*time.Minute)
	}

	// Setup event handlers
	p.On("request_forwarded", func(event proxy.Event) {
		log.Printf("Request forwarded: %s %s", event.Request.Method, event.Request.URL.Path)
	})

	p.On("cache_hit", func(event proxy.Event) {
		log.Printf("Cache hit: %s %s", event.Request.Method, event.Request.URL.Path)
	})

	p.On("proxy_error", func(event proxy.Event) {
		log.Printf("Proxy error: %v", event.Error)
	})

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	server := &http.Server{
		Addr:    addr,
		Handler: p,
	}

	log.Printf("Starting reverse proxy on %s", addr)

	// Graceful shutdown
	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
		<-sigch

		log.Println("Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Server shutdown error: %v", err)
		}
	}()

	// Start HTTP server
	if cfg.Server.TLS {
		if err := server.ListenAndServeTLS(cfg.Server.CertFile, cfg.Server.KeyFile); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	} else {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}
}
