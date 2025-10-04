// Package httpclient provides a convenient HTTP client with request builders.
package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/nativebpm/http-client/internal/formdata"
	"github.com/nativebpm/http-client/internal/request"
)

// Client wraps http.Client and provides request builders for different HTTP methods.
type Client struct {
	baseURL *url.URL
	client  *http.Client
}

// NewClient creates a new HTTP client with the given base URL.
// Returns an error if the base URL is invalid.
func NewClient(client *http.Client, baseURL string) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	return &Client{
		client:  client,
		baseURL: u,
	}, nil
}

func (c *Client) url(path string) string {
	return c.baseURL.JoinPath(path).String()
}

// Multipart creates a multipart/form-data POST request builder.
func (c *Client) Multipart(ctx context.Context, path string) *formdata.Multipart {
	return formdata.NewMultipart(ctx, c.client, http.MethodPost, c.url(path))
}

// MultipartWithMethod creates a multipart/form-data request builder with HTTP method.
func (c *Client) MultipartWithMethod(ctx context.Context, path, method string) *formdata.Multipart {
	return formdata.NewMultipart(ctx, c.client, method, c.url(path))
}

// Request creates a standard HTTP request builder.
func (c *Client) Request(ctx context.Context, method, path string) *request.Request {
	return request.NewRequest(ctx, c.client, method, c.url(path))
}

// GET creates a GET request builder.
func (c *Client) GET(ctx context.Context, path string) *request.Request {
	return c.Request(ctx, http.MethodGet, path)
}

// POST creates a POST request builder.
func (c *Client) POST(ctx context.Context, path string) *request.Request {
	return c.Request(ctx, http.MethodPost, path)
}

// PUT creates a PUT request builder.
func (c *Client) PUT(ctx context.Context, path string) *request.Request {
	return c.Request(ctx, http.MethodPut, path)
}

// PATCH creates a PATCH request builder.
func (c *Client) PATCH(ctx context.Context, path string) *request.Request {
	return c.Request(ctx, http.MethodPatch, path)
}

// DELETE creates a DELETE request builder.
func (c *Client) DELETE(ctx context.Context, path string) *request.Request {
	return c.Request(ctx, http.MethodDelete, path)
}
