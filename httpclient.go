package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/nativebpm/http-client/request"
)

type Client struct {
	baseURL *url.URL
	client  *http.Client
}

func NewClient(client *http.Client, baseURL string) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %v", err)
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

func (c *Client) MultipartPOST(ctx context.Context, path string) *request.Multipart {
	return request.NewMultipart(ctx, c.client, http.MethodPost, c.url(path))
}

func (c *Client) MultipartPUT(ctx context.Context, path string) *request.Multipart {
	return request.NewMultipart(ctx, c.client, http.MethodPut, c.url(path))
}

func (c *Client) RequestGET(ctx context.Context, path string) *request.Request {
	return request.NewRequest(ctx, c.client, http.MethodGet, c.url(path))
}

func (c *Client) RequestPOST(ctx context.Context, path string) *request.Request {
	return request.NewRequest(ctx, c.client, http.MethodPost, c.url(path))
}

func (c *Client) RequestPUT(ctx context.Context, path string) *request.Request {
	return request.NewRequest(ctx, c.client, http.MethodPut, c.url(path))
}

func (c *Client) RequestPATCH(ctx context.Context, path string) *request.Request {
	return request.NewRequest(ctx, c.client, http.MethodPatch, c.url(path))
}

func (c *Client) RequestDELETE(ctx context.Context, path string) *request.Request {
	return request.NewRequest(ctx, c.client, http.MethodDelete, c.url(path))
}
