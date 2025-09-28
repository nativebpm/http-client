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
	req    *Request
	writer *multipart.Writer
	buffer *bytes.Buffer
}

func (m *MultipartRequest) File(fieldName, filename string, content io.Reader) *MultipartRequest {
	if m.req.err != nil {
		return m
	}
	part, err := m.writer.CreateFormFile(fieldName, filename)
	if err != nil {
		m.req.err = fmt.Errorf("failed to create form file: %w", err)
		return m
	}
	buf := make([]byte, m.req.bufferSize)
	if _, err := io.CopyBuffer(part, content, buf); err != nil {
		m.req.err = fmt.Errorf("failed to copy file content: %w", err)
	}
	return m
}

func (m *MultipartRequest) Field(name, value string) *MultipartRequest {
	if m.req.err != nil {
		return m
	}
	if err := m.writer.WriteField(name, value); err != nil {
		m.req.err = fmt.Errorf("failed to write form field %q: %w", name, err)
	}
	return m
}

func (m *MultipartRequest) Send() (*http.Response, error) {
	if m.req.err != nil {
		return nil, m.req.err
	}
	if err := m.writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}
	m.req.request.Body = io.NopCloser(m.buffer)
	length := m.buffer.Len()
	m.req.request.ContentLength = int64(length)
	m.req.Header(ContentType, m.writer.FormDataContentType())
	m.req.Header(ContentLength, strconv.Itoa(length))
	return m.req.client.Do(m.req.request)
}
