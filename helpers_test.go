package httpclient

import "testing"

func TestBytePoolGetReturnsSizedBuffer(t *testing.T) {
	const size = 32

	pool := newBytePool(size)

	buf := pool.Get()
	if len(buf) != size {
		t.Fatalf("expected buffer length %d, got %d", size, len(buf))
	}
	if cap(buf) < size {
		t.Fatalf("expected buffer capacity >= %d, got %d", size, cap(buf))
	}

	pool.Put(buf)

	reused := pool.Get()
	if len(reused) != size {
		t.Fatalf("expected reused buffer length %d, got %d", size, len(reused))
	}
	if cap(reused) < size {
		t.Fatalf("expected reused buffer capacity >= %d, got %d", size, cap(reused))
	}
}

// Test channel-based pool specific behavior
func TestBytePoolChannelBehavior(t *testing.T) {
	const size = 64
	pool := newBytePool(size)

	// Pool should be pre-populated with 64 buffers
	var buffers [][]byte
	for i := 0; i < 64; i++ {
		buf := pool.Get()
		if len(buf) != size {
			t.Fatalf("buffer %d: expected length %d, got %d", i, size, len(buf))
		}
		buffers = append(buffers, buf)
	}

	// 65th Get should create new buffer (not from channel)
	buf65 := pool.Get()
	if len(buf65) != size {
		t.Fatalf("overflow buffer: expected length %d, got %d", size, len(buf65))
	}

	// Put all buffers back
	for _, buf := range buffers {
		pool.Put(buf)
	}
	pool.Put(buf65) // This should be discarded as channel is full

	// Verify we can get 64 buffers again
	for i := 0; i < 64; i++ {
		buf := pool.Get()
		if len(buf) != size {
			t.Fatalf("reused buffer %d: expected length %d, got %d", i, size, len(buf))
		}
	}
}

// Test Put with invalid buffers
func TestBytePoolPutValidation(t *testing.T) {
	const size = 128
	pool := newBytePool(size)

	// Put nil buffer - should be ignored
	pool.Put(nil)

	// Put buffer with insufficient capacity - should be ignored
	smallBuf := make([]byte, size/2)
	pool.Put(smallBuf)

	// Valid put should work
	validBuf := make([]byte, size)
	pool.Put(validBuf)

	// Get should return valid buffer
	retrieved := pool.Get()
	if len(retrieved) != size {
		t.Fatalf("expected length %d, got %d", size, len(retrieved))
	}
}

func TestBufferPoolResetsBuffers(t *testing.T) {
	const size = 64

	pool := newBufferPool(size)

	buf := pool.Get()
	if buf.Len() != 0 {
		t.Fatalf("expected fresh buffer to be empty, got length %d", buf.Len())
	}

	_, err := buf.WriteString("hello world")
	if err != nil {
		t.Fatalf("failed to write to buffer: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatalf("expected buffer to contain data after write")
	}

	pool.Put(buf)

	reused := pool.Get()
	if reused.Len() != 0 {
		t.Fatalf("expected reused buffer to be reset, got length %d", reused.Len())
	}
}

func BenchmarkBytePool(b *testing.B) {
	const size = 1 << 12

	pool := newBytePool(size)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := pool.Get()
			buf[0] = 1
			pool.Put(buf)
		}
	})
}

// Benchmark specifically for Put operation to verify zero allocations
func BenchmarkBytePoolPutOnly(b *testing.B) {
	const size = 1024
	pool := newBytePool(size)
	buf := make([]byte, size)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pool.Put(buf)
	}
}

// Benchmark for Get operation
func BenchmarkBytePoolGetOnly(b *testing.B) {
	const size = 1024
	pool := newBytePool(size)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = pool.Get()
	}
}

// Benchmark comparing with make([]byte) directly
func BenchmarkMakeByteSlice(b *testing.B) {
	const size = 1024

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = make([]byte, size)
	}
}

// Benchmark realistic usage pattern with pre-warmed pool
func BenchmarkBytePoolRealistic(b *testing.B) {
	const size = 1024
	pool := newBytePool(size)

	// Pre-warm: get and put some buffers to populate the channel
	for i := 0; i < 10; i++ {
		buf := pool.Get()
		pool.Put(buf)
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := pool.Get()
			// Simulate some work
			buf[0] = 1
			buf[size-1] = 255
			pool.Put(buf)
		}
	})
}

func BenchmarkBufferPool(b *testing.B) {
	const size = 1 << 12

	pool := newBufferPool(size)
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := pool.Get()
			if _, err := buf.WriteString("benchmark"); err != nil {
				b.Fatalf("write failed: %v", err)
			}
			pool.Put(buf)
		}
	})
}
