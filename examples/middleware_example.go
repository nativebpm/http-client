package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	httpclient "github.com/nativebpm/http-client"
)

// roundTripperFunc is a helper type to implement http.RoundTripper
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func main() {
	// Create a new HTTP client
	client, err := httpclient.NewClient(&http.Client{Timeout: 10 * time.Second}, "https://httpbin.org")
	if err != nil {
		log.Fatal(err)
	}

	// Add logging middleware
	client.Use(func(next http.RoundTripper) http.RoundTripper {
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
	})

	// Add authorization header middleware
	client.Use(func(next http.RoundTripper) http.RoundTripper {
		return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			req.Header.Set("Authorization", "Bearer example-token")
			return next.RoundTrip(req)
		})
	})

	// Example request
	resp, err := client.GET(context.Background(), "/get").Send()
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Printf("Final status: %d\n", resp.StatusCode)
}
