package middleware

import (
	"fmt"
	"net/http"
	"time"
)

// LoggingMiddleware returns a middleware that logs requests and responses
func LoggingMiddleware() func(http.RoundTripper) http.RoundTripper {
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			fmt.Printf("Sending request: %s %s\n", req.Method, req.URL)
			start := time.Now()
			resp, err := next.RoundTrip(req)
			if err != nil {
				fmt.Printf("Request failed: %v\n", err)
				return nil, err
			}
			fmt.Printf("Response received: %d in %v\n", resp.StatusCode, time.Since(start))
			return resp, nil
		})
	}
}
