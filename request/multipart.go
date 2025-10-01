package request

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
)

// Multipart represents a multipart/form-data request builder.
// It uses streaming with io.Pipe to avoid buffering the entire request in memory.
type Multipart struct {
	client  *http.Client
	request *http.Request
	mw      *multipart.Writer
	pr      *io.PipeReader
	pw      *io.PipeWriter
	ops     []func() error
}

// NewMultipart creates a new multipart/form-data request builder.
// If the request creation fails, the error will be returned when Send is called.
func NewMultipart(ctx context.Context, c *http.Client, method, url string) *Multipart {
	r := &Multipart{client: c}
	r.pr, r.pw = io.Pipe()
	r.mw = multipart.NewWriter(r.pw)
	request, err := http.NewRequestWithContext(ctx, method, url, r.pr)
	if err != nil {
		r.ops = append(r.ops, func() error {
			return err
		})
		return r
	}
	r.request = request
	return r
}

// Send executes all deferred operations in a goroutine and sends the multipart request.
// The multipart data is streamed using io.Pipe to avoid memory buffering.
func (r *Multipart) Send() (*http.Response, error) {
	r.Header(ContentType, r.mw.FormDataContentType())

	go func() {
		defer r.pw.Close()
		defer r.mw.Close()
		for _, op := range r.ops {
			if err := op(); err != nil {
				r.pw.CloseWithError(err)
				return
			}
		}
	}()

	return r.client.Do(r.request)
}

// Header adds a header to the request. This is applied immediately, not deferred.
func (r *Multipart) Header(key, value string) *Multipart {
	r.request.Header.Set(key, value)
	return r
}

// Param adds a form field to the multipart request.
func (r *Multipart) Param(key, value string) *Multipart {
	r.ops = append(r.ops, func() error {
		if err := r.mw.WriteField(key, value); err != nil {
			return fmt.Errorf("failed to write form field %q: %w", key, err)
		}
		return nil
	})
	return r
}

// Bool adds a boolean form field to the multipart request.
func (r *Multipart) Bool(fieldName string, value bool) *Multipart {
	return r.Param(fieldName, strconv.FormatBool(value))
}

// Float adds a float64 form field to the multipart request.
func (r *Multipart) Float(fieldName string, value float64) *Multipart {
	return r.Param(fieldName, strconv.FormatFloat(value, 'f', -1, 64))
}

// File adds a file field to the multipart request.
// The content will be read when Send is called.
func (r *Multipart) File(key, filename string, content io.Reader) *Multipart {
	r.ops = append(r.ops, func() error {
		part, err := r.mw.CreateFormFile(key, filename)
		if err != nil {
			return fmt.Errorf("failed to create form file: %w", err)
		}
		if _, err := io.Copy(part, content); err != nil {
			return fmt.Errorf("failed to copy file content: %w", err)
		}
		return nil
	})
	return r
}
