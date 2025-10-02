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
// Uses io.Pipe to stream data without buffering in memory.
// Methods are chainable for fluent API usage.
type Multipart struct {
	client  *http.Client
	request *http.Request
	data    chan formData
}

// NewMultipart creates a new streaming multipart/form-data request builder.
func NewMultipart(ctx context.Context, client *http.Client, method, url string) *Multipart {
	r := &Multipart{
		client: client,
		data:   make(chan formData, 100),
	}

	r.request, _ = http.NewRequestWithContext(ctx, method, url, nil)

	return r
}

// Param adds a string field to the multipart form.
func (r *Multipart) Param(key, value string) *Multipart {
	r.data <- formData{dataType: ParamType, key: key, value: value}
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
	r.data <- formData{dataType: FileType, key: key, value: filename, file: content}
	return r
}

// Header sets an HTTP header on the request.
func (r *Multipart) Header(key, value string) *Multipart {
	r.request.Header.Set(key, value)
	return r
}

// Send executes the HTTP request and returns the response.
// Starts a worker goroutine to write multipart data from the buffered channel.
func (r *Multipart) Send() (*http.Response, error) {
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)

	r.request.Body = pr
	r.request.Header.Set("Content-Type", mw.FormDataContentType())

	errCh := make(chan error, 1)

	// Worker goroutine: reads from channel and writes to multipart writer
	go func() {
		defer pw.Close()
		defer mw.Close()

		close(r.data)

		for form := range r.data {
			var err error
			switch form.dataType {
			case ParamType:
				err = mw.WriteField(form.key, form.value)
			case FileType:
				var part io.Writer
				part, err = mw.CreateFormFile(form.key, form.value)
				if err == nil {
					_, err = io.Copy(part, form.file)
				}
			}
			if err != nil {
				pw.CloseWithError(err)
				errCh <- err
				return
			}
		}
	}()

	resp, err := r.client.Do(r.request)
	if err != nil {
		return nil, err
	}

	select {
	case err := <-errCh:
		resp.Body.Close()
		return nil, err
	default:
		return resp, nil
	}
}
