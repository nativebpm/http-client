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

	// Use fluent builder methods for circuit breaker and logging
	client.WithCircuitBreaker().WithLogging()

	// Example requests - simulate failures to trigger circuit breaker
	for i := 0; i < 10; i++ {
		// Use a non-existent endpoint to simulate failures
		resp, err := client.GET(context.Background(), "/status/500").Send()
		if err != nil {
			log.Printf("Request %d failed: %v", i+1, err)
			continue
		}
		resp.Body.Close()
		fmt.Printf("Request %d succeeded: %d\n", i+1, resp.StatusCode)

		time.Sleep(1 * time.Second) // Wait a bit between requests
	}

	// Wait for circuit breaker to potentially go to half-open
	fmt.Println("Waiting for circuit breaker timeout...")
	time.Sleep(12 * time.Second)

	// Try again - should allow some requests
	for i := 10; i < 15; i++ {
		resp, err := client.GET(context.Background(), "/get").Send() // Use valid endpoint
		if err != nil {
			log.Printf("Request %d failed: %v", i+1, err)
			continue
		}
		resp.Body.Close()
		fmt.Printf("Request %d succeeded: %d\n", i+1, resp.StatusCode)

		time.Sleep(1 * time.Second)
	}
}
