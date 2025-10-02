package request

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

// Request represents an HTTP request builder with minimal allocations.
// Uses direct header/param manipulation without intermediate slices.
type Request struct {
	client  *http.Client
	request *http.Request
}

// NewRequest creates a new HTTP request builder.
// Error from NewRequestWithContext is ignored as it only fails for invalid method/URL.
func NewRequest(ctx context.Context, client *http.Client, method, url string) *Request {
	req, _ := http.NewRequestWithContext(ctx, method, url, nil)
	return &Request{
		client:  client,
		request: req,
	}
}

// Send executes the HTTP request.
// Returns the HTTP response or an error.
func (r *Request) Send() (*http.Response, error) {
	return r.client.Do(r.request)
}

// Header sets an HTTP header on the request.
// Returns the Request instance for method chaining.
func (r *Request) Header(key, value string) *Request {
	r.request.Header.Set(key, value)
	return r
}

// Param adds a query parameter to the request.
// Returns the Request instance for method chaining.
func (r *Request) Param(key, value string) *Request {
	q := r.request.URL.Query()
	q.Set(key, value)
	r.request.URL.RawQuery = q.Encode()
	return r
}

// Bool adds a boolean query parameter to the request.
// The value is converted to "true" or "false" string.
func (r *Request) Bool(key string, value bool) *Request {
	return r.Param(key, strconv.FormatBool(value))
}

// Float adds a float64 query parameter to the request.
// The value is formatted using strconv.FormatFloat with 'f' format.
func (r *Request) Float(key string, value float64) *Request {
	return r.Param(key, strconv.FormatFloat(value, 'f', -1, 64))
}

// Int adds an integer query parameter to the request.
func (r *Request) Int(key string, value int) *Request {
	return r.Param(key, strconv.Itoa(value))
}

// Body sets the request body from an io.ReadCloser.
// Caller is responsible for closing the body if needed.
func (r *Request) Body(body io.ReadCloser, contentType string) *Request {
	r.request.Body = body
	r.request.Header.Set(ContentType, contentType)
	return r
}

// JSON sets the request body as JSON using json.Encoder for streaming.
// This avoids buffering the entire JSON in memory via json.Marshal.
func (r *Request) JSON(data any) *Request {
	pr, pw := io.Pipe()

	r.request.Body = pr
	r.request.Header.Set(ContentType, ApplicationJSON)

	go func() {
		encoder := json.NewEncoder(pw)
		if err := encoder.Encode(data); err != nil {
			pw.CloseWithError(err)
			return
		}
		pw.Close()
	}()

	return r
}
