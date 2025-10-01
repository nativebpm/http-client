package request

import (
	"bytes"
	"context"
	"net/http"
	"testing"
)

func TestNewMultipart(t *testing.T) {
	m := NewMultipart(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")
	if m == nil {
		t.Fatal("expected non-nil multipart")
	}
}

func BenchmarkNewMultipart(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewMultipart(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")
	}
}

func TestMultipartMethods(t *testing.T) {
	m := NewMultipart(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")
	m.Header("X-Test", "1").Param("foo", "bar").Bool("b", true).Float("f", 1.23)
	m.File("file", "test.txt", bytes.NewReader([]byte("test content")))
}

func BenchmarkMultipartMethods(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := NewMultipart(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")
		m.Header("X-Test", "1").Param("foo", "bar").Bool("b", true).Float("f", 1.23)
		m.File("file", "test.txt", bytes.NewReader([]byte("test content")))
	}
}
