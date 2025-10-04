package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	httpclient "github.com/nativebpm/http-client"
)

func main() {
	// Create a new HTTP client with fluent builder methods
	client, err := httpclient.NewClient(&http.Client{Timeout: 10 * time.Second}, "https://httpbin.org")
	if err != nil {
		log.Fatal(err)
	}

	// Use fluent builder methods for retry, rate limiting, and logging
	client.WithRetry().WithRateLimit(10, time.Minute/10).WithLogging()

	// Example requests
	for i := 0; i < 5; i++ {
		resp, err := client.GET(context.Background(), "/get").Send()
		if err != nil {
			log.Printf("Request %d failed: %v", i+1, err)
			continue
		}
		resp.Body.Close()
		fmt.Printf("Request %d succeeded: %d\n", i+1, resp.StatusCode)
		time.Sleep(100 * time.Millisecond) // Small delay to test rate limiting
	}
}
