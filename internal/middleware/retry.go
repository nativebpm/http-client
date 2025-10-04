package middleware

import (
	"fmt"
	"net/http"
	"time"
)

// RetryConfig holds configuration for retry middleware
type RetryConfig struct {
	MaxRetries     int
	Backoff        func(attempt int) time.Duration
	RetryCondition func(resp *http.Response, err error) bool
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		Backoff: func(attempt int) time.Duration {
			return time.Duration(attempt) * time.Second
		},
		RetryCondition: func(resp *http.Response, err error) bool {
			if err != nil {
				return true
			}
			if resp != nil && resp.StatusCode >= 500 {
				return true
			}
			return false
		},
	}
}

// RetryMiddleware returns a middleware that retries requests based on the config
func RetryMiddleware(config RetryConfig) func(http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			var lastResp *http.Response
			var lastErr error

			for attempt := 0; attempt <= config.MaxRetries; attempt++ {
				if attempt > 0 {
					delay := config.Backoff(attempt - 1)
					fmt.Printf("Retrying request in %v (attempt %d/%d)\n", delay, attempt, config.MaxRetries)
					time.Sleep(delay)
				}

				// Clone the request for retry
				reqCopy := req.Clone(req.Context())
				lastResp, lastErr = next.RoundTrip(reqCopy)

				if !config.RetryCondition(lastResp, lastErr) {
					break
				}

				if lastResp != nil {
					lastResp.Body.Close() // Close body before retry
				}
			}

			return lastResp, lastErr
		})
	}
}
