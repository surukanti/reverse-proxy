package router

import (
	"net/http"
	"testing"

	"github.com/surukanti/reverse-proxy/internal/backend"
)

func TestNewRouter(t *testing.T) {
	r := NewRouter()
	if r == nil {
		t.Fatal("expected router to be non-nil")
	}
	routes := r.ListRoutes()
	if len(routes) != 0 {
		t.Errorf("expected empty routes, got %d", len(routes))
	}
}

func TestAddRoute(t *testing.T) {
	r := NewRouter()
	pool := backend.NewPool()

	route := &Route{
		Name:     "users",
		Pattern:  "/api/users",
		Methods:  []string{"GET", "POST"},
		Priority: 10,
		Backend:  pool,
	}

	err := r.AddRoute(route)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	routes := r.ListRoutes()
	if len(routes) != 1 {
		t.Errorf("expected 1 route, got %d", len(routes))
	}
}

func TestAddMultipleRoutes(t *testing.T) {
	r := NewRouter()
	pool := backend.NewPool()

	routes := []*Route{
		{Name: "users", Pattern: "/api/users", Priority: 10, Backend: pool},
		{Name: "products", Pattern: "/api/products", Priority: 20, Backend: pool},
		{Name: "orders", Pattern: "/api/orders", Priority: 15, Backend: pool},
	}

	for _, route := range routes {
		r.AddRoute(route)
	}

	listed := r.ListRoutes()
	if len(listed) != 3 {
		t.Errorf("expected 3 routes, got %d", len(listed))
	}
}

func TestRemoveRoute(t *testing.T) {
	r := NewRouter()
	pool := backend.NewPool()

	route := &Route{
		Name:     "users",
		Pattern:  "/api/users",
		Methods:  []string{"GET"},
		Priority: 10,
		Backend:  pool,
	}

	r.AddRoute(route)
	if len(r.ListRoutes()) != 1 {
		t.Fatal("route not added")
	}

	removed := r.RemoveRoute("users")
	if !removed {
		t.Fatal("expected route to be removed")
	}

	if len(r.ListRoutes()) != 0 {
		t.Errorf("expected 0 routes after removal, got %d", len(r.ListRoutes()))
	}
}

func TestRemoveRouteNotFound(t *testing.T) {
	r := NewRouter()

	removed := r.RemoveRoute("nonexistent")
	if removed {
		t.Fatal("expected remove to return false for nonexistent route")
	}
}

func TestListRoutes(t *testing.T) {
	r := NewRouter()
	pool := backend.NewPool()

	routes := []*Route{
		{Name: "users", Pattern: "/api/users", Priority: 10, Backend: pool},
		{Name: "products", Pattern: "/api/products", Priority: 20, Backend: pool},
	}

	for _, route := range routes {
		r.AddRoute(route)
	}

	listed := r.ListRoutes()
	if len(listed) != 2 {
		t.Errorf("expected 2 routes, got %d", len(listed))
	}
}

func TestMatchPattern(t *testing.T) {
	r := NewRouter()
	pool := backend.NewPool()

	route := &Route{
		Name:     "users",
		Pattern:  "/api/users",
		Methods:  []string{"GET"},
		Priority: 10,
		Backend:  pool,
	}
	r.AddRoute(route)

	req, _ := http.NewRequest("GET", "http://localhost/api/users", nil)
	matched := r.Match(req)

	if matched == nil {
		t.Fatal("expected route to match")
	}
	if matched.Name != "users" {
		t.Errorf("expected route name users, got %s", matched.Name)
	}
}

func TestMatchPatternNotFound(t *testing.T) {
	r := NewRouter()
	pool := backend.NewPool()

	route := &Route{
		Name:     "users",
		Pattern:  "/api/users",
		Methods:  []string{"GET"},
		Priority: 10,
		Backend:  pool,
	}
	r.AddRoute(route)

	req, _ := http.NewRequest("GET", "http://localhost/api/products", nil)
	matched := r.Match(req)

	if matched != nil {
		t.Fatal("expected no match for different path")
	}
}

func TestMatchMethod(t *testing.T) {
	r := NewRouter()
	pool := backend.NewPool()

	route := &Route{
		Name:     "users",
		Pattern:  "/api/users",
		Methods:  []string{"GET"},
		Priority: 10,
		Backend:  pool,
	}
	r.AddRoute(route)

	// Test matching method
	req, _ := http.NewRequest("GET", "http://localhost/api/users", nil)
	matched := r.Match(req)
	if matched == nil {
		t.Fatal("expected GET to match")
	}

	// Test non-matching method
	req, _ = http.NewRequest("POST", "http://localhost/api/users", nil)
	matched = r.Match(req)
	if matched != nil {
		t.Fatal("expected POST to not match when only GET allowed")
	}
}

func TestMatchPriority(t *testing.T) {
	r := NewRouter()
	pool := backend.NewPool()

	route1 := &Route{
		Name:     "users1",
		Pattern:  "/api/users",
		Methods:  []string{"GET"},
		Priority: 10,
		Backend:  pool,
	}
	route2 := &Route{
		Name:     "users2",
		Pattern:  "/api/users",
		Methods:  []string{"GET"},
		Priority: 20,
		Backend:  pool,
	}

	r.AddRoute(route1)
	r.AddRoute(route2)

	req, _ := http.NewRequest("GET", "http://localhost/api/users", nil)
	matched := r.Match(req)

	// Higher priority should match first
	if matched.Name != "users2" {
		t.Errorf("expected users2 (higher priority), got %s", matched.Name)
	}
}

func TestMatchPathPrefix(t *testing.T) {
	r := NewRouter()
	pool := backend.NewPool()

	route := &Route{
		Name:       "api",
		PathPrefix: "/api",
		Methods:    []string{"GET"},
		Priority:   10,
		Backend:    pool,
	}
	r.AddRoute(route)

	req, _ := http.NewRequest("GET", "http://localhost/api/users", nil)
	matched := r.Match(req)

	if matched == nil {
		t.Fatal("expected path prefix to match")
	}
}

func TestMatchSubdomain(t *testing.T) {
	r := NewRouter()
	pool := backend.NewPool()

	route := &Route{
		Name:      "api",
		Pattern:   "/users",
		Subdomain: "api",
		Methods:   []string{"GET"},
		Priority:  10,
		Backend:   pool,
	}
	r.AddRoute(route)

	req, _ := http.NewRequest("GET", "http://api.localhost/users", nil)
	req.Host = "api.localhost"
	matched := r.Match(req)

	if matched == nil {
		t.Fatal("expected subdomain to match")
	}
}

func TestMatchSubdomainMismatch(t *testing.T) {
	r := NewRouter()
	pool := backend.NewPool()

	route := &Route{
		Name:      "api",
		Pattern:   "/users",
		Subdomain: "api",
		Methods:   []string{"GET"},
		Priority:  10,
		Backend:   pool,
	}
	r.AddRoute(route)

	req, _ := http.NewRequest("GET", "http://other.localhost/users", nil)
	req.Host = "other.localhost"
	matched := r.Match(req)

	if matched != nil {
		t.Fatal("expected subdomain mismatch")
	}
}

func TestMatchHeader(t *testing.T) {
	r := NewRouter()
	pool := backend.NewPool()

	route := &Route{
		Name:    "api",
		Pattern: "/users",
		Methods: []string{"GET"},
		Headers: map[string]string{
			"X-API-Version": "v2",
		},
		Priority: 10,
		Backend:  pool,
	}
	r.AddRoute(route)

	req, _ := http.NewRequest("GET", "http://localhost/users", nil)
	req.Header.Set("X-API-Version", "v2")
	matched := r.Match(req)

	if matched == nil {
		t.Fatal("expected header match")
	}
}

func TestMatchHeaderMismatch(t *testing.T) {
	r := NewRouter()
	pool := backend.NewPool()

	route := &Route{
		Name:    "api",
		Pattern: "/users",
		Methods: []string{"GET"},
		Headers: map[string]string{
			"X-API-Version": "v2",
		},
		Priority: 10,
		Backend:  pool,
	}
	r.AddRoute(route)

	req, _ := http.NewRequest("GET", "http://localhost/users", nil)
	req.Header.Set("X-API-Version", "v1")
	matched := r.Match(req)

	if matched != nil {
		t.Fatal("expected no match for different header value")
	}
}

func TestAddRouteWithInvalidRegex(t *testing.T) {
	r := NewRouter()
	pool := backend.NewPool()

	route := &Route{
		Name:    "invalid",
		Pattern: "[invalid(regex",
		Backend: pool,
	}

	err := r.AddRoute(route)
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestContentRouter(t *testing.T) {
	cr := NewContentRouter()
	if cr == nil {
		t.Fatal("expected content router to be non-nil")
	}

	router := cr.GetRouter()
	if router == nil {
		t.Fatal("expected to get router from content router")
	}
}

func TestAddContentRoute(t *testing.T) {
	cr := NewContentRouter()
	pool := backend.NewPool()

	route := &Route{
		Name:     "users",
		Pattern:  "/api/users",
		Methods:  []string{"GET"},
		Priority: 10,
		Backend:  pool,
	}

	router := cr.GetRouter()
	err := router.AddRoute(route)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	routes := router.ListRoutes()
	if len(routes) != 1 {
		t.Errorf("expected 1 route, got %d", len(routes))
	}
}

func TestContentRouterByContentType(t *testing.T) {
	cr := NewContentRouter()
	pool := backend.NewPool()

	contentTypeRoutes := map[string]*backend.Pool{
		"application/json": pool,
	}

	req, _ := http.NewRequest("GET", "http://localhost/api/users", nil)
	req.Header.Set("Content-Type", "application/json")

	result := cr.RouteByContentType(req, contentTypeRoutes)

	if result != pool {
		t.Fatal("expected content type route to match")
	}
}
