package request

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
)

func TestNewRequest(t *testing.T) {
	r := NewRequest(context.Background(), &http.Client{}, http.MethodGet, "http://example.com")
	if r.Err() != nil {
		t.Fatalf("unexpected error: %v", r.Err())
	}
}

func BenchmarkNewRequest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewRequest(context.Background(), &http.Client{}, http.MethodGet, "http://example.com")
	}
}

func TestRequestMethods(t *testing.T) {
	r := NewRequest(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")
	r.Header("X-Test", "1").Param("foo", "bar").Bool("b", true).Float("f", 1.23)
	body := []byte(`{"a":1}`)
	r.Bytes(body)
	if r.Err() != nil {
		t.Fatalf("unexpected error: %v", r.Err())
	}
	r.JSON(map[string]any{"foo": "bar"})
	if r.Err() != nil {
		t.Fatalf("unexpected error: %v", r.Err())
	}
	r.Body(io.NopCloser(bytes.NewReader([]byte("test"))))
}

func BenchmarkRequestMethods(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := NewRequest(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")
		r.Header("X-Test", "1").Param("foo", "bar").Bool("b", true).Float("f", 1.23)
		_ = r.Bytes([]byte(`{"a":1}`))
		_ = r.JSON(map[string]any{"foo": "bar"})
		_ = r.Body(io.NopCloser(bytes.NewReader([]byte("test"))))
	}
}
