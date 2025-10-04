// Package httpclient provides a convenient HTTP client with request builders.
package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/nativebpm/http-client/internal/formdata"
	"github.com/nativebpm/http-client/internal/request"
)

// Public types re-exports
type Multipart = formdata.Multipart
type Request = request.Request

// Client wraps http.Client and provides request builders for different HTTP methods.
type Client struct {
	baseURL     *url.URL
	client      *http.Client
	middlewares []func(http.RoundTripper) http.RoundTripper
	transport   http.RoundTripper // cached transport with applied middlewares
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

// Use adds a middleware to the client. Middlewares are applied in the order they are added.
// Each middleware is a function that takes a RoundTripper and returns a wrapped RoundTripper.
func (c *Client) Use(middleware func(http.RoundTripper) http.RoundTripper) *Client {
	c.middlewares = append(c.middlewares, middleware)
	c.transport = nil // invalidate cache
	return c
}

// SetTransport sets a custom base RoundTripper. This will be wrapped by middlewares.
func (c *Client) SetTransport(rt http.RoundTripper) *Client {
	c.client.Transport = rt
	c.transport = nil // invalidate cache
	return c
}

// getTransport returns the effective RoundTripper with middlewares applied.
// It caches the result to avoid recomputation.
func (c *Client) getTransport() http.RoundTripper {
	if c.transport != nil {
		return c.transport
	}
	base := c.client.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	for _, m := range c.middlewares {
		base = m(base)
	}
	c.transport = base
	return base
}

func (c *Client) url(path string) string {
	return c.baseURL.JoinPath(path).String()
}

// Request creates a standard HTTP request builder.
func (c *Client) Request(ctx context.Context, method, path string) *request.Request {
	return request.NewRequest(ctx, c.client, method, c.url(path), c.getTransport)
}

// MultipartRequest creates a multipart/form-data request builder with HTTP method.
func (c *Client) MultipartRequest(ctx context.Context, path, method string) *formdata.Multipart {
	return formdata.NewMultipart(ctx, c.client, method, c.url(path), c.getTransport)
}

// Multipart creates a multipart/form-data POST request builder.
func (c *Client) Multipart(ctx context.Context, path string) *formdata.Multipart {
	return formdata.NewMultipart(ctx, c.client, http.MethodPost, c.url(path), c.getTransport)
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

// WithRetry adds retry middleware with default config
func (c *Client) WithRetry() *Client {
	return c.Use(RetryMiddleware(DefaultRetryConfig()))
}

// WithRetryConfig adds retry middleware with custom config
func (c *Client) WithRetryConfig(config RetryConfig) *Client {
	return c.Use(RetryMiddleware(config))
}

// WithRateLimit adds rate limiting middleware
func (c *Client) WithRateLimit(capacity int, refillRate time.Duration) *Client {
	limiter := NewSimpleRateLimiter(capacity, refillRate)
	return c.Use(RateLimitMiddleware(limiter))
}

// WithLogging adds logging middleware
func (c *Client) WithLogging() *Client {
	return c.Use(LoggingMiddleware())
}
