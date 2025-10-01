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
type Multipart struct {
	client  *http.Client
	request *http.Request
	mw      *multipart.Writer
	pr      *io.PipeReader
	pw      *io.PipeWriter
	headers []ItemOp
	params  []ItemOp
	files   []FileOp
}

// NewMultipart creates a new multipart/form-data request builder.
func NewMultipart(ctx context.Context, c *http.Client, method, url string) *Multipart {
	return NewMultipartWithOpsCapacity(ctx, c, method, url, defaultOpsCapacity)
}

func NewMultipartWithOpsCapacity(ctx context.Context, c *http.Client, method, url string, opsCapacity int) *Multipart {
	if opsCapacity < defaultOpsCapacity {
		opsCapacity = defaultOpsCapacity
	}

	r := &Multipart{
		client:  c,
		headers: make([]ItemOp, 0, opsCapacity/4),
		params:  make([]ItemOp, 0, opsCapacity/2),
		files:   make([]FileOp, 0, opsCapacity/4),
	}
	r.pr, r.pw = io.Pipe()
	r.mw = multipart.NewWriter(r.pw)
	request, _ := http.NewRequestWithContext(ctx, method, url, r.pr)
	r.request = request
	return r
}

// Send executes all operations and sends the multipart request.
func (r *Multipart) Send() (*http.Response, error) {
	for _, h := range r.headers {
		r.request.Header.Set(h.Key, h.Value)
	}

	r.request.Header.Set(ContentType, r.mw.FormDataContentType())

	go func() {
		defer r.pw.Close()
		defer r.mw.Close()

		// Write all params
		for _, param := range r.params {
			if err := r.mw.WriteField(param.Key, param.Value); err != nil {
				r.pw.CloseWithError(fmt.Errorf("failed to write form field %q: %w", param.Key, err))
				return
			}
		}

		// Write all files
		for _, file := range r.files {
			part, err := r.mw.CreateFormFile(file.Key, file.Filename)
			if err != nil {
				r.pw.CloseWithError(fmt.Errorf("failed to create form file: %w", err))
				return
			}
			if _, err := io.Copy(part, file.Content); err != nil {
				r.pw.CloseWithError(fmt.Errorf("failed to copy file content: %w", err))
				return
			}
		}
	}()

	return r.client.Do(r.request)
}

func (r *Multipart) Header(key, value string) *Multipart {
	r.headers = append(r.headers, ItemOp{Key: key, Value: value})
	return r
}

func (r *Multipart) Param(key, value string) *Multipart {
	r.params = append(r.params, ItemOp{Key: key, Value: value})
	return r
}

func (r *Multipart) Bool(fieldName string, value bool) *Multipart {
	return r.Param(fieldName, strconv.FormatBool(value))
}

func (r *Multipart) Float(fieldName string, value float64) *Multipart {
	return r.Param(fieldName, strconv.FormatFloat(value, 'f', -1, 64))
}

func (r *Multipart) File(key, filename string, content io.Reader) *Multipart {
	r.files = append(r.files, FileOp{Key: key, Filename: filename, Content: content})
	return r
}

func (r *Multipart) ParamCount() int {
	return len(r.params)
}

func (r *Multipart) FileCount() int {
	return len(r.files)
}

func (r *Multipart) TotalOps() int {
	return len(r.params) + len(r.files)
}
