package httpclient

import (
	"bytes"
	"sync"
)

type bytePool struct {
	p    sync.Pool
	size int
}

func newBytePool(bufferSize int) *bytePool {
	return &bytePool{
		p: sync.Pool{
			New: func() any {
				b := make([]byte, bufferSize)
				return &b
			},
		},
		size: bufferSize,
	}
}
func (tp *bytePool) Get() []byte {
	if v := tp.p.Get(); v != nil {
		if bp, ok := v.(*[]byte); ok && bp != nil {
			b := *bp
			if cap(b) < tp.size {
				return make([]byte, tp.size)
			}
			return b[:tp.size]
		}
	}
	return make([]byte, tp.size)
}
func (tp *bytePool) Put(b []byte) {
	if b == nil {
		return
	}
	if cap(b) < tp.size {
		return
	}
	s := b[:tp.size]
	tp.p.Put(&s)
}

type bufferPool struct {
	p    sync.Pool
	size int
}

func newBufferPool(bufferSize int) *bufferPool {
	return &bufferPool{
		p: sync.Pool{
			New: func() any {
				return bytes.NewBuffer(make([]byte, 0, bufferSize))
			},
		},
		size: bufferSize,
	}
}

func (tp *bufferPool) Get() *bytes.Buffer {
	if v := tp.p.Get(); v != nil {
		if b, ok := v.(*bytes.Buffer); ok {
			b.Reset()
			return b
		}
	}
	return bytes.NewBuffer(make([]byte, 0, tp.size))
}

func (tp *bufferPool) Put(b *bytes.Buffer) {
	if b == nil {
		return
	}
	b.Reset()
	tp.p.Put(b)
}
