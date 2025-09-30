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

type Request struct {
	client  *http.Client
	request *http.Request
	err     error
}

func NewRequest(ctx context.Context, c *http.Client, method, url string) *Request {
	r := &Request{client: c}
	r.request, r.err = http.NewRequestWithContext(ctx, method, url, nil)
	return r
}

func (r *Request) Err() error {
	return r.err
}

func (r *Request) Send() (*http.Response, error) {
	if r.err != nil {
		return nil, r.err
	}

	return r.client.Do(r.request)
}

func (r *Request) Header(key, value string) *Request {
	if r.err != nil {
		return r
	}
	r.request.Header.Set(key, value)
	return r
}

func (r *Request) Param(key, value string) *Request {
	if r.err != nil {
		return r
	}

	q := r.request.URL.Query()
	q.Set(key, value)
	r.request.URL.RawQuery = q.Encode()

	return r
}

func (r *Request) Bool(fieldName string, value bool) *Request {
	return r.Param(fieldName, strconv.FormatBool(value))
}

func (r *Request) Float(fieldName string, value float64) *Request {
	return r.Param(fieldName, strconv.FormatFloat(value, 'f', -1, 64))
}

func (r *Request) Body(body io.ReadCloser) *Request {
	if r.err != nil {
		return r
	}

	if r.request.Body != nil {
		_ = r.request.Body.Close()
	}

	r.request.Body = body
	return r
}

func (r *Request) Bytes(body []byte) *Request {
	if r.err != nil {
		return r
	}

	r.request.Body = io.NopCloser(bytes.NewReader(body))
	r.request.ContentLength = int64(len(body))
	r.Header(ContentLength, strconv.Itoa(len(body)))

	return r
}

func (r *Request) JSON(body any) *Request {
	if r.err != nil {
		return r
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		r.err = fmt.Errorf("failed to marshal JSON: %w", err)
		return r
	}

	r.Bytes(jsonData)
	r.Header(ContentType, ApplicationJSON)

	return r
}
