
# http-client

Fluent HTTP client for Go with streaming support and zero dependencies.

## Features

- Fluent API with method chaining
- Path variables with `{placeholder}` syntax
- Streaming for large payloads via `io.Pipe`
- Context-aware with proper cancellation handling
- Multipart/form-data file uploads
- Type-safe methods for parameters
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

## Examples

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

## Testing

```bash
go test ./...
```

## License

MIT