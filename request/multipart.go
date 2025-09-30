package request

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
)

type Multipart struct {
	client  *http.Client
	request *http.Request
	err     error
	mw      *multipart.Writer
	pr      *io.PipeReader
	pw      *io.PipeWriter
	ops     []func() error
}

func NewMultipart(ctx context.Context, c *http.Client, method, url string) *Multipart {
	r := &Multipart{client: c}
	r.pr, r.pw = io.Pipe()
	r.mw = multipart.NewWriter(r.pw)
	r.request, r.err = http.NewRequestWithContext(ctx, method, url, r.pr)
	return r
}

func (r *Multipart) Err() error {
	return r.err
}

func (r *Multipart) Send() (*http.Response, error) {
	if r.err != nil {
		return nil, r.err
	}

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

func (r *Multipart) Header(key, value string) *Multipart {
	if r.err != nil {
		return r
	}
	r.request.Header.Set(key, value)
	return r
}

func (r *Multipart) Param(key, value string) *Multipart {
	if r.err != nil {
		return r
	}
	r.ops = append(r.ops, func() error {
		err := r.mw.WriteField(key, value)
		if err != nil {
			return fmt.Errorf("failed to write form field %q: %w", key, err)
		}
		return nil
	})
	return r
}

func (r *Multipart) Bool(fieldName string, value bool) *Multipart {
	return r.Param(fieldName, strconv.FormatBool(value))
}

func (r *Multipart) Float(fieldName string, value float64) *Multipart {
	return r.Param(fieldName, strconv.FormatFloat(value, 'f', -1, 64))
}

func (r *Multipart) File(key, filename string, content io.Reader) *Multipart {
	if r.err != nil {
		return r
	}
	r.ops = append(r.ops, func() error {
		part, err := r.mw.CreateFormFile(key, filename)
		if err != nil {
			return fmt.Errorf("failed to create form file: %w", err)
		}
		_, err = io.Copy(part, content)
		if err != nil {
			return fmt.Errorf("failed to copy file content: %w", err)
		}
		return nil
	})
	return r
}
