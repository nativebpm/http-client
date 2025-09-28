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
)

const (
	ApplicationJSON = "application/json"
	ContentType     = "Content-Type"
	ContentLength   = "Content-Length"
)

type Request struct {
	*Client
	request *http.Request
	writer  *multipart.Writer
	buffer  *bytes.Buffer
	err     error
}

type Client struct {
	baseURL    *url.URL
	client     *http.Client
	bufferSize int
}

func NewClient(client *http.Client, baseURL string) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %v", err)
	}

	const bufferSize = 1 << 12 // 4096 bytes

	return &Client{
		client:     client,
		baseURL:    u,
		bufferSize: bufferSize,
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

func (r *Request) multipart() bool {
	return r.writer != nil && r.buffer != nil
}

func (r *Request) Multipart() error {
	if r.err != nil {
		return r.err
	}

	if r.multipart() {
		return nil
	}

	r.buffer = bytes.NewBuffer(make([]byte, 0, r.bufferSize))
	r.writer = multipart.NewWriter(r.buffer)

	return nil
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
		r.Header(key, value)
	}

	return r
}

func (r *Request) ContentType(contentType string) *Request {
	return r.Header(ContentType, contentType)
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
	r.request.ContentLength = int64(len(body))
	r.Header(ContentLength, strconv.Itoa(len(body)))

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

	r.BytesBody(jsonData)
	r.ContentType(ApplicationJSON)

	return r
}

func (r *Request) File(fieldName, filename string, content io.Reader) *Request {
	if r.err != nil {
		return r
	}

	if !r.multipart() {
		if r.err = r.Multipart(); r.err != nil {
			return r
		}
	}

	part, err := r.writer.CreateFormFile(fieldName, filename)
	if err != nil {
		r.err = fmt.Errorf("failed to create form file: %w", err)
		return r
	}

	buf := make([]byte, r.bufferSize)

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

	if !r.multipart() {
		if r.err = r.Multipart(); r.err != nil {
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

func (r *Request) GetRequest() (*Request, error) {
	return r, r.err
}

func (r *Request) Send() (*http.Response, error) {
	if r.err != nil {
		return nil, r.err
	}

	if r.multipart() {
		defer func() {
			r.buffer = nil
			r.writer = nil
		}()

		if err := r.writer.Close(); err != nil {
			return nil, fmt.Errorf("failed to close multipart writer: %w", err)
		}

		r.request.Body = io.NopCloser(r.buffer)
		r.request.ContentLength = int64(r.buffer.Len())
		r.Header(ContentType, r.writer.FormDataContentType())
		r.Header(ContentLength, strconv.Itoa(r.buffer.Len()))
	}

	return r.client.Do(r.request)
}
