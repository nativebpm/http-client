package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreakerConfig holds configuration for circuit breaker middleware
type CircuitBreakerConfig struct {
	FailureThreshold int           // Number of failures to open the circuit
	SuccessThreshold int           // Number of successes to close the circuit from half-open
	Timeout          time.Duration // Time to wait before trying half-open
}

// DefaultCircuitBreakerConfig returns a default circuit breaker configuration
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: 5,
		SuccessThreshold: 2,
		Timeout:          10 * time.Second,
	}
}

// CircuitBreaker implements a simple circuit breaker
type CircuitBreaker struct {
	config          CircuitBreakerConfig
	state           CircuitBreakerState
	failureCount    int
	successCount    int
	lastFailureTime time.Time
	mu              sync.Mutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}
}

// Call executes the function with circuit breaker logic
func (cb *CircuitBreaker) Call(fn func() (*http.Response, error)) (*http.Response, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateOpen:
		if time.Since(cb.lastFailureTime) > cb.config.Timeout {
			cb.state = StateHalfOpen
			cb.successCount = 0
		} else {
			return nil, fmt.Errorf("circuit breaker is open")
		}
	case StateHalfOpen:
		// Allow request to test
	case StateClosed:
		// Allow request
	}

	resp, err := fn()

	if err != nil || (resp != nil && resp.StatusCode >= 500) {
		cb.failureCount++
		cb.lastFailureTime = time.Now()
		if cb.failureCount >= cb.config.FailureThreshold {
			cb.state = StateOpen
			fmt.Printf("Circuit breaker opened due to %d failures\n", cb.failureCount)
		}
		return resp, err
	}

	// Success
	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.config.SuccessThreshold {
			cb.state = StateClosed
			cb.failureCount = 0
			fmt.Printf("Circuit breaker closed after %d successes\n", cb.successCount)
		}
	} else {
		cb.failureCount = 0 // Reset on success in closed state
	}

	return resp, err
}

// CircuitBreakerMiddleware returns a middleware that implements circuit breaker
func CircuitBreakerMiddleware(cb *CircuitBreaker) func(http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return cb.Call(func() (*http.Response, error) {
				return next.RoundTrip(req)
			})
		})
	}
}
