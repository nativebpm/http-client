package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestNewClientInvalidURL(t *testing.T) {
	_, err := NewClient(&http.Client{}, "://bad url")
	if err == nil {
		t.Fatalf("expected error when parsing invalid URL")
	}
}

func TestNewClientInitializesCorrectly(t *testing.T) {
	httpClient := &http.Client{}
	client, err := NewClient(httpClient, "https://example.com/api")
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}

	if client.client != httpClient {
		t.Fatalf("expected client to wrap provided http.Client instance")
	}
	if client.baseURL.String() != "https://example.com/api" {
		t.Fatalf("unexpected base URL: %s", client.baseURL)
	}
}

func TestRequestJSONBody(t *testing.T) {
	t.Parallel()

	type recorded struct {
		Method        string
		Path          string
		RawQuery      string
		Body          []byte
		Headers       http.Header
		ContentLength int64
	}

	record := new(recorded)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		if err := r.Body.Close(); err != nil {
			t.Fatalf("failed to close request body: %v", err)
		}

		record.Method = r.Method
		record.Path = r.URL.Path
		record.RawQuery = r.URL.RawQuery
		record.Body = body
		record.Headers = r.Header.Clone()
		record.ContentLength = r.ContentLength

		w.WriteHeader(http.StatusAccepted)
	}))
	t.Cleanup(srv.Close)

	client, err := NewClient(srv.Client(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}

	payload := map[string]any{"message": "hello"}

	req := client.RequestPOST(context.Background(), "/submit").
		QueryParam("page", "1").
		JSONBody(payload)

	resp, err := req.Send()
	if err != nil {
		t.Fatalf("unexpected error sending request: %v", err)
	}
	t.Cleanup(func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	})

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	}

	expectedBody, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	if string(record.Body) != string(expectedBody) {
		t.Fatalf("unexpected body: %q", string(record.Body))
	}

	if record.Headers.Get(ContentType) != ApplicationJSON {
		t.Fatalf("expected content type %q, got %q", ApplicationJSON, record.Headers.Get(ContentType))
	}

	if record.ContentLength != int64(len(expectedBody)) {
		t.Fatalf("expected content length %d, got %d", len(expectedBody), record.ContentLength)
	}

	headerLength := record.Headers.Get(ContentLength)
	if headerLength != "" && headerLength != strconv.Itoa(len(expectedBody)) {
		t.Fatalf("expected content-length header %d, got %s", len(expectedBody), headerLength)
	}

	if record.RawQuery != "page=1" {
		t.Fatalf("unexpected query string: %s", record.RawQuery)
	}

	if req.err != nil {
		t.Fatalf("expected request error to be nil, got %v", req.err)
	}
}

func TestRequestMultipartFileAndFormField(t *testing.T) {
	const (
		fileName    = "test.txt"
		fileContent = "hello from file"
		fieldName   = "description"
		fieldValue  = "testing"
	)

	type recorded struct {
		Method        string
		Path          string
		Headers       http.Header
		Fields        map[string]string
		Files         map[string][]byte
		Filenames     map[string]string
		BodySize      int
		ContentLength int64
	}

	record := new(recorded)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mediaType, params, err := mime.ParseMediaType(r.Header.Get(ContentType))
		if err != nil {
			t.Fatalf("failed to parse content type: %v", err)
		}
		if mediaType != "multipart/form-data" {
			t.Fatalf("expected multipart/form-data, got %s", mediaType)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read multipart body: %v", err)
		}
		if err := r.Body.Close(); err != nil {
			t.Fatalf("failed to close multipart body: %v", err)
		}

		reader := multipart.NewReader(bytes.NewReader(body), params["boundary"])

		fields := make(map[string]string)
		files := make(map[string][]byte)
		filenames := make(map[string]string)

		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("failed to read multipart part: %v", err)
			}

			data, err := io.ReadAll(part)
			if err != nil {
				t.Fatalf("failed to read part data: %v", err)
			}

			if filename := part.FileName(); filename != "" {
				files[part.FormName()] = data
				filenames[part.FormName()] = filename
			} else {
				fields[part.FormName()] = string(data)
			}
		}

		record.Method = r.Method
		record.Path = r.URL.Path
		record.Headers = r.Header.Clone()
		record.Fields = fields
		record.Files = files
		record.Filenames = filenames
		record.BodySize = len(body)
		record.ContentLength = r.ContentLength

		w.WriteHeader(http.StatusCreated)
	}))
	t.Cleanup(srv.Close)

	client, err := NewClient(srv.Client(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}

	req := client.MultipartPOST(context.Background(), "/upload")

	resp, err := req.
		FormField(fieldName, fieldValue).
		File("file", fileName, strings.NewReader(fileContent)).
		Send()
	if err != nil {
		t.Fatalf("unexpected error sending multipart request: %v", err)
	}
	t.Cleanup(func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	})

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	}
	if req.mw != nil {
		t.Fatalf("expected request writer to be cleared after send")
	}
	if req.err != nil {
		t.Fatalf("expected request error to remain nil, got %v", req.err)
	}

	if record.Method != http.MethodPost {
		t.Fatalf("unexpected method: %s", record.Method)
	}
	if record.Path != "/upload" {
		t.Fatalf("unexpected path: %s", record.Path)
	}

	if got := record.Fields[fieldName]; got != fieldValue {
		t.Fatalf("unexpected form field value: %s", got)
	}

	fileData, ok := record.Files["file"]
	if !ok {
		t.Fatalf("expected file field to be present in multipart payload")
	}
	if string(fileData) != fileContent {
		t.Fatalf("unexpected file content: %q", string(fileData))
	}

	if record.Filenames["file"] != fileName {
		t.Fatalf("unexpected filename: %s", record.Filenames["file"])
	}

	if !strings.HasPrefix(record.Headers.Get(ContentType), "multipart/form-data") {
		t.Fatalf("unexpected content-type header: %s", record.Headers.Get(ContentType))
	}

	if record.ContentLength != int64(record.BodySize) {
		t.Fatalf("expected content length %d, got %d", record.BodySize, record.ContentLength)
	}

	expectedLength := strconv.Itoa(record.BodySize)
	headerLength := record.Headers.Get(ContentLength)
	if headerLength != "" && headerLength != expectedLength {
		t.Fatalf("expected content-length header %s, got %s", expectedLength, headerLength)
	}
}

func BenchmarkRequestJSONBody(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
		w.WriteHeader(http.StatusAccepted)
	}))
	b.Cleanup(srv.Close)

	client, err := NewClient(srv.Client(), srv.URL)
	if err != nil {
		b.Fatalf("unexpected error creating client: %v", err)
	}

	payload := map[string]any{"message": "hello"}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp, err := client.RequestPOST(context.Background(), "/submit").
			JSONBody(payload).
			Send()
		if err != nil {
			b.Fatalf("unexpected error sending request: %v", err)
		}
		if resp.Body != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}
	}
}

func BenchmarkRequestMultipart(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reader, err := r.MultipartReader()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err)
			}
			_, _ = io.Copy(io.Discard, part)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	b.Cleanup(srv.Close)

	client, err := NewClient(srv.Client(), srv.URL)
	if err != nil {
		b.Fatalf("unexpected error creating client: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp, err := client.MultipartPOST(context.Background(), "/upload").
			FormField("description", "benchmark").
			File("file", "bench.txt", strings.NewReader("file payload")).
			Send()
		if err != nil {
			b.Fatalf("unexpected error sending multipart request: %v", err)
		}
		if resp.Body != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}
	}
}
