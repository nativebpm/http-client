package request

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// Request represents an HTTP request builder that uses deferred operations.
// All configuration methods are applied only when Send is called.
type Request struct {
	client  *http.Client
	request *http.Request
	ops     []func() error
}

// NewRequest creates a new HTTP request builder.
// If the request creation fails, the error will be returned when Send is called.
func NewRequest(ctx context.Context, c *http.Client, method, url string) *Request {
	return NewRequestWithOpsCapacity(ctx, c, method, url, defaultOpsCapacity)
}

func NewRequestWithOpsCapacity(ctx context.Context, c *http.Client, method, url string, opsCapacity int) *Request {
	r := &Request{client: c, ops: make([]func() error, 0, opsCapacity)}
	request, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		r.ops = append(r.ops, func() error {
			return err
		})
		return r
	}
	r.request = request
	return r
}

// Send executes all deferred operations and sends the HTTP request.
// Returns the HTTP response or the first error encountered.
func (r *Request) Send() (*http.Response, error) {
	for _, op := range r.ops {
		if err := op(); err != nil {
			return nil, err
		}
	}

	return r.client.Do(r.request)
}

// Header adds a header to the request. The header is set when Send is called.
func (r *Request) Header(key, value string) *Request {
	r.ops = append(r.ops, func() error {
		r.request.Header.Set(key, value)
		return nil
	})
	return r
}

// Param adds a URL query parameter to the request.
func (r *Request) Param(key, value string) *Request {
	r.ops = append(r.ops, func() error {
		q := r.request.URL.Query()
		q.Set(key, value)
		r.request.URL.RawQuery = q.Encode()
		return nil
	})
	return r
}

// Bool adds a boolean URL query parameter to the request.
func (r *Request) Bool(fieldName string, value bool) *Request {
	return r.Param(fieldName, strconv.FormatBool(value))
}

// Float adds a float64 URL query parameter to the request.
func (r *Request) Float(fieldName string, value float64) *Request {
	return r.Param(fieldName, strconv.FormatFloat(value, 'f', -1, 64))
}

// Body sets the request body. Any existing body will be closed when Send is called.
func (r *Request) Body(body io.ReadCloser) *Request {
	r.ops = append(r.ops, func() error {
		if r.request.Body != nil {
			_ = r.request.Body.Close()
		}
		r.request.Body = body
		return nil
	})
	return r
}

// Bytes sets the request body from a byte slice and sets the Content-Length header.
func (r *Request) Bytes(body []byte) *Request {
	r.ops = append(r.ops, func() error {
		r.request.Body = io.NopCloser(bytes.NewReader(body))
		r.request.ContentLength = int64(len(body))
		r.request.Header.Set(ContentLength, strconv.Itoa(len(body)))
		return nil
	})
	return r
}

// JSON marshals the provided value to JSON and sets it as the request body.
// Also sets the Content-Type header to application/json.
func (r *Request) JSON(body any) *Request {
	r.ops = append(r.ops, func() error {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		r.request.Body = io.NopCloser(bytes.NewReader(jsonData))
		r.request.ContentLength = int64(len(jsonData))
		r.request.Header.Set(ContentLength, strconv.Itoa(len(jsonData)))
		r.request.Header.Set(ContentType, ApplicationJSON)
		return nil
	})
	return r
}
