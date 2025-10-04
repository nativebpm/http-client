package httpclient

import (
	"context"
	"net/http"
	"testing"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient(&http.Client{}, "http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.baseURL.String() != "http://example.com" {
		t.Errorf("unexpected baseURL: %s", client.baseURL.String())
	}
}

func BenchmarkNewClient(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewClient(&http.Client{}, "http://example.com")
	}
}

func TestClientMethods(t *testing.T) {
	client, _ := NewClient(&http.Client{}, "http://example.com")
	ctx := context.Background()

	// Test multipart methods
	_ = client.Multipart(ctx, "/test")
	_ = client.MultipartRequest(ctx, "/test", http.MethodPut)

	// Test request methods
	_ = client.Request(ctx, http.MethodGet, "/test")
	_ = client.Request(ctx, http.MethodPost, "/test")

	// Test convenience methods
	_ = client.GET(ctx, "/test")
	_ = client.POST(ctx, "/test")
	_ = client.PUT(ctx, "/test")
	_ = client.PATCH(ctx, "/test")
	_ = client.DELETE(ctx, "/test")
}

func BenchmarkClientMethods(b *testing.B) {
	client, _ := NewClient(&http.Client{}, "http://example.com")
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.Multipart(ctx, "/test")
		_ = client.MultipartRequest(ctx, "/test", http.MethodPut)
		_ = client.GET(ctx, "/test")
		_ = client.POST(ctx, "/test")
		_ = client.PUT(ctx, "/test")
		_ = client.PATCH(ctx, "/test")
		_ = client.DELETE(ctx, "/test")
	}
}
