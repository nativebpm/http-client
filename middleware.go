package httpclient

import (
	"net/http"
	"time"

	"github.com/nativebpm/http-client/internal/middleware"
)

// Re-export middleware types and functions for public API

// RetryConfig holds configuration for retry middleware
type RetryConfig = middleware.RetryConfig

// RateLimiter implements a basic token bucket rate limiter
type RateLimiter = middleware.RateLimiter

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState = middleware.CircuitBreakerState

// CircuitBreakerConfig holds configuration for circuit breaker middleware
type CircuitBreakerConfig = middleware.CircuitBreakerConfig

// CircuitBreaker implements a simple circuit breaker
type CircuitBreaker = middleware.CircuitBreaker

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() RetryConfig {
	return middleware.DefaultRetryConfig()
}

// RetryMiddleware returns a middleware that retries requests based on the config
func RetryMiddleware(config RetryConfig) func(http.RoundTripper) http.RoundTripper {
	return middleware.RetryMiddleware(config)
}

// NewSimpleRateLimiter creates a new rate limiter
func NewSimpleRateLimiter(capacity int, refillRate time.Duration) *RateLimiter {
	return middleware.NewRateLimiter(capacity, refillRate)
}

// RateLimitMiddleware returns a middleware that enforces rate limiting
func RateLimitMiddleware(limiter *RateLimiter) func(http.RoundTripper) http.RoundTripper {
	return middleware.RateLimitMiddleware(limiter)
}

// LoggingMiddleware returns a middleware that logs requests and responses
func LoggingMiddleware() func(http.RoundTripper) http.RoundTripper {
	return middleware.LoggingMiddleware()
}

// DefaultCircuitBreakerConfig returns a default circuit breaker configuration
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return middleware.DefaultCircuitBreakerConfig()
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return middleware.NewCircuitBreaker(config)
}

// CircuitBreakerMiddleware returns a middleware that implements circuit breaker
func CircuitBreakerMiddleware(cb *CircuitBreaker) func(http.RoundTripper) http.RoundTripper {
	return middleware.CircuitBreakerMiddleware(cb)
}
