package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
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

func (c *Client) MultipartMethodPost(ctx context.Context, path string) *Multipart {
	req := &Multipart{
		Client: c,
	}
	req.request, req.err = http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL.JoinPath(path).String(), nil)
	return req
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
