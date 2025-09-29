
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
        JSONBody(map[string]string{"key": "value"}).
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
    FormField("desc", "test").
    File("file", "file.txt", strings.NewReader("hello")).
    Send()
```

## API

**Client:**
- `NewClient(httpClient *http.Client, baseURL string) (*Client, error)`
- `RequestGET/POST/PUT/PATCH/DELETE(ctx, path) *Request`
- `MultipartPOST(ctx, path) *Multipart`

**Request:**
- `Header(key, value string) *Request`
- `QueryParam(key, value string) *Request`
- `JSONBody(body any) *Request`
- `Send() (*http.Response, error)`

**Multipart:**
- `FormField(fieldName, value string) *Multipart`
- `File(fieldName, filename string, content io.Reader) *Multipart`
- `Header(key, value string) *Multipart`
- `Send() (*http.Response, error)`

## Testing

Run all tests and benchmarks:

```sh
go test -v -bench=. ./...
```

## Project Structure

- `httpclient.go` — main client and request logic
- `request.go` — request builder, JSON, query, headers
- `multipart.go` — multipart/form-data support
- `httpclient_test.go` — tests and benchmarks

## License

MIT — see [LICENSE](LICENSE)