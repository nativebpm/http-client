package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	httpclient "github.com/nativebpm/http-client"
)

func main() {
	// Create a new HTTP client
	client, err := httpclient.NewClient(&http.Client{}, "https://jsonplaceholder.typicode.com")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Example 1: Simple path parameter
	fmt.Println("Example 1: Get user by ID")
	resp, err := client.GET(ctx, "/users/{id}").
		PathInt("id", 1).
		Send()
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Status: %d\nResponse: %s\n\n", resp.StatusCode, string(body))
	}

	// Example 2: Multiple path parameters
	fmt.Println("Example 2: Get user's post")
	resp, err = client.GET(ctx, "/users/{userId}/posts").
		PathInt("userId", 1).
		Send()
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Status: %d\nResponse (truncated): %.200s...\n\n", resp.StatusCode, string(body))
	}

	// Example 3: Path parameters with query parameters
	fmt.Println("Example 3: Get post with path and query params")
	resp, err = client.GET(ctx, "/posts/{id}").
		PathInt("id", 1).
		Param("_limit", "1").
		Send()
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Status: %d\nResponse: %s\n\n", resp.StatusCode, string(body))
	}

	// Example 4: Complex RESTful path
	fmt.Println("Example 4: Complex RESTful path")
	resp, err = client.GET(ctx, "/users/{userId}/posts/{postId}/comments").
		PathParam("userId", "1").
		PathInt("postId", 1).
		Int("_limit", 2).
		Send()
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Status: %d\nResponse (truncated): %.300s...\n\n", resp.StatusCode, string(body))
	}

	// Example 5: POST with path parameters and JSON body
	fmt.Println("Example 5: POST with path parameters and JSON body")
	newPost := map[string]interface{}{
		"title":  "My New Post",
		"body":   "This is the content of my post",
		"userId": 1,
	}
	resp, err = client.POST(ctx, "/users/{userId}/posts").
		PathInt("userId", 1).
		Header("Content-Type", "application/json").
		JSON(newPost).
		Send()
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Status: %d\nResponse: %s\n\n", resp.StatusCode, string(body))
	}

	// Example 6: Typed path parameters
	fmt.Println("Example 6: Different typed path parameters")
	// Using boolean and float path parameters (for demonstration)
	localClient, _ := httpclient.NewClient(&http.Client{}, "http://example.com")

	// This demonstrates the API - would work with a real API that accepts these types
	_ = localClient.GET(ctx, "/products/{id}/available/{inStock}/price/{amount}").
		PathInt("id", 42).
		PathBool("inStock", true).
		PathFloat("amount", 99.99)

	fmt.Println("Typed path parameters: /products/42/available/true/price/99.99")
	fmt.Println("(Example URL only - not sent to real API)")
}
