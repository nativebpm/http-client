package request

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"sync"
)

type formData struct {
	dataType   DataType
	file       io.Reader
	key, value string
}

// Multipart provides a streaming multipart/form-data builder for HTTP requests.
// It uses io.Pipe to stream data without buffering the entire payload in memory.
// All methods (except Send) are chainable for fluent API usage.
type Multipart struct {
	client  *http.Client
	request *http.Request
	data    chan formData
	done    chan struct{}
	mw      *multipart.Writer
	pw      *io.PipeWriter
	wg      sync.WaitGroup
}

// NewMultipart creates a new Multipart builder for streaming multipart/form-data requests.
// The worker goroutine starts immediately and waits for data to be added via Param/File methods.
// Call Send to execute the HTTP request with all accumulated headers and form data.
func NewMultipart(ctx context.Context, client *http.Client, method, url string) *Multipart {
	pr, pw := io.Pipe()
	done := make(chan struct{})

	r := &Multipart{
		client: client,
		data:   make(chan formData, 100),
		pw:     pw,
		mw:     multipart.NewWriter(pw),
		done:   done,
	}

	rdr := &readerWrapper{reader: pr, done: done}
	// Error from NewRequestWithContext is ignored as it only fails for invalid method/URL
	r.request, _ = http.NewRequestWithContext(ctx, method, url, rdr)
	r.request.Header.Set("Content-Type", r.mw.FormDataContentType())

	r.wg.Add(1)
	go r.worker()

	return r
}

func (r *Multipart) worker() {
	defer r.wg.Done()
	for form := range r.data {
		var err error
		switch form.dataType {
		case ParamType:
			err = r.mw.WriteField(form.key, form.value)
		case FileType:
			var part io.Writer
			part, err = r.mw.CreateFormFile(form.key, form.value)
			if err == nil {
				_, err = io.Copy(part, form.file)
			}
		}
		if err != nil {
			r.pw.CloseWithError(err)
			return
		}
	}
}

// Param adds a string field to the multipart form.
// Returns the Multipart instance for method chaining.
func (r *Multipart) Param(key, value string) *Multipart {
	r.data <- formData{dataType: ParamType, key: key, value: value}
	return r
}

// Bool adds a boolean field to the multipart form.
// The value is converted to "true" or "false" string.
func (r *Multipart) Bool(key string, value bool) *Multipart {
	return r.Param(key, strconv.FormatBool(value))
}

// Float adds a float64 field to the multipart form.
// The value is formatted using strconv.FormatFloat with 'f' format.
func (r *Multipart) Float(key string, value float64) *Multipart {
	return r.Param(key, strconv.FormatFloat(value, 'f', -1, 64))
}

// Int adds an integer field to the multipart form.
func (r *Multipart) Int(key string, value int) *Multipart {
	return r.Param(key, strconv.Itoa(value))
}

// File adds a file field to the multipart form.
// The content is read from the provided io.Reader and streamed to the request.
func (r *Multipart) File(key, filename string, content io.Reader) *Multipart {
	r.data <- formData{dataType: FileType, key: key, value: filename, file: content}
	return r
}

// Header sets an HTTP header on the request.
// This should be called before Send to ensure headers are included in the request.
func (r *Multipart) Header(key, value string) *Multipart {
	r.request.Header.Set(key, value)
	return r
}

// Send executes the HTTP request with all accumulated form data and headers.
// It waits for the HTTP client to start reading, then closes the form writer
// and waits for the response. Returns the HTTP response or an error.
func (r *Multipart) Send() (*http.Response, error) {
	respCh := make(chan *http.Response, 1)
	errCh := make(chan error, 1)

	go func() {
		resp, err := r.client.Do(r.request)
		if err != nil {
			errCh <- err
			return
		}
		respCh <- resp
	}()

	<-r.done

	close(r.data)
	r.wg.Wait()
	r.mw.Close()
	r.pw.Close()

	select {
	case resp := <-respCh:
		return resp, nil
	case err := <-errCh:
		return nil, err
	}
}
