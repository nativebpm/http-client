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
	_ = client.MultipartPOST(ctx, "/test")
	_ = client.MultipartPUT(ctx, "/test")
	_ = client.RequestGET(ctx, "/test")
	_ = client.RequestPOST(ctx, "/test")
	_ = client.RequestPUT(ctx, "/test")
	_ = client.RequestPATCH(ctx, "/test")
	_ = client.RequestDELETE(ctx, "/test")
}

func BenchmarkClientMethods(b *testing.B) {
	client, _ := NewClient(&http.Client{}, "http://example.com")
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.MultipartPOST(ctx, "/test")
		_ = client.MultipartPUT(ctx, "/test")
		_ = client.RequestGET(ctx, "/test")
		_ = client.RequestPOST(ctx, "/test")
		_ = client.RequestPUT(ctx, "/test")
		_ = client.RequestPATCH(ctx, "/test")
		_ = client.RequestDELETE(ctx, "/test")
	}
}
