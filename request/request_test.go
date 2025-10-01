package request

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestNewRequest(t *testing.T) {
	r := NewRequest(context.Background(), &http.Client{}, http.MethodGet, "http://example.com")
	if r == nil {
		t.Fatal("expected non-nil request")
	}
}

func TestRequestOpsCapacity(t *testing.T) {
	// Test default capacity
	r := NewRequest(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")
	if cap(r.ops) != defaultOpsCapacity {
		t.Errorf("expected ops capacity %d, got %d", defaultOpsCapacity, cap(r.ops))
	}

	// Test custom capacity
	customCapacity := 64
	r2 := NewRequestWithOpsCapacity(context.Background(), &http.Client{}, http.MethodPost, "http://example.com", customCapacity)
	if cap(r2.ops) != customCapacity {
		t.Errorf("expected ops capacity %d, got %d", customCapacity, cap(r2.ops))
	}
}

func TestRequestOpsGrowth(t *testing.T) {
	r := NewRequest(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")
	initialCap := cap(r.ops)

	// Add operations beyond initial capacity to test growth
	for i := 0; i < defaultOpsCapacity*2; i++ {
		r.Header(fmt.Sprintf("X-Test-%d", i), "value")
	}

	// Capacity should have grown
	if cap(r.ops) <= initialCap {
		t.Errorf("expected ops capacity to grow beyond %d, got %d", initialCap, cap(r.ops))
	}
}

// Test performance with operations below capacity (no reallocation)
func BenchmarkRequestWithinCapacity(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := NewRequest(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")
		// Add 16 operations (half of defaultOpsCapacity=32)
		for j := 0; j < 16; j++ {
			r.Header(fmt.Sprintf("X-Header-%d", j), "value")
		}
	}
}

// Test performance when exceeding capacity (triggers reallocation)
func BenchmarkRequestExceedCapacity(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := NewRequest(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")
		// Add 40 operations (exceeds defaultOpsCapacity=32)
		for j := 0; j < 40; j++ {
			r.Header(fmt.Sprintf("X-Header-%d", j), "value")
		}
	}
}

// Compare default vs custom capacity
func BenchmarkRequestCustomCapacity(b *testing.B) {
	b.Run("Default32", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r := NewRequest(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")
			for j := 0; j < 20; j++ {
				r.Header(fmt.Sprintf("X-Header-%d", j), "value")
			}
		}
	})
	b.Run("Custom16", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r := NewRequestWithOpsCapacity(context.Background(), &http.Client{}, http.MethodPost, "http://example.com", 16)
			for j := 0; j < 20; j++ {
				r.Header(fmt.Sprintf("X-Header-%d", j), "value")
			}
		}
	})
	b.Run("Custom64", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r := NewRequestWithOpsCapacity(context.Background(), &http.Client{}, http.MethodPost, "http://example.com", 64)
			for j := 0; j < 20; j++ {
				r.Header(fmt.Sprintf("X-Header-%d", j), "value")
			}
		}
	})
}
