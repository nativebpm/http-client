package request

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type dataType int

const (
	NoneType dataType = iota
	jsonType
)

type reqData struct {
	dataType dataType
	body     any
}

// Request provides a builder for standard HTTP requests.
type Request struct {
	client     *http.Client
	request    *http.Request
	data       reqData
	cancelFunc context.CancelFunc
}

// NewRequest creates a new HTTP request builder.
func NewRequest(ctx context.Context, client *http.Client, method, url string) *Request {
	request, _ := http.NewRequestWithContext(ctx, method, url, nil)
	return &Request{
		client:  client,
		request: request,
	}
}

// Timeout sets a timeout for the request.
func (r *Request) Timeout(duration time.Duration) *Request {
	ctx, cancel := context.WithTimeout(r.request.Context(), duration)
	r.cancelFunc = cancel
	r.request = r.request.WithContext(ctx)
	return r
}

// Send executes the HTTP request and returns the response.
func (r *Request) Send() (*http.Response, error) {
	if r.cancelFunc != nil {
		defer r.cancelFunc()
	}

	if r.data.body != nil {
		switch r.data.dataType {
		case jsonType:
			pr, pw := io.Pipe()
			r.request.Body = pr
			ctx := r.request.Context()

			go func() {
				defer pw.Close()

				select {
				case <-ctx.Done():
					pw.CloseWithError(ctx.Err())
					return
				default:
				}

				encoder := json.NewEncoder(pw)
				if err := encoder.Encode(r.data.body); err != nil {
					pw.CloseWithError(err)
					return
				}
			}()
		}
	}

	return r.client.Do(r.request)
}

// Header sets an HTTP header on the request.
func (r *Request) Header(key, value string) *Request {
	r.request.Header.Set(key, value)
	return r
}

// PathParam replaces a path variable placeholder in the URL.
// Replaces {key} with the provided value.
// Example: "/users/{id}" with PathParam("id", "123") becomes "/users/123"
func (r *Request) PathParam(key, value string) *Request {
	placeholder := "{" + key + "}"
	r.request.URL.Path = strings.ReplaceAll(r.request.URL.Path, placeholder, value)
	return r
}

// PathInt replaces a path variable placeholder with an integer value.
func (r *Request) PathInt(key string, value int) *Request {
	return r.PathParam(key, strconv.Itoa(value))
}

// PathBool replaces a path variable placeholder with a boolean value.
func (r *Request) PathBool(key string, value bool) *Request {
	return r.PathParam(key, strconv.FormatBool(value))
}

// PathFloat replaces a path variable placeholder with a float64 value.
func (r *Request) PathFloat(key string, value float64) *Request {
	return r.PathParam(key, strconv.FormatFloat(value, 'f', -1, 64))
}

// Param adds a query parameter to the request.
func (r *Request) Param(key, value string) *Request {
	q := r.request.URL.Query()
	q.Set(key, value)
	r.request.URL.RawQuery = q.Encode()
	return r
}

// Bool adds a boolean query parameter to the request.
func (r *Request) Bool(key string, value bool) *Request {
	return r.Param(key, strconv.FormatBool(value))
}

// Float adds a float64 query parameter to the request.
func (r *Request) Float(key string, value float64) *Request {
	return r.Param(key, strconv.FormatFloat(value, 'f', -1, 64))
}

// Int adds an integer query parameter to the request.
func (r *Request) Int(key string, value int) *Request {
	return r.Param(key, strconv.Itoa(value))
}

// Body sets the request body and Content-Type header.
func (r *Request) Body(body io.ReadCloser, contentType string) *Request {
	r.request.Body = body
	r.request.Header.Set("Content-Type", contentType)
	return r
}

// JSON sets the request body as JSON.
func (r *Request) JSON(body any) *Request {
	r.data = reqData{dataType: jsonType, body: body}
	r.request.Header.Set("Content-Type", "application/json")
	return r
}
