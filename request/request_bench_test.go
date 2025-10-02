package request_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nativebpm/http-client/request"
)

func BenchmarkRequest_Simple(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{}
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		resp, err := request.NewRequest(ctx, client, http.MethodGet, server.URL).
			Header("X-API-Key", "secret").
			Param("page", "1").
			Send()
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}

func BenchmarkRequest_ManyParams(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{}
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := request.NewRequest(ctx, client, http.MethodGet, server.URL)
		for j := 0; j < 10; j++ {
			req.Param("param", "value")
		}
		resp, err := req.Send()
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}

func BenchmarkRequest_JSON(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{}
	ctx := context.Background()

	data := map[string]any{
		"name":  "John Doe",
		"age":   30,
		"email": "john@example.com",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		resp, err := request.NewRequest(ctx, client, http.MethodPost, server.URL).
			JSON(data).
			Send()
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}
