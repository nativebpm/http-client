package middleware

import (
	"net/http"
)

// roundTripperFunc is a helper to implement RoundTripper
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
