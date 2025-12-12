package advanced

import (
	"testing"
	"time"

	"github.com/surukanti/reverse-proxy/internal/backend"
)

func TestABTestManager(t *testing.T) {
	manager := NewABTestManager()
	if manager == nil {
		t.Fatal("expected AB test manager to be non-nil")
	}
}

func TestCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(5, 3, 1*time.Second)
	if cb == nil {
		t.Fatal("expected circuit breaker to be non-nil")
	}
}

func TestBlueGreenManager(t *testing.T) {
	pool1 := backend.NewPool()
	pool2 := backend.NewPool()
	manager := NewBlueGreenManager(pool1, pool2)
	if manager == nil {
		t.Fatal("expected blue-green manager to be non-nil")
	}
}

func TestTenantRateLimiter(t *testing.T) {
	limiter := NewTenantRateLimiter()
	if limiter == nil {
		t.Fatal("expected tenant rate limiter to be non-nil")
	}
}
