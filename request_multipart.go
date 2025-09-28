package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
)

type MultipartRequest struct {
	*Request
	writer *multipart.Writer
	buffer *bytes.Buffer
	err    error
}

func (r *MultipartRequest) GetMultipartRequest() (*MultipartRequest, error) {
	return r, r.err
}

func (r *MultipartRequest) Header(key, value string) *MultipartRequest {
	if r.err != nil {
		return r
	}

	if r.request.Header == nil {
		r.request.Header = make(http.Header)
	}
	r.request.Header.Set(key, value)

	return r
}

func (r *MultipartRequest) File(fieldName, filename string, content io.Reader) *MultipartRequest {
	if r.err != nil {
		return r
	}
	part, err := r.writer.CreateFormFile(fieldName, filename)
	if err != nil {
		r.err = fmt.Errorf("failed to create form file: %w", err)
		return r
	}
	buf := make([]byte, r.Request.bufferSize)
	if _, err := io.CopyBuffer(part, content, buf); err != nil {
		r.err = fmt.Errorf("failed to copy file content: %w", err)
	}
	return r
}

func (r *MultipartRequest) Field(name, value string) *MultipartRequest {
	if r.err != nil {
		return r
	}
	if err := r.writer.WriteField(name, value); err != nil {
		r.err = fmt.Errorf("failed to write form field %q: %w", name, err)
	}
	return r
}

func (r *MultipartRequest) Send() (*http.Response, error) {
	if r.err != nil {
		return nil, r.err
	}
	if err := r.writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}
	r.Request.request.Body = io.NopCloser(r.buffer)
	length := r.buffer.Len()
	r.Request.request.ContentLength = int64(length)
	r.Request.Header(ContentType, r.writer.FormDataContentType())
	r.Request.Header(ContentLength, strconv.Itoa(length))
	return r.Request.client.Do(r.Request.request)
}
