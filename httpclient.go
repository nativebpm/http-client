// Package httpclient provides a convenient HTTP client with request builders.
// It supports both regular and multipart requests with deferred operations.
package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/nativebpm/http-client/request"
)

// Client is an HTTP client that wraps http.Client with convenience methods.
// It maintains a base URL and provides request builders for different HTTP methods.
type Client struct {
	baseURL *url.URL
	client  *http.Client
}

// NewClient creates a new HTTP client with the given http.Client and base URL.
// Returns an error if the base URL is invalid.
func NewClient(client *http.Client, baseURL string) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	c := &Client{
		client:  client,
		baseURL: u,
	}
	return c, nil
}

func (c *Client) url(path string) string {
	return c.baseURL.JoinPath(path).String()
}

// MultipartPOST creates a multipart POST request builder for the given path.
func (c *Client) MultipartPOST(ctx context.Context, path string) *request.Multipart {
	return request.NewMultipart(ctx, c.client, http.MethodPost, c.url(path))
}

// MultipartPUT creates a multipart PUT request builder for the given path.
func (c *Client) MultipartPUT(ctx context.Context, path string) *request.Multipart {
	return request.NewMultipart(ctx, c.client, http.MethodPut, c.url(path))
}

// RequestGET creates a GET request builder for the given path.
func (c *Client) RequestGET(ctx context.Context, path string) *request.Request {
	return request.NewRequest(ctx, c.client, http.MethodGet, c.url(path))
}

// RequestPOST creates a POST request builder for the given path.
func (c *Client) RequestPOST(ctx context.Context, path string) *request.Request {
	return request.NewRequest(ctx, c.client, http.MethodPost, c.url(path))
}

// RequestPUT creates a PUT request builder for the given path.
func (c *Client) RequestPUT(ctx context.Context, path string) *request.Request {
	return request.NewRequest(ctx, c.client, http.MethodPut, c.url(path))
}

// RequestPATCH creates a PATCH request builder for the given path.
func (c *Client) RequestPATCH(ctx context.Context, path string) *request.Request {
	return request.NewRequest(ctx, c.client, http.MethodPatch, c.url(path))
}

// RequestDELETE creates a DELETE request builder for the given path.
func (c *Client) RequestDELETE(ctx context.Context, path string) *request.Request {
	return request.NewRequest(ctx, c.client, http.MethodDelete, c.url(path))
}
