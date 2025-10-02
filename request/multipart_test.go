package request_test

import (
	"bytes"
	"context"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nativebpm/http-client/request"
)

func TestNewMultipart(t *testing.T) {
	client := &http.Client{}
	ctx := context.Background()

	mp := request.NewMultipart(ctx, client, http.MethodPost, "http://example.com/upload")
	if mp == nil {
		t.Fatal("NewMultipart returned nil")
	}
}

func TestMultipart_Param(t *testing.T) {
	receivedFields := make(map[string]string)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		mediaType, params, err := mime.ParseMediaType(contentType)
		if err != nil {
			t.Errorf("failed to parse media type: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if mediaType != "multipart/form-data" {
			t.Errorf("expected multipart/form-data, got %s", mediaType)
			http.Error(w, "invalid content type", http.StatusBadRequest)
			return
		}

		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Errorf("failed to read part: %v", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			data, err := io.ReadAll(p)
			if err != nil {
				t.Errorf("failed to read part data: %v", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			receivedFields[p.FormName()] = string(data)
			p.Close()
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{}
	ctx := context.Background()

	resp, err := request.NewMultipart(ctx, client, http.MethodPost, server.URL).
		Param("name", "John Doe").
		Param("email", "john@example.com").
		Send()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if receivedFields["name"] != "John Doe" {
		t.Errorf("expected name=John Doe, got %s", receivedFields["name"])
	}
	if receivedFields["email"] != "john@example.com" {
		t.Errorf("expected email=john@example.com, got %s", receivedFields["email"])
	}
}

func TestMultipart_TypedFields(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*request.Multipart) *request.Multipart
		expected map[string]string
	}{
		{
			name: "bool_true",
			setup: func(mp *request.Multipart) *request.Multipart {
				return mp.Bool("active", true)
			},
			expected: map[string]string{"active": "true"},
		},
		{
			name: "bool_false",
			setup: func(mp *request.Multipart) *request.Multipart {
				return mp.Bool("active", false)
			},
			expected: map[string]string{"active": "false"},
		},
		{
			name: "float",
			setup: func(mp *request.Multipart) *request.Multipart {
				return mp.Float("price", 19.99)
			},
			expected: map[string]string{"price": "19.99"},
		},
		{
			name: "int",
			setup: func(mp *request.Multipart) *request.Multipart {
				return mp.Int("count", 42)
			},
			expected: map[string]string{"count": "42"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receivedFields := make(map[string]string)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				contentType := r.Header.Get("Content-Type")
				mediaType, params, err := mime.ParseMediaType(contentType)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if mediaType != "multipart/form-data" {
					http.Error(w, "invalid content type", http.StatusBadRequest)
					return
				}

				mr := multipart.NewReader(r.Body, params["boundary"])
				for {
					p, err := mr.NextPart()
					if err == io.EOF {
						break
					}
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
					data, err := io.ReadAll(p)
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
					receivedFields[p.FormName()] = string(data)
					p.Close()
				}

				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := &http.Client{}
			ctx := context.Background()

			mp := request.NewMultipart(ctx, client, http.MethodPost, server.URL)
			resp, err := tt.setup(mp).Send()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected status 200, got %d", resp.StatusCode)
			}

			for key, expected := range tt.expected {
				if receivedFields[key] != expected {
					t.Errorf("expected %s=%s, got %s", key, expected, receivedFields[key])
				}
			}
		})
	}
}

func TestMultipart_File(t *testing.T) {
	receivedFiles := make(map[string][]byte)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		mediaType, params, err := mime.ParseMediaType(contentType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if mediaType != "multipart/form-data" {
			http.Error(w, "invalid content type", http.StatusBadRequest)
			return
		}

		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			data, err := io.ReadAll(p)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			receivedFiles[p.FormName()] = data
			p.Close()
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{}
	ctx := context.Background()

	fileContent := []byte("Hello, World!")
	resp, err := request.NewMultipart(ctx, client, http.MethodPost, server.URL).
		File("document", "test.txt", bytes.NewReader(fileContent)).
		Send()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if !bytes.Equal(receivedFiles["document"], fileContent) {
		t.Errorf("expected file content %q, got %q", fileContent, receivedFiles["document"])
	}
}

func TestMultipart_MixedParamsAndFiles(t *testing.T) {
	receivedFields := make(map[string]string)
	receivedFiles := make(map[string][]byte)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		mediaType, params, err := mime.ParseMediaType(contentType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if mediaType != "multipart/form-data" {
			http.Error(w, "invalid content type", http.StatusBadRequest)
			return
		}

		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			data, err := io.ReadAll(p)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if p.FileName() != "" {
				receivedFiles[p.FormName()] = data
			} else {
				receivedFields[p.FormName()] = string(data)
			}
			p.Close()
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{}
	ctx := context.Background()

	fileContent := []byte("PDF content here")
	resp, err := request.NewMultipart(ctx, client, http.MethodPost, server.URL).
		Param("title", "My Document").
		Int("version", 2).
		Bool("published", true).
		File("pdf", "document.pdf", bytes.NewReader(fileContent)).
		Param("author", "Jane Doe").
		Send()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	expectedFields := map[string]string{
		"title":     "My Document",
		"version":   "2",
		"published": "true",
		"author":    "Jane Doe",
	}

	for key, expected := range expectedFields {
		if receivedFields[key] != expected {
			t.Errorf("expected %s=%s, got %s", key, expected, receivedFields[key])
		}
	}

	if !bytes.Equal(receivedFiles["pdf"], fileContent) {
		t.Errorf("expected file content %q, got %q", fileContent, receivedFiles["pdf"])
	}
}

func TestMultipart_Header(t *testing.T) {
	var receivedHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("X-API-Key")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{}
	ctx := context.Background()

	resp, err := request.NewMultipart(ctx, client, http.MethodPost, server.URL).
		Header("X-API-Key", "secret-token-123").
		Param("data", "value").
		Send()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if receivedHeader != "secret-token-123" {
		t.Errorf("expected header X-API-Key=secret-token-123, got %s", receivedHeader)
	}
}

func TestMultipart_LargeFile(t *testing.T) {
	receivedSize := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		mediaType, params, err := mime.ParseMediaType(contentType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if mediaType != "multipart/form-data" {
			http.Error(w, "invalid content type", http.StatusBadRequest)
			return
		}

		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			n, err := io.Copy(io.Discard, p)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			receivedSize += int(n)
			p.Close()
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{}
	ctx := context.Background()

	// Create a 1MB file content
	largeContent := bytes.Repeat([]byte("x"), 1024*1024)
	resp, err := request.NewMultipart(ctx, client, http.MethodPost, server.URL).
		File("largefile", "large.dat", bytes.NewReader(largeContent)).
		Send()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if receivedSize != len(largeContent) {
		t.Errorf("expected file size %d, got %d", len(largeContent), receivedSize)
	}
}

func TestMultipart_ContextCancellation(t *testing.T) {
	blockCh := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-blockCh // Block until test cleanup
		w.WriteHeader(http.StatusOK)
	}))
	defer func() {
		close(blockCh)
		server.Close()
	}()

	client := &http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := request.NewMultipart(ctx, client, http.MethodPost, server.URL).
		Param("data", "value").
		Send()

	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}

	if !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("expected context deadline exceeded error, got: %v", err)
	}
}

func TestMultipart_MultipleFiles(t *testing.T) {
	receivedFiles := make(map[string][]byte)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		mediaType, params, err := mime.ParseMediaType(contentType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if mediaType != "multipart/form-data" {
			http.Error(w, "invalid content type", http.StatusBadRequest)
			return
		}

		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			data, err := io.ReadAll(p)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			key := p.FormName() + ":" + p.FileName()
			receivedFiles[key] = data
			p.Close()
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{}
	ctx := context.Background()

	file1 := []byte("content1")
	file2 := []byte("content2")
	file3 := []byte("content3")

	resp, err := request.NewMultipart(ctx, client, http.MethodPost, server.URL).
		File("file", "doc1.txt", bytes.NewReader(file1)).
		File("file", "doc2.txt", bytes.NewReader(file2)).
		File("attachment", "doc3.txt", bytes.NewReader(file3)).
		Send()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if !bytes.Equal(receivedFiles["file:doc1.txt"], file1) {
		t.Errorf("file1 content mismatch")
	}
	if !bytes.Equal(receivedFiles["file:doc2.txt"], file2) {
		t.Errorf("file2 content mismatch")
	}
	if !bytes.Equal(receivedFiles["attachment:doc3.txt"], file3) {
		t.Errorf("file3 content mismatch")
	}
}

func TestMultipart_EmptyForm(t *testing.T) {
	partCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		mediaType, params, err := mime.ParseMediaType(contentType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if mediaType != "multipart/form-data" {
			http.Error(w, "invalid content type", http.StatusBadRequest)
			return
		}

		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			_, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			partCount++
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{}
	ctx := context.Background()

	resp, err := request.NewMultipart(ctx, client, http.MethodPost, server.URL).Send()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if partCount != 0 {
		t.Errorf("expected 0 parts, got %d", partCount)
	}
}

func TestMultipart_ChainedCalls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{}
	ctx := context.Background()

	// Test that all methods return *Multipart for chaining
	mp := request.NewMultipart(ctx, client, http.MethodPost, server.URL).
		Header("X-Custom", "value").
		Param("key1", "value1").
		Bool("flag", true).
		Float("price", 9.99).
		Int("count", 5).
		File("doc", "file.txt", strings.NewReader("content")).
		Param("key2", "value2")

	resp, err := mp.Send()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}
