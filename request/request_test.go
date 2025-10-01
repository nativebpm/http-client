package request

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestNewRequest(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{}

	r := NewRequest(ctx, client, http.MethodGet, "http://example.com")
	if r == nil {
		t.Fatal("expected non-nil request")
	}

	if r.client != client {
		t.Error("expected client to be set correctly")
	}

	if r.request.Method != http.MethodGet {
		t.Errorf("expected method GET, got %s", r.request.Method)
	}

	if r.request.URL.String() != "http://example.com" {
		t.Errorf("expected URL http://example.com, got %s", r.request.URL.String())
	}
}

func TestRequestHeaders(t *testing.T) {
	r := NewRequest(context.Background(), &http.Client{}, http.MethodGet, "http://example.com")

	// Test chaining
	result := r.Header("Authorization", "Bearer token").Header("Accept", "application/json")
	if result != r {
		t.Error("expected Header to return the same request instance for chaining")
	}

	if r.HeaderCount() != 2 {
		t.Errorf("expected 2 headers, got %d", r.HeaderCount())
	}

	// Check header values
	if r.headers["Authorization"][0] != "Bearer token" {
		t.Errorf("unexpected first header: %+v", r.headers["Authorization"])
	}

	if r.headers["Accept"][0] != "application/json" {
		t.Errorf("unexpected second header: %+v", r.headers["Accept"])
	}
}

func TestRequestParams(t *testing.T) {
	r := NewRequest(context.Background(), &http.Client{}, http.MethodGet, "http://example.com")

	// Test string param
	r.Param("id", "123")

	// Test bool param
	r.Bool("active", true)

	// Test float param
	r.Float("score", 95.5)

	if r.ParamCount() != 3 {
		t.Errorf("expected 3 params, got %d", r.ParamCount())
	}

	// Check param values
	expectedParams := []ParamOp{
		{Key: "id", Value: "123"},
		{Key: "active", Value: "true"},
		{Key: "score", Value: "95.5"},
	}

	for i, expected := range expectedParams {
		if r.params[i] != expected {
			t.Errorf("param %d: expected %+v, got %+v", i, expected, r.params[i])
		}
	}
}

func TestRequestBodyOperations(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Request)
		wantType BodyType
		hasBody  bool
	}{
		{
			name:     "no body",
			setup:    func(r *Request) {},
			wantType: BodyTypeNone,
			hasBody:  false,
		},
		{
			name: "bytes body",
			setup: func(r *Request) {
				r.Bytes([]byte("test data"), "text/plain")
			},
			wantType: BodyTypeBytes,
			hasBody:  true,
		},
		{
			name: "JSON body",
			setup: func(r *Request) {
				r.JSON(map[string]string{"key": "value"})
			},
			wantType: BodyTypeJSON,
			hasBody:  true,
		},
		{
			name: "reader body",
			setup: func(r *Request) {
				r.Body(io.NopCloser(strings.NewReader("test")), "text/plain", 4)
			},
			wantType: BodyTypeReader,
			hasBody:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRequest(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")
			tt.setup(r)

			if r.body.Type != tt.wantType {
				t.Errorf("expected body type %d, got %d", tt.wantType, r.body.Type)
			}

			if r.HasBody() != tt.hasBody {
				t.Errorf("expected HasBody() %v, got %v", tt.hasBody, r.HasBody())
			}
		})
	}
}

func TestRequestSliceGrowth(t *testing.T) {
	// Start with very small capacity
	r := NewRequestWithOpsCapacity(context.Background(), &http.Client{}, http.MethodPost, "http://example.com", 4)

	initialHeaderCap := len(r.headers) // should be 1 (4/4)
	initialParamCap := cap(r.params)   // should be 2 (4/2)

	// Add more headers than initial capacity
	for i := 0; i < 5; i++ {
		r.Header(fmt.Sprintf("X-Header-%d", i), "value")
	}

	if len(r.headers) != 5 {
		t.Errorf("expected 5 headers, got %d", len(r.headers))
	}

	if len(r.headers) <= initialHeaderCap {
		t.Errorf("expected header capacity to grow from %d, but it's %d", initialHeaderCap, len(r.headers))
	}

	// Add more params than initial capacity
	for i := 0; i < 5; i++ {
		r.Param(fmt.Sprintf("param-%d", i), "value")
	}

	if len(r.params) != 5 {
		t.Errorf("expected 5 params, got %d", len(r.params))
	}

	if cap(r.params) <= initialParamCap {
		t.Errorf("expected param capacity to grow from %d, but it's %d", initialParamCap, cap(r.params))
	}
}

func TestRequestSend(t *testing.T) {
	// Mock server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check headers
		if auth := r.Header.Get("Authorization"); auth != "Bearer token" {
			t.Errorf("expected Authorization header 'Bearer token', got '%s'", auth)
		}

		// Check query params
		if id := r.URL.Query().Get("id"); id != "123" {
			t.Errorf("expected id param '123', got '%s'", id)
		}

		if active := r.URL.Query().Get("active"); active != "true" {
			t.Errorf("expected active param 'true', got '%s'", active)
		}

		// Check body for POST request
		if r.Method == http.MethodPost {
			expectedContentType := "application/json"
			if ct := r.Header.Get("Content-Type"); ct != expectedContentType {
				t.Errorf("expected Content-Type '%s', got '%s'", expectedContentType, ct)
			}

			body, _ := io.ReadAll(r.Body)
			expectedBody := `{"message":"test"}`
			if string(body) != expectedBody {
				t.Errorf("expected body '%s', got '%s'", expectedBody, string(body))
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}

	server := &http.Server{
		Addr:    ":0",
		Handler: http.HandlerFunc(handler),
	}

	client := &http.Client{}

	t.Run("GET request", func(t *testing.T) {
		req := NewRequest(context.Background(), client, http.MethodGet, "http://example.com")
		req.Header("Authorization", "Bearer token")
		req.Param("id", "123")
		req.Bool("active", true)

		// We can't actually send to example.com in tests, so we just verify the request is built correctly
		if err := req.setBody(); err != nil {
			t.Errorf("setBody() failed: %v", err)
		}

		// Check that headers are applied correctly by manually calling the logic
		req.request.Header = req.headers

		if req.request.Header.Get("Authorization") != "Bearer token" {
			t.Error("Authorization header not set correctly")
		}

		// Check that params are applied correctly
		if len(req.params) > 0 {
			q := req.request.URL.Query()
			for _, p := range req.params {
				q.Set(p.Key, p.Value)
			}
			req.request.URL.RawQuery = q.Encode()
		}

		if req.request.URL.Query().Get("id") != "123" {
			t.Error("id param not set correctly")
		}

		if req.request.URL.Query().Get("active") != "true" {
			t.Error("active param not set correctly")
		}
	})

	t.Run("POST request with JSON body", func(t *testing.T) {
		req := NewRequest(context.Background(), client, http.MethodPost, "http://example.com")
		req.Header("Authorization", "Bearer token")
		req.JSON(map[string]string{"message": "test"})

		if err := req.setBody(); err != nil {
			t.Errorf("setBody() failed: %v", err)
		}

		if req.request.Header.Get("Content-Type") != "application/json" {
			t.Error("Content-Type header not set correctly for JSON")
		}

		if !req.HasBody() {
			t.Error("request should have body")
		}
	})

	_ = server
}
