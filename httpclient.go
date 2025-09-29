package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"sync"
)

const (
	ApplicationJSON = "application/json"
	ContentType     = "Content-Type"
	ContentLength   = "Content-Length"
)

type Client struct {
	baseURL    *url.URL
	client     *http.Client
	bufferSize int
	bufferPool sync.Pool
}

func NewClient(client *http.Client, baseURL string) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %v", err)
	}

	const bufferSize = 1 << 12 // 4096 bytes

	c := &Client{
		client:     client,
		baseURL:    u,
		bufferSize: bufferSize,
	}
	c.bufferPool = sync.Pool{
		New: func() any {
			return bytes.NewBuffer(make([]byte, 0, bufferSize))
		},
	}
	return c, nil
}

func (c *Client) requestURL(path string) string {
	return c.baseURL.JoinPath(path).String()
}

func (c *Client) NewRequest(ctx context.Context, method, path string) *Request {
	req := &Request{Client: c}
	req.request, req.err = http.NewRequestWithContext(ctx, method, c.requestURL(path), nil)
	return req
}

func (c *Client) NewMultipartRequest(ctx context.Context, method, path string) *Multipart {
	req := &Multipart{Client: c}
	buf := c.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	req.buffer = buf
	req.writer = multipart.NewWriter(req.buffer)
	req.request, req.err = http.NewRequestWithContext(ctx, method, c.requestURL(path), nil)
	return req
}

func (c *Client) MultipartPOST(ctx context.Context, path string) *Multipart {
	return c.NewMultipartRequest(ctx, http.MethodPost, path)
}

func (c *Client) RequestGET(ctx context.Context, path string) *Request {
	return c.NewRequest(ctx, http.MethodGet, path)
}

func (c *Client) RequestPOST(ctx context.Context, path string) *Request {
	return c.NewRequest(ctx, http.MethodPost, path)
}

func (c *Client) RequestPUT(ctx context.Context, path string) *Request {
	return c.NewRequest(ctx, http.MethodPut, path)
}

func (c *Client) RequestPATCH(ctx context.Context, path string) *Request {
	return c.NewRequest(ctx, http.MethodPatch, path)
}

func (c *Client) RequestDELETE(ctx context.Context, path string) *Request {
	return c.NewRequest(ctx, http.MethodDelete, path)
}
