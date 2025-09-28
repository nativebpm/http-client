package httpclient

import (
	"bytes"
	"sync"
)

type bytePool struct {
	ch   chan []byte
	size int
}

func newBytePool(bufferSize int) *bytePool {
	const poolSize = 64
	ch := make(chan []byte, poolSize)
	for i := 0; i < poolSize; i++ {
		ch <- make([]byte, bufferSize)
	}

	return &bytePool{
		ch:   ch,
		size: bufferSize,
	}
}
func (tp *bytePool) Get() []byte {
	select {
	case buf := <-tp.ch:
		return buf[:tp.size]
	default:
		return make([]byte, tp.size)
	}
}
func (tp *bytePool) Put(b []byte) {
	if b == nil || cap(b) < tp.size {
		return
	}

	select {
	case tp.ch <- b:
	default:
	}
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
