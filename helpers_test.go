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

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := pool.Get()
			buf[0] = 1
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
