package request

import (
	"io"
	"sync"
)

var _ io.Reader = (*readerWrapper)(nil)

type readerWrapper struct {
	reader io.Reader
	once   sync.Once
	done   chan struct{}
}

func (r *readerWrapper) Read(p []byte) (n int, err error) {
	r.once.Do(func() {
		close(r.done)
	})
	return r.reader.Read(p)
}
