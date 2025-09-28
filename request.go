package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
)

type Request struct {
	*Client
	request *http.Request
	err     error
}

func (r *Request) Multipart() *MultipartRequest {
	buf := bytes.NewBuffer(make([]byte, 0, r.bufferSize))
	w := multipart.NewWriter(buf)
	return &MultipartRequest{req: r, writer: w, buffer: buf}
}

func (r *Request) Header(key, value string) *Request {
	if r.err != nil {
		return r
	}

	if r.request.Header == nil {
		r.request.Header = make(http.Header)
	}
	r.request.Header.Set(key, value)

	return r
}

func (r *Request) ContentType(contentType string) *Request {
	return r.Header(ContentType, contentType)
}

func (r *Request) QueryParam(key, value string) *Request {
	if r.err != nil {
		return r
	}

	q := r.request.URL.Query()
	q.Set(key, value)
	r.request.URL.RawQuery = q.Encode()

	return r
}

func (r *Request) QueryParams(params map[string]string) *Request {
	if r.err != nil {
		return r
	}

	q := r.request.URL.Query()
	for key, value := range params {
		q.Set(key, value)
	}
	r.request.URL.RawQuery = q.Encode()

	return r
}

func (r *Request) QueryValues(values url.Values) *Request {
	if r.err != nil {
		return r
	}

	q := r.request.URL.Query()
	for k := range values {
		v := values.Get(k)
		if v == "" {
			q.Del(k)
			continue
		}
		q.Add(k, values.Get(k))
	}
	r.request.URL.RawQuery = q.Encode()

	return r
}

func (r *Request) Body(body io.ReadCloser) *Request {
	if r.err != nil {
		return r
	}

	if r.request.Body != nil {
		_ = r.request.Body.Close()
	}

	r.request.Body = body
	return r
}

func (r *Request) BytesBody(body []byte) *Request {
	if r.err != nil {
		return r
	}

	r.request.Body = io.NopCloser(bytes.NewReader(body))
	r.request.ContentLength = int64(len(body))
	r.Header(ContentLength, strconv.Itoa(len(body)))

	return r
}

func (r *Request) StringBody(body string) *Request {
	return r.BytesBody([]byte(body))
}

func (r *Request) JSONBody(body any) *Request {
	if r.err != nil {
		return r
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		r.err = fmt.Errorf("failed to marshal JSON: %w", err)
		return r
	}

	r.BytesBody(jsonData)
	r.ContentType(ApplicationJSON)

	return r
}
func (r *Request) Send() (*http.Response, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.client.Do(r.request)
}
