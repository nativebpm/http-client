# http-client

Fluent HTTP client for Go with streaming support and zero dependencies.

## Features

- Fluent API with method chaining
- Path variables with `{placeholder}` syntax
- Streaming for large payloads via `io.Pipe`
- Context-aware with proper cancellation handling
- Multipart/form-data file uploads
- Type-safe methods for parameters
- Built-in middleware for retry, rate limiting, and logging
- Zero third-party dependencies

## Installation

```bash
go get github.com/nativebpm/http-client
```

## Quick Start

```go
client, _ := httpclient.NewClient(&http.Client{}, "https://api.example.com")

// Simple GET
resp, err := client.GET(ctx, "/users").Send()

// POST with JSON
resp, err = client.POST(ctx, "/users").
    JSON(map[string]string{"name": "John"}).
    Send()

// Path variables
resp, err = client.GET(ctx, "/users/{id}").
    PathInt("id", 123).
    Send()

// Multipart upload
resp, err = client.Multipart(ctx, "/upload").
    File("document", "file.pdf", fileReader).
    Send()
```

## API

### Client Methods

```go
GET(ctx, path string) *Request
POST(ctx, path string) *Request
PUT(ctx, path string) *Request
PATCH(ctx, path string) *Request
DELETE(ctx, path string) *Request
Request(ctx, method, path string) *Request
Multipart(ctx, path string) *Multipart
```

### Request Builder

```go
Header(key, value string)           // Set header
PathParam(key, value string)        // Replace {key} in path
PathInt(key string, value int)      // Path param as int
PathBool(key string, value bool)    // Path param as bool
PathFloat(key string, value float64) // Path param as float
Param(key, value string)            // Query parameter
Int(key string, value int)          // Query param as int
Bool(key string, value bool)        // Query param as bool
Float(key string, value float64)    // Query param as float
JSON(data any)                      // JSON body
Body(body io.ReadCloser, contentType string) // Custom body
Timeout(duration time.Duration)     // Request timeout
Send() (*http.Response, error)      // Execute request
```

### Multipart Builder

Same as Request, plus:
```go
File(key, filename string, content io.Reader) // Add file
```

## Middleware

The client supports middleware for intercepting and modifying requests/responses. Middleware are functions that wrap `http.RoundTripper`.

### Fluent Builder Methods

For convenience, the client provides fluent methods for common middleware:

```go
client, _ := httpclient.NewClient(&http.Client{}, "https://api.example.com")

// Add retry with default config
client.WithRetry()

// Add retry with custom config
client.WithRetryConfig(httpclient.RetryConfig{
    MaxRetries: 5,
    Backoff: func(attempt int) time.Duration {
        return time.Duration(attempt*2) * time.Second
    },
})

// Add rate limiting (10 requests per minute)
client.WithRateLimit(10, time.Minute/10)

// Add logging
client.WithLogging()

// Chain them
client.WithRetry().WithRateLimit(5, time.Second).WithLogging()
```

### Custom Middleware

```go
client.Use(func(next http.RoundTripper) http.RoundTripper {
    return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
        // Pre-request logic
        resp, err := next.RoundTrip(req)
        // Post-response logic
        return resp, err
    })
})
```

### Built-in Middleware Functions

You can also use the middleware functions directly:

```go
import "github.com/nativebpm/http-client"

// Retry
client.Use(httpclient.RetryMiddleware(httpclient.DefaultRetryConfig()))

// Rate limiting
limiter := httpclient.NewSimpleRateLimiter(10, time.Minute/10)
client.Use(httpclient.RateLimitMiddleware(limiter))

// Logging
client.Use(httpclient.LoggingMiddleware())
```

See `examples/` for complete working examples.

## Examples

See the `examples/` directory for complete working examples, including middleware usage.

### Path Variables

```go
// Single parameter
client.GET(ctx, "/users/{id}").PathParam("id", "123").Send()
// → GET /users/123

// Multiple parameters
client.POST(ctx, "/api/{version}/users/{id}").
    PathParam("version", "v1").
    PathInt("id", 123).
    Send()
// → POST /api/v1/users/123

// Mixed with query params
client.GET(ctx, "/users/{id}/posts").
    PathInt("id", 123).
    Param("page", "2").
    Send()
// → GET /users/123/posts?page=2
```

### File Uploads

```go
// Single file
client.Multipart(ctx, "/upload").
    File("document", "report.pdf", fileReader).
    Send()

// Multiple files with params
client.Multipart(ctx, "/users/{id}/documents").
    PathInt("id", 123).
    Param("category", "invoices").
    File("file1", "doc1.pdf", reader1).
    File("file2", "doc2.pdf", reader2).
    Send()
```

### Timeouts

```go
// Method timeout
client.POST(ctx, "/slow").
    Timeout(30 * time.Second).
    JSON(data).
    Send()

// Context timeout
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()
client.POST(ctx, "/slow").JSON(data).Send()
```

### Use Cases

- [Gotenberg Client](https://github.com/nativebpm/gotenberg-client) - A Go client for Gotenberg PDF generation service using this [http-client](https://github.com/nativebpm/http-client).

## Testing

```bash
go test ./...
```

## License

MIT