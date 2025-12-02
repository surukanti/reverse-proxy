package router

import (
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/surukanti/reverse-proxy/internal/backend"
)

// Route represents a routing rule
type Route struct {
	Name       string
	Pattern    string
	PathPrefix string
	Subdomain  string
	Headers    map[string]string
	Methods    []string
	Backend    *backend.Pool
	Priority   int
	regex      *regexp.Regexp
}

// Router manages routing rules
type Router struct {
	routes []*Route
	mu     sync.RWMutex
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		routes: make([]*Route, 0),
	}
}

// AddRoute adds a new routing rule
func (r *Router) AddRoute(route *Route) error {
	if route.Pattern != "" {
		regex, err := regexp.Compile(route.Pattern)
		if err != nil {
			return err
		}
		route.regex = regex
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.routes = append(r.routes, route)
	// Sort by priority (higher priority first)
	r.sortRoutes()

	return nil
}

// Match finds the best matching route for a request
func (r *Router) Match(req *http.Request) *Route {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, route := range r.routes {
		if r.matchRoute(route, req) {
			return route
		}
	}

	return nil
}

// matchRoute checks if a request matches a route
func (r *Router) matchRoute(route *Route, req *http.Request) bool {
	// Check method
	if len(route.Methods) > 0 {
		methodMatch := false
		for _, method := range route.Methods {
			if method == req.Method {
				methodMatch = true
				break
			}
		}
		if !methodMatch {
			return false
		}
	}

	// Check subdomain
	if route.Subdomain != "" {
		host := req.Host
		if strings.HasPrefix(host, ":") {
			host = strings.Split(host, ":")[0]
		}
		subdomain := strings.Split(host, ".")[0]
		if subdomain != route.Subdomain {
			return false
		}
	}

	// Check headers
	if len(route.Headers) > 0 {
		for key, value := range route.Headers {
			if req.Header.Get(key) != value {
				return false
			}
		}
	}

	// Check path prefix
	if route.PathPrefix != "" {
		if !strings.HasPrefix(req.URL.Path, route.PathPrefix) {
			return false
		}
	}

	// Check regex pattern
	if route.regex != nil {
		if !route.regex.MatchString(req.URL.Path) {
			return false
		}
	}

	return true
}

// RemoveRoute removes a route by name
func (r *Router) RemoveRoute(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, route := range r.routes {
		if route.Name == name {
			r.routes = append(r.routes[:i], r.routes[i+1:]...)
			return true
		}
	}

	return false
}

// ListRoutes returns all routes
func (r *Router) ListRoutes() []*Route {
	r.mu.RLock()
	defer r.mu.RUnlock()

	routes := make([]*Route, len(r.routes))
	copy(routes, r.routes)
	return routes
}

// sortRoutes sorts routes by priority (higher first)
func (r *Router) sortRoutes() {
	// Simple bubble sort (in production, use better algorithm)
	for i := 0; i < len(r.routes); i++ {
		for j := i + 1; j < len(r.routes); j++ {
			if r.routes[j].Priority > r.routes[i].Priority {
				r.routes[i], r.routes[j] = r.routes[j], r.routes[i]
			}
		}
	}
}

// ContentRouter routes based on request content
type ContentRouter struct {
	router *Router
}

// NewContentRouter creates a content-aware router
func NewContentRouter() *ContentRouter {
	return &ContentRouter{
		router: NewRouter(),
	}
}

// RouteByContentType returns a backend based on content type
func (cr *ContentRouter) RouteByContentType(req *http.Request, contentTypeRoutes map[string]*backend.Pool) *backend.Pool {
	contentType := req.Header.Get("Content-Type")
	if pool, ok := contentTypeRoutes[contentType]; ok {
		return pool
	}
	return nil
}

// GetRouter returns the underlying router
func (cr *ContentRouter) GetRouter() *Router {
	return cr.router
}
