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

// Request represents an HTTP request builder with minimal allocations.
type Request struct {
	client  *http.Client
	request *http.Request
	headers []ItemOp
	params  []ItemOp
	body    BodyOp
}

// NewRequest creates a new HTTP request builder.
func NewRequest(ctx context.Context, c *http.Client, method, url string) *Request {
	return NewRequestWithOpsCapacity(ctx, c, method, url, defaultOpsCapacity)
}

func NewRequestWithOpsCapacity(ctx context.Context, c *http.Client, method, url string, opsCapacity int) *Request {
	if opsCapacity < defaultOpsCapacity {
		opsCapacity = defaultOpsCapacity
	}

	r := &Request{
		client:  c,
		headers: make([]ItemOp, 0, opsCapacity/4),
		params:  make([]ItemOp, 0, opsCapacity/2),
	}

	request, _ := http.NewRequestWithContext(ctx, method, url, nil)
	r.request = request
	return r
}

// Send executes all operations and sends the HTTP request.
func (r *Request) Send() (*http.Response, error) {
	for _, h := range r.headers {
		r.request.Header.Set(h.Key, h.Value)
	}

	if len(r.params) > 0 {
		q := r.request.URL.Query()
		for _, p := range r.params {
			q.Set(p.Key, p.Value)
		}
		r.request.URL.RawQuery = q.Encode()
	}

	if err := r.setBody(); err != nil {
		return nil, err
	}

	return r.client.Do(r.request)
}

func (r *Request) setBody() error {
	switch r.body.Type {
	case BodyTypeBytes:
		{
			r.request.Body = io.NopCloser(bytes.NewReader(r.body.Bytes))

			contentLength := len(r.body.Bytes)
			r.request.ContentLength = int64(contentLength)
			r.request.Header.Set(ContentLength, strconv.Itoa(contentLength))
			r.request.Header.Set(ContentType, r.body.ContentType)
		}
	case BodyTypeReader:
		{
			if r.request.Body != nil {
				_ = r.request.Body.Close()
			}
			r.request.Body = r.body.Reader

			contentLength := int(r.body.ContentLength)
			r.request.ContentLength = int64(contentLength)
			r.request.Header.Set(ContentLength, strconv.Itoa(contentLength))
			r.request.Header.Set(ContentType, r.body.ContentType)
		}
	case BodyTypeJSON:
		{
			jsonData, err := json.Marshal(r.body.JSON)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			r.request.Body = io.NopCloser(bytes.NewReader(jsonData))

			contentLength := len(jsonData)
			r.request.ContentLength = int64(contentLength)
			r.request.Header.Set(ContentLength, strconv.Itoa(contentLength))
			r.request.Header.Set(ContentType, ApplicationJSON)
		}
	}

	return nil
}

func (r *Request) Header(key, value string) *Request {
	r.headers = append(r.headers, ItemOp{Key: key, Value: value})
	return r
}

func (r *Request) Param(key, value string) *Request {
	r.params = append(r.params, ItemOp{Key: key, Value: value})
	return r
}

func (r *Request) Bool(fieldName string, value bool) *Request {
	return r.Param(fieldName, strconv.FormatBool(value))
}

func (r *Request) Float(fieldName string, value float64) *Request {
	return r.Param(fieldName, strconv.FormatFloat(value, 'f', -1, 64))
}

func (r *Request) Body(body io.ReadCloser, contentType string, contentLength int64) *Request {
	r.body = BodyOp{Type: BodyTypeReader, Reader: body, ContentType: contentType, ContentLength: contentLength}
	return r
}

func (r *Request) Bytes(body []byte, contentType string) *Request {
	r.body = BodyOp{Type: BodyTypeBytes, Bytes: body, ContentType: contentType, ContentLength: int64(len(body))}
	return r
}

func (r *Request) JSON(body any) *Request {
	r.body = BodyOp{Type: BodyTypeJSON, JSON: body}
	return r
}

func (r *Request) HeaderCount() int {
	return len(r.headers)
}

func (r *Request) ParamCount() int {
	return len(r.params)
}

func (r *Request) HasBody() bool {
	return r.body.Type != BodyTypeNone
}
