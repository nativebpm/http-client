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

func TestMultipartOperations(t *testing.T) {
	m := NewMultipart(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")

	// Test parameter operations
	m.Param("name", "John")
	m.Bool("active", true)
	m.Float("score", 95.5)

	if m.ParamCount() != 3 {
		t.Errorf("expected 3 params, got %d", m.ParamCount())
	}

	// Test file operations
	m.File("avatar", "avatar.jpg", bytes.NewReader([]byte("image data")))
	m.File("document", "doc.pdf", bytes.NewReader([]byte("pdf data")))

	if m.FileCount() != 2 {
		t.Errorf("expected 2 files, got %d", m.FileCount())
	}

	if m.TotalOps() != 5 {
		t.Errorf("expected 5 total operations, got %d", m.TotalOps())
	}
}

func TestMultipartCapacity(t *testing.T) {
	// Test with custom capacity
	m := NewMultipartWithOpsCapacity(context.Background(), &http.Client{}, http.MethodPost, "http://example.com", 64)

	// Add many params to test capacity
	for i := 0; i < 20; i++ {
		m.Param(fmt.Sprintf("param-%d", i), "value")
	}

	if m.ParamCount() != 20 {
		t.Errorf("expected 20 params, got %d", m.ParamCount())
	}
}

func TestMultipartGrowth(t *testing.T) {
	// Start with small capacity and test growth
	m := NewMultipartWithOpsCapacity(context.Background(), &http.Client{}, http.MethodPost, "http://example.com", 8)

	// Add more params than initial capacity
	for i := 0; i < 20; i++ {
		m.Param(fmt.Sprintf("param-%d", i), "value")
	}

	if m.ParamCount() != 20 {
		t.Errorf("expected 20 params after growth, got %d", m.ParamCount())
	}
}

// Test performance with operations below capacity
func BenchmarkMultipartWithinCapacity(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := NewMultipart(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")
		// Add 10 operations (within typical capacity)
		for j := 0; j < 10; j++ {
			m.Param(fmt.Sprintf("param-%d", j), "value")
		}
	}
}

// Test performance when exceeding capacity
func BenchmarkMultipartExceedCapacity(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := NewMultipart(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")
		// Add 40 operations (exceeds typical capacity)
		for j := 0; j < 40; j++ {
			m.Param(fmt.Sprintf("param-%d", j), "value")
		}
	}
}

// Compare different capacity settings
func BenchmarkMultipartCapacity(b *testing.B) {
	b.Run("Default", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := NewMultipart(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")
			for j := 0; j < 15; j++ {
				m.Param(fmt.Sprintf("param-%d", j), "value")
			}
		}
	})
	b.Run("Small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := NewMultipartWithOpsCapacity(context.Background(), &http.Client{}, http.MethodPost, "http://example.com", 16)
			for j := 0; j < 15; j++ {
				m.Param(fmt.Sprintf("param-%d", j), "value")
			}
		}
	})
	b.Run("Large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := NewMultipartWithOpsCapacity(context.Background(), &http.Client{}, http.MethodPost, "http://example.com", 64)
			for j := 0; j < 15; j++ {
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
	b.Run("ZeroAlloc", func(b *testing.B) {
		// Test zero-allocation approach
		for i := 0; i < b.N; i++ {
			m := NewMultipart(context.Background(), &http.Client{}, http.MethodPost, "http://example.com")
			m.Param("name", "John")
			m.Param("email", "john@example.com")
			m.Bool("active", true)
			m.Float("score", 98.6)
			m.File("file", "test.txt", bytes.NewReader([]byte("test content")))
		}
	})
}
