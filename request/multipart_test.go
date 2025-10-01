package request

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestNewMultipart(t *testing.T) {
	m := NewMultipart(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")
	if m == nil {
		t.Fatal("expected non-nil multipart")
	}
}

func TestMultipartOpsCapacity(t *testing.T) {
	// Test default capacity
	m := NewMultipart(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")
	if cap(m.ops) != defaultOpsCapacity {
		t.Errorf("expected ops capacity %d, got %d", defaultOpsCapacity, cap(m.ops))
	}

	// Test custom capacity
	customCapacity := 64
	m2 := NewMultipartWithOpsCapacity(context.Background(), &http.Client{}, http.MethodPost, "http://example.com", customCapacity)
	if cap(m2.ops) != customCapacity {
		t.Errorf("expected ops capacity %d, got %d", customCapacity, cap(m2.ops))
	}
}

func TestMultipartOpsGrowth(t *testing.T) {
	m := NewMultipart(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")
	initialCap := cap(m.ops)

	// Add operations beyond initial capacity to test growth
	for i := 0; i < defaultOpsCapacity*2; i++ {
		m.Param(fmt.Sprintf("param-%d", i), "value")
	}

	// Capacity should have grown
	if cap(m.ops) <= initialCap {
		t.Errorf("expected ops capacity to grow beyond %d, got %d", initialCap, cap(m.ops))
	}
}

// Test performance with operations below capacity (no reallocation)
func BenchmarkMultipartWithinCapacity(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := NewMultipart(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")
		// Add 16 operations (half of defaultOpsCapacity=32)
		for j := 0; j < 16; j++ {
			m.Param(fmt.Sprintf("param-%d", j), "value")
		}
	}
}

// Test performance when exceeding capacity (triggers reallocation)
func BenchmarkMultipartExceedCapacity(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := NewMultipart(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")
		// Add 40 operations (exceeds defaultOpsCapacity=32)
		for j := 0; j < 40; j++ {
			m.Param(fmt.Sprintf("param-%d", j), "value")
		}
	}
}

// Compare default vs custom capacity
func BenchmarkMultipartCustomCapacity(b *testing.B) {
	b.Run("Default32", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := NewMultipart(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")
			for j := 0; j < 20; j++ {
				m.Param(fmt.Sprintf("param-%d", j), "value")
			}
		}
	})
	b.Run("Custom16", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := NewMultipartWithOpsCapacity(context.Background(), &http.Client{}, http.MethodPost, "http://example.com", 16)
			for j := 0; j < 20; j++ {
				m.Param(fmt.Sprintf("param-%d", j), "value")
			}
		}
	})
	b.Run("Custom64", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := NewMultipartWithOpsCapacity(context.Background(), &http.Client{}, http.MethodPost, "http://example.com", 64)
			for j := 0; j < 20; j++ {
				m.Param(fmt.Sprintf("param-%d", j), "value")
			}
		}
	})
}

// Test realistic multipart scenarios
func BenchmarkMultipartRealistic(b *testing.B) {
	b.Run("SmallForm", func(b *testing.B) {
		// Typical small form: 5-8 operations
		for i := 0; i < b.N; i++ {
			m := NewMultipart(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")
			m.Param("name", "John").Param("email", "john@example.com")
			m.Bool("subscribe", true).Float("score", 95.5)
			m.File("avatar", "avatar.jpg", bytes.NewReader([]byte("image data")))
		}
	})
	b.Run("LargeForm", func(b *testing.B) {
		// Large form: 25-30 operations
		for i := 0; i < b.N; i++ {
			m := NewMultipart(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")
			// User info
			m.Param("firstName", "John").Param("lastName", "Doe")
			m.Param("email", "john@example.com").Param("phone", "+1234567890")
			// Preferences
			m.Bool("newsletter", true).Bool("sms", false)
			m.Float("budget", 1000.50).Param("currency", "USD")
			// Multiple files
			m.File("avatar", "avatar.jpg", bytes.NewReader([]byte("avatar data")))
			m.File("resume", "resume.pdf", bytes.NewReader([]byte("resume data")))
			m.File("cover", "cover.txt", bytes.NewReader([]byte("cover letter")))
			// Additional fields
			for j := 0; j < 15; j++ {
				m.Param(fmt.Sprintf("custom-%d", j), "value")
			}
		}
	})
}
