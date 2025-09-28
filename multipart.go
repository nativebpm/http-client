package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
)

type Multipart struct {
	*Client
	request *http.Request
	writer  *multipart.Writer
	buffer  *bytes.Buffer
	err     error
}

func (r *Multipart) multipart() bool {
	return r.writer != nil && r.buffer != nil
}

func (r *Multipart) Multipart() error {
	if r.err != nil {
		return r.err
	}

	if r.multipart() {
		return nil
	}

	r.buffer = bytes.NewBuffer(make([]byte, 0, r.bufferSize))
	r.writer = multipart.NewWriter(r.buffer)

	return nil
}

func (r *Multipart) Header(key, value string) *Multipart {
	if r.err != nil {
		return r
	}

	if r.request.Header == nil {
		r.request.Header = make(http.Header)
	}
	r.request.Header.Set(key, value)

	return r
}

func (r *Multipart) Headers(headers map[string]string) *Multipart {
	if r.err != nil {
		return r
	}

	for key, value := range headers {
		r.Header(key, value)
	}

	return r
}

func (r *Multipart) ContentType(contentType string) *Multipart {
	return r.Header(ContentType, contentType)
}

func (r *Multipart) File(fieldName, filename string, content io.Reader) *Multipart {
	if r.err != nil {
		return r
	}

	if !r.multipart() {
		if r.err = r.Multipart(); r.err != nil {
			return r
		}
	}

	part, err := r.writer.CreateFormFile(fieldName, filename)
	if err != nil {
		r.err = fmt.Errorf("failed to create form file: %w", err)
		return r
	}

	buf := make([]byte, r.bufferSize)

	_, err = io.CopyBuffer(part, content, buf)
	if err != nil {
		r.err = fmt.Errorf("failed to copy file content: %w", err)
		return r
	}

	return r
}

func (r *Multipart) FormField(fieldName, value string) *Multipart {
	if r.err != nil {
		return r
	}

	if !r.multipart() {
		if r.err = r.Multipart(); r.err != nil {
			return r
		}
	}

	err := r.writer.WriteField(fieldName, value)
	if err != nil {
		r.err = fmt.Errorf("failed to write form field %q: %w", fieldName, err)
		return r
	}

	return r
}

func (r *Multipart) GetRequest() (*Multipart, error) {
	return r, r.err
}

func (r *Multipart) Send() (*http.Response, error) {
	if r.err != nil {
		return nil, r.err
	}

	if r.multipart() {
		defer func() {
			r.buffer = nil
			r.writer = nil
		}()

		if err := r.writer.Close(); err != nil {
			return nil, fmt.Errorf("failed to close multipart writer: %w", err)
		}

		r.request.Body = io.NopCloser(r.buffer)
		r.request.ContentLength = int64(r.buffer.Len())
		r.Header(ContentType, r.writer.FormDataContentType())
		r.Header(ContentLength, strconv.Itoa(r.buffer.Len()))
	}

	return r.client.Do(r.request)
}
