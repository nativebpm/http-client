package formdata

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type DataType int

const (
	NoneType DataType = iota
	ParamType
	FileType
)

type formData struct {
	dataType   DataType
	key, value string
	file       io.Reader
}

// Multipart provides a streaming multipart/form-data builder for HTTP requests.
type Multipart struct {
	client     *http.Client
	request    *http.Request
	data       []formData
	cancelFunc context.CancelFunc
}

// NewMultipart creates a new streaming multipart/form-data request builder.
func NewMultipart(ctx context.Context, client *http.Client, method, url string) *Multipart {
	request, _ := http.NewRequestWithContext(ctx, method, url, nil)
	return &Multipart{
		client:  client,
		request: request,
		data:    make([]formData, 0, 16),
	}
}

// Timeout sets a timeout for the request.
func (r *Multipart) Timeout(duration time.Duration) *Multipart {
	ctx, cancel := context.WithTimeout(r.request.Context(), duration)
	r.cancelFunc = cancel
	r.request = r.request.WithContext(ctx)
	return r
}

// Send executes the HTTP request and returns the response.
func (r *Multipart) Send() (*http.Response, error) {
	if r.cancelFunc != nil {
		defer r.cancelFunc()
	}

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

// PathParam replaces a path variable placeholder in the URL.
// Replaces {key} with the provided value.
// Example: "/users/{id}" with PathParam("id", "123") becomes "/users/123"
func (r *Multipart) PathParam(key, value string) *Multipart {
	placeholder := "{" + key + "}"
	r.request.URL.Path = strings.ReplaceAll(r.request.URL.Path, placeholder, value)
	return r
}

// PathInt replaces a path variable placeholder with an integer value.
func (r *Multipart) PathInt(key string, value int) *Multipart {
	return r.PathParam(key, strconv.Itoa(value))
}

// PathBool replaces a path variable placeholder with a boolean value.
func (r *Multipart) PathBool(key string, value bool) *Multipart {
	return r.PathParam(key, strconv.FormatBool(value))
}

// PathFloat replaces a path variable placeholder with a float64 value.
func (r *Multipart) PathFloat(key string, value float64) *Multipart {
	return r.PathParam(key, strconv.FormatFloat(value, 'f', -1, 64))
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

// File adds a file field to the multipart form.
func (r *Multipart) File(key, filename string, content io.Reader) *Multipart {
	r.data = append(r.data, formData{dataType: FileType, key: key, value: filename, file: content})
	return r
}
