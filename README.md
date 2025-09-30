
# http-client


Universal HTTP client for Go with method chaining, multipart support, and zero third-party dependencies.

## Features
  
- Method chaining for building requests
- Multipart/form-data file uploads
- Header and query param management
- Minimal API, no external dependencies

## Usage Example

```go
package main

import (
    "context"
    "log"
    "net/http"
    "github.com/nativebpm/http-client"
)

func main() {
    client, err := httpclient.NewClient(&http.Client{}, "https://api.example.com")
    if err != nil {
        log.Fatal(err)
    }

    resp, err := client.RequestPOST(context.Background(), "/endpoint").
        Header("Authorization", "Bearer token").
        JSON(map[string]string{"key": "value"}).
        Send()
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()
    // ... handle resp ...
}
```

### Multipart Example

```go
resp, err := client.MultipartPOST(ctx, "/upload").
    Param("desc", "test").
    File("file", "file.txt", strings.NewReader("hello")).
    Send()
```

## API

**Client:**
- `NewClient(httpClient *http.Client, baseURL string) (*Client, error)`
- `RequestGET(ctx, path) *Request`
- `RequestPOST(ctx, path) *Request`
- `RequestPUT(ctx, path) *Request`
- `RequestPATCH(ctx, path) *Request`
- `RequestDELETE(ctx, path) *Request`
- `MultipartPOST(ctx, path) *Multipart`
- `MultipartPUT(ctx, path) *Multipart`

**Request:**
- `Header(key, value string) *Request`
- `Param(key, value string) *Request` (adds query param)
- `Bool(fieldName string, value bool) *Request` (adds query param)
- `Float(fieldName string, value float64) *Request` (adds query param)
- `Body(body io.ReadCloser) *Request`
- `Bytes(body []byte) *Request`
- `JSON(body any) *Request`
- `Send() (*http.Response, error)`

**Multipart:**
- `Header(key, value string) *Multipart`
- `Param(key, value string) *Multipart` (adds form field)
- `Bool(fieldName string, value bool) *Multipart` (adds form field)
- `Float(fieldName string, value float64) *Multipart` (adds form field)
- `File(fieldName, filename string, content io.Reader) *Multipart`
- `Send() (*http.Response, error)`

## Testing

Run all tests and benchmarks:

```sh
go test -v -bench=. ./...
```

## Project Structure

- `httpclient.go` — main client and request logic
- `request/request.go` — request builder, JSON, query, headers
- `request/multipart.go` — multipart/form-data support
- `httpclient_test.go`, `request/request_test.go`, `request/multipart_test.go` — tests and benchmarks

## License

MIT — see [LICENSE](LICENSE)