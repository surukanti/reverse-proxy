package proxy

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/surukanti/reverse-proxy/internal/backend"
	"github.com/surukanti/reverse-proxy/internal/middleware"
)

// ABTestManager manages A/B testing
type ABTestManager struct {
	tests map[string]*ABTest
}

// ABTest represents an A/B test
type ABTest struct {
	Name         string
	VariantA     *backend.Pool
	VariantB     *backend.Pool
	SplitPercent float64 // 0-100, percentage for variant B
	ErrorRateA   float64
	ErrorRateB   float64
	SuccessRateA float64
	SuccessRateB float64
	requestsA    int64
	requestsB    int64
	successA     int64
	successB     int64
	errorsA      int64
	errorsB      int64
}

// NewABTestManager creates a new A/B test manager
func NewABTestManager() *ABTestManager {
	return &ABTestManager{
		tests: make(map[string]*ABTest),
	}
}

// AddTest adds a new A/B test
func (atm *ABTestManager) AddTest(test *ABTest) {
	atm.tests[test.Name] = test
}

// SelectVariant selects the variant for a request
func (atm *ABTestManager) SelectVariant(testName string, req *http.Request) *backend.Pool {
	test, ok := atm.tests[testName]
	if !ok {
		return test.VariantA
	}

	// Use user ID or cookie for consistent routing
	userID := req.Header.Get("X-User-ID")
	if userID == "" {
		if cookie, err := req.Cookie("user_id"); err == nil {
			userID = cookie.Value
		}
	}

	hash := HashString(userID)
	if hash%100 < int64(test.SplitPercent) {
		atomic.AddInt64(&test.requestsB, 1)
		return test.VariantB
	}

	atomic.AddInt64(&test.requestsA, 1)
	return test.VariantA
}

// RecordSuccess records a successful request
func (atm *ABTestManager) RecordSuccess(testName string, variantB bool) {
	test, ok := atm.tests[testName]
	if !ok {
		return
	}

	if variantB {
		atomic.AddInt64(&test.successB, 1)
	} else {
		atomic.AddInt64(&test.successA, 1)
	}
}

// RecordError records a failed request
func (atm *ABTestManager) RecordError(testName string, variantB bool) {
	test, ok := atm.tests[testName]
	if !ok {
		return
	}

	if variantB {
		atomic.AddInt64(&test.errorsB, 1)
	} else {
		atomic.AddInt64(&test.errorsA, 1)
	}
}

// GetStats returns stats for a test
func (atm *ABTestManager) GetStats(testName string) (requestsA, requestsB, successA, successB, errorsA, errorsB int64) {
	test, ok := atm.tests[testName]
	if !ok {
		return
	}

	requestsA = atomic.LoadInt64(&test.requestsA)
	requestsB = atomic.LoadInt64(&test.requestsB)
	successA = atomic.LoadInt64(&test.successA)
	successB = atomic.LoadInt64(&test.successB)
	errorsA = atomic.LoadInt64(&test.errorsA)
	errorsB = atomic.LoadInt64(&test.errorsB)

	// Calculate rates
	if requestsA > 0 {
		test.SuccessRateA = float64(successA) / float64(requestsA)
		test.ErrorRateA = float64(errorsA) / float64(requestsA)
	}
	if requestsB > 0 {
		test.SuccessRateB = float64(successB) / float64(requestsB)
		test.ErrorRateB = float64(errorsB) / float64(requestsB)
	}

	return
}

// BlueGreenManager manages blue-green deployments
type BlueGreenManager struct {
	blue          *backend.Pool
	green         *backend.Pool
	activeVersion string  // "blue" or "green"
	trafficShift  float64 // 0-100, percentage to shift to new version
	startTime     time.Time
	shiftDuration time.Duration
}

// NewBlueGreenManager creates a new blue-green manager
func NewBlueGreenManager(blue, green *backend.Pool) *BlueGreenManager {
	return &BlueGreenManager{
		blue:          blue,
		green:         green,
		activeVersion: "blue",
	}
}

// SelectBackend selects the backend based on traffic shift
func (bgm *BlueGreenManager) SelectBackend(req *http.Request) *backend.Pool {
	// Use user ID for consistent routing
	userID := req.Header.Get("X-User-ID")
	if userID == "" {
		if cookie, err := req.Cookie("user_id"); err == nil {
			userID = cookie.Value
		}
	}

	hash := HashString(userID)
	if hash%100 < int64(bgm.trafficShift) {
		if bgm.activeVersion == "blue" {
			return bgm.green
		}
		return bgm.blue
	}

	if bgm.activeVersion == "blue" {
		return bgm.blue
	}
	return bgm.green
}

// StartGradualShift starts a gradual traffic shift
func (bgm *BlueGreenManager) StartGradualShift(targetVersion string, duration time.Duration) {
	bgm.startTime = time.Now()
	bgm.shiftDuration = duration

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			elapsed := time.Since(bgm.startTime)
			if elapsed >= duration {
				bgm.trafficShift = 100
				bgm.activeVersion = targetVersion
				return
			}

			progress := float64(elapsed.Milliseconds()) / float64(duration.Milliseconds())
			bgm.trafficShift = progress * 100
		}
	}()
}

// GetStatus returns the current status
func (bgm *BlueGreenManager) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"active_version": bgm.activeVersion,
		"traffic_shift":  bgm.trafficShift,
		"shift_duration": bgm.shiftDuration.String(),
		"elapsed":        time.Since(bgm.startTime).String(),
	}
}

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	state            string // "closed", "open", "half-open"
	failureCount     int64
	successCount     int64
	failureThreshold int64
	successThreshold int64
	timeout          time.Duration
	lastFailureTime  time.Time
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failureThreshold, successThreshold int64, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            "closed",
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
	}
}

// Call executes a call with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	if cb.state == "open" {
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.state = "half-open"
			cb.successCount = 0
		} else {
			return fmt.Errorf("circuit breaker is open")
		}
	}

	err := fn()

	if err != nil {
		atomic.AddInt64(&cb.failureCount, 1)
		cb.lastFailureTime = time.Now()

		if atomic.LoadInt64(&cb.failureCount) >= cb.failureThreshold {
			cb.state = "open"
		}

		return err
	}

	atomic.AddInt64(&cb.successCount, 1)

	if cb.state == "half-open" && atomic.LoadInt64(&cb.successCount) >= cb.successThreshold {
		cb.state = "closed"
		atomic.StoreInt64(&cb.failureCount, 0)
	}

	return nil
}

// GetState returns the current state
func (cb *CircuitBreaker) GetState() string {
	return cb.state
}

// RateLimitByTenant implements per-tenant rate limiting
type TenantRateLimiter struct {
	tenantLimits map[string]*middleware.RateLimiter
}

// NewTenantRateLimiter creates a new tenant rate limiter
func NewTenantRateLimiter() *TenantRateLimiter {
	return &TenantRateLimiter{
		tenantLimits: make(map[string]*middleware.RateLimiter),
	}
}

// SetTenantLimit sets the rate limit for a tenant
func (trl *TenantRateLimiter) SetTenantLimit(tenantID string, maxRequests int, window time.Duration) {
	trl.tenantLimits[tenantID] = middleware.NewRateLimiter(maxRequests, window)
}

// Check checks if a request is allowed for a tenant
func (trl *TenantRateLimiter) Check(tenantID, identifier string) bool {
	limiter, ok := trl.tenantLimits[tenantID]
	if !ok {
		return true // No limit set
	}

	return limiter.Handle(identifier)
}

// HashString is a simple hash function for consistent routing
func HashString(s string) int64 {
	hash := int64(0)
	for _, c := range s {
		hash = hash*31 + int64(c)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}
