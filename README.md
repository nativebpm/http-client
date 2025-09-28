# HTTP Client

Universal HTTP client with method chaining for building requests.

## Features

- Method chaining
- Multipart/form-data support  
- Header management
- Thread-safe operations
- Buffer pooling

## Usage

```go
client, err := httpclient.NewClient(&http.Client{}, "https://api.example.com")
if err != nil {
    panic(err)
}

resp, err := client.MethodPost(context.Background(), "/endpoint").
    Header("Authorization", "Bearer token").
    JSONBody(map[string]string{"key": "value"}).
    Send()
```

## API

**Client:**
- `NewClient(httpClient *http.Client, baseURL string) (*Client, error)`
- `MethodGet/Post/Put/Patch/Delete(ctx context.Context, path string) *Request`

**Request:**
- `Header(key, value string) *Request`
- `QueryParam(key, value string) *Request`  
- `JSONBody(body any) *Request`
- `File(fieldName, filename string, content io.Reader) *Request`
- `FormField(fieldName, value string) *Request`
- `Send() (*http.Response, error)`