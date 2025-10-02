package request

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
)

type formData struct {
	dataType   DataType
	file       io.Reader
	key, value string
}

// Multipart provides a streaming multipart/form-data builder for HTTP requests.
// Form data is collected in a slice and streamed via io.Pipe when Send() is called.
// Methods are chainable for fluent API usage.
type Multipart struct {
	client  *http.Client
	request *http.Request
	data    []formData
}

// NewMultipart creates a new streaming multipart/form-data request builder.
func NewMultipart(ctx context.Context, client *http.Client, method, url string) *Multipart {
	r := &Multipart{
		client: client,
		data:   make([]formData, 0, 16),
	}

	r.request, _ = http.NewRequestWithContext(ctx, method, url, nil)

	return r
}

// Send executes the HTTP request and returns the response.
// Starts a goroutine that streams collected form data via io.Pipe.
// The goroutine respects context cancellation to prevent leaks.
func (r *Multipart) Send() (*http.Response, error) {
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)

	r.request.Body = pr
	r.request.Header.Set("Content-Type", mw.FormDataContentType())

	ctx := r.request.Context()

	go func() {
		defer pw.Close()
		defer mw.Close()

		for _, form := range r.data {
			select {
			case <-ctx.Done():
				pw.CloseWithError(ctx.Err())
				return
			default:
			}

			switch form.dataType {
			case ParamType:
				if err := mw.WriteField(form.key, form.value); err != nil {
					pw.CloseWithError(err)
					return
				}
			case FileType:
				part, err := mw.CreateFormFile(form.key, form.value)
				if err != nil {
					pw.CloseWithError(err)
					return
				}
				if _, err := io.Copy(part, form.file); err != nil {
					pw.CloseWithError(err)
					return
				}

			}
		}
	}()

	return r.client.Do(r.request)
}

// Header sets an HTTP header on the request.
func (r *Multipart) Header(key, value string) *Multipart {
	r.request.Header.Set(key, value)
	return r
}

// Param adds a string field to the multipart form.
func (r *Multipart) Param(key, value string) *Multipart {
	r.data = append(r.data, formData{dataType: ParamType, key: key, value: value})
	return r
}

// Bool adds a boolean field to the multipart form.
func (r *Multipart) Bool(key string, value bool) *Multipart {
	return r.Param(key, strconv.FormatBool(value))
}

// Float adds a float64 field to the multipart form.
func (r *Multipart) Float(key string, value float64) *Multipart {
	return r.Param(key, strconv.FormatFloat(value, 'f', -1, 64))
}

// Int adds an integer field to the multipart form.
func (r *Multipart) Int(key string, value int) *Multipart {
	return r.Param(key, strconv.Itoa(value))
}

// File adds a file field to the multipart form and streams the content.
func (r *Multipart) File(key, filename string, content io.Reader) *Multipart {
	r.data = append(r.data, formData{dataType: FileType, key: key, value: filename, file: content})
	return r
}
