package httpclient

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type Multipart struct {
	Client  *Client
	request *http.Request
	mw      *multipart.Writer
	pr      *io.PipeReader
	pw      *io.PipeWriter
	err     error
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
	go func() {
		defer r.pw.Close()
		defer r.mw.Close()
		part, err := r.mw.CreateFormFile(fieldName, filename)
		if err != nil {
			r.pw.CloseWithError(fmt.Errorf("failed to create form file: %w", err))
			return
		}
		_, err = io.Copy(part, content)
		if err != nil {
			r.pw.CloseWithError(fmt.Errorf("failed to copy file content: %w", err))
			return
		}
	}()
	return r
}

func (r *Multipart) FormField(fieldName, value string) *Multipart {
	if r.err != nil {
		return r
	}
	go func() {
		err := r.mw.WriteField(fieldName, value)
		if err != nil {
			r.pw.CloseWithError(fmt.Errorf("failed to write form field %q: %w", fieldName, err))
			return
		}
	}()
	return r
}

func (r *Multipart) GetRequest() (*Multipart, error) {
	return r, r.err
}

func (r *Multipart) Send() (*http.Response, error) {
	if r.err != nil {
		return nil, r.err
	}
	r.Header(ContentType, r.mw.FormDataContentType())
	resp, err := r.Client.client.Do(r.request)
	r.mw = nil
	r.pr = nil
	r.pw = nil
	return resp, err
}
