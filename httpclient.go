package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

const (
	ApplicationJSON = "application/json"
	ContentType     = "Content-Type"
	ContentLength   = "Content-Length"
)

const (
	bufferSize = 1 << 12
)

var bufferPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 0, bufferSize)
		return &buf
	},
}

type Request struct {
	*Client
	request   *http.Request
	multipart bool
	writer    *multipart.Writer
	buffer    *bytes.Buffer
	err       error
}

type Client struct {
	baseURL *url.URL
	client  *http.Client
}

func NewClient(client *http.Client, baseURL string) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %v", err)
	}

	return &Client{
		client:  client,
		baseURL: u,
	}, nil
}

func (c *Client) MethodGet(ctx context.Context, path string) *Request {
	req := &Request{
		Client: c,
	}
	req.request, req.err = http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL.JoinPath(path).String(), nil)
	return req
}

func (c *Client) MethodPost(ctx context.Context, path string) *Request {
	req := &Request{
		Client: c,
	}
	req.request, req.err = http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL.JoinPath(path).String(), nil)
	return req
}

func (c *Client) MethodPut(ctx context.Context, path string) *Request {
	req := &Request{
		Client: c,
	}
	req.request, req.err = http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL.JoinPath(path).String(), nil)
	return req
}

func (c *Client) MethodPatch(ctx context.Context, path string) *Request {
	req := &Request{
		Client: c,
	}
	req.request, req.err = http.NewRequestWithContext(ctx, http.MethodPatch, c.baseURL.JoinPath(path).String(), nil)
	return req
}

func (c *Client) MethodDelete(ctx context.Context, path string) *Request {
	req := &Request{
		Client: c,
	}
	req.request, req.err = http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL.JoinPath(path).String(), nil)
	return req
}

func (r *Request) Multipart() *Request {
	if r.err != nil {
		return r
	}

	if r.multipart {
		return r
	}

	r.buffer = &bytes.Buffer{}
	r.buffer.Grow(bufferSize)
	r.writer = multipart.NewWriter(r.buffer)
	r.multipart = true

	return r
}

func (r *Request) Header(key, value string) *Request {
	if r.err != nil {
		return r
	}

	if r.request.Header == nil {
		r.request.Header = make(http.Header)
	}
	r.request.Header.Set(key, value)

	return r
}

func (r *Request) Headers(headers map[string]string) *Request {
	if r.err != nil {
		return r
	}

	for key, value := range headers {
		r = r.Header(key, value)
	}

	return r
}

func (r *Request) ContentType(contentType string) *Request {
	return r.Header(ContentType, contentType)
}

func (r *Request) JSONContentType() *Request {
	return r.ContentType(ApplicationJSON)
}

func (r *Request) QueryParam(key, value string) *Request {
	if r.err != nil {
		return r
	}

	q := r.request.URL.Query()
	q.Set(key, value)
	r.request.URL.RawQuery = q.Encode()

	return r
}

func (r *Request) QueryParams(params map[string]string) *Request {
	if r.err != nil {
		return r
	}

	q := r.request.URL.Query()
	for key, value := range params {
		q.Set(key, value)
	}
	r.request.URL.RawQuery = q.Encode()

	return r
}

func (r *Request) QueryValues(values url.Values) *Request {
	if r.err != nil {
		return r
	}

	q := r.request.URL.Query()
	for k := range values {
		v := values.Get(k)
		if v == "" {
			q.Del(k)
			continue
		}
		q.Add(k, values.Get(k))
	}
	r.request.URL.RawQuery = q.Encode()

	return r
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

func (r *Request) BytesBody(body []byte) *Request {
	if r.err != nil {
		return r
	}

	r.request.Body = io.NopCloser(bytes.NewReader(body))
	r = r.Header(ContentLength, strconv.Itoa(len(body)))

	return r
}

func (r *Request) StringBody(body string) *Request {
	return r.BytesBody([]byte(body))
}

func (r *Request) JSONBody(body any) *Request {
	if r.err != nil {
		return r
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		r.err = fmt.Errorf("failed to marshal JSON: %w", err)
		return r
	}

	r = r.BytesBody(jsonData)
	r = r.JSONContentType()

	return r
}

func (r *Request) File(fieldName, filename string, content io.Reader) *Request {
	if r.err != nil {
		return r
	}

	if !r.multipart {
		r = r.Multipart()
		if r.err != nil {
			return r
		}
	}

	part, err := r.writer.CreateFormFile(fieldName, filename)
	if err != nil {
		r.err = fmt.Errorf("failed to create form file: %w", err)
		return r
	}

	var buf []byte
	if p := bufferPool.Get(); p != nil {
		if bufPtr, ok := p.(*[]byte); ok {
			buf = (*bufPtr)[:cap(*bufPtr)]
		} else {
			buf = make([]byte, bufferSize)
		}
	} else {
		buf = make([]byte, bufferSize)
	}
	defer func() {
		if len(buf) > 0 {
			buf = buf[:0]
			bufferPool.Put(&buf)
		}
	}()

	_, err = io.CopyBuffer(part, content, buf)
	if err != nil {
		r.err = fmt.Errorf("failed to copy file content: %w", err)
		return r
	}

	return r
}

func (r *Request) FormField(fieldName, value string) *Request {
	if r.err != nil {
		return r
	}

	if !r.multipart {
		r = r.Multipart()
		if r.err != nil {
			return r
		}
	}

	err := r.writer.WriteField(fieldName, value)
	if err != nil {
		r.err = fmt.Errorf("failed to write form field %q: %w", fieldName, err)
		return r
	}

	return r
}

func (r *Request) Err() error {
	return r.err
}

func (r *Request) Send() (*http.Response, error) {
	if r.err != nil {
		return nil, r.err
	}

	if r.multipart && r.writer != nil && r.buffer != nil {
		if err := r.writer.Close(); err != nil {
			return nil, fmt.Errorf("failed to close multipart writer: %w", err)
		}

		r.request.Body = io.NopCloser(r.buffer)
		r = r.Header(ContentType, r.writer.FormDataContentType())
		r = r.Header(ContentLength, strconv.Itoa(r.buffer.Len()))
	}

	return r.client.Do(r.request)
}
