package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a basic token bucket rate limiter
type RateLimiter struct {
	tokens     int
	capacity   int
	refillRate time.Duration
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(capacity int, refillRate time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:     capacity,
		capacity:   capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed and consumes a token
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)
	tokensToAdd := int(elapsed / rl.refillRate)
	if tokensToAdd > 0 {
		rl.tokens = min(rl.capacity, rl.tokens+tokensToAdd)
		rl.lastRefill = now
	}

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	return false
}

// RateLimitMiddleware returns a middleware that enforces rate limiting
func RateLimitMiddleware(limiter *RateLimiter) func(http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			if !limiter.Allow() {
				return nil, fmt.Errorf("rate limit exceeded")
			}

			fmt.Printf("Request allowed by rate limiter\n")
			return next.RoundTrip(req)
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
