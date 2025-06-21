package jjson

import (
	"bytes"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPool(t *testing.T) {
	t.Run("Get and Put", func(t *testing.T) {
		p := NewPool()
		
		// Get a buffer
		buf1 := p.Get()
		assert.NotNil(t, buf1)
		
		// Write some data
		buf1.WriteString("test data")
		assert.Equal(t, "test data", buf1.String())
		
		// Put it back
		p.Put(buf1)
		
		// Buffer should be reset when put back
		assert.Equal(t, 0, buf1.Len())
		
		// Get another buffer (should reuse the same one)
		buf2 := p.Get()
		assert.NotNil(t, buf2)
		assert.Equal(t, 0, buf2.Len())
	})

	t.Run("Multiple buffers", func(t *testing.T) {
		p := NewPool()
		
		// Get multiple buffers
		bufs := make([]*bytes.Buffer, 5)
		for i := range bufs {
			bufs[i] = p.Get()
			bufs[i].WriteString("data" + string(rune(i)))
		}
		
		// Put them all back
		for _, buf := range bufs {
			p.Put(buf)
		}
		
		// Get them again - they should be reset
		for i := 0; i < 5; i++ {
			buf := p.Get()
			assert.Equal(t, 0, buf.Len())
			p.Put(buf)
		}
	})
}

func TestGlobalPool(t *testing.T) {
	t.Run("GetBuf and PutBuf", func(t *testing.T) {
		// Get a buffer from global pool
		buf := GetBuf()
		assert.NotNil(t, buf)
		
		// Use it
		buf.WriteString("global pool test")
		assert.Equal(t, "global pool test", buf.String())
		
		// Put it back
		PutBuf(buf)
		assert.Equal(t, 0, buf.Len())
	})

	t.Run("Concurrent access", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, 100)
		
		// Run 100 goroutines that get and put buffers
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				
				buf := GetBuf()
				if buf == nil {
					errors <- assert.AnError
					return
				}
				
				// Write some data
				data := "goroutine " + strconv.Itoa(n)
				buf.WriteString(data)
				
				// Verify data
				if buf.String() != data {
					errors <- assert.AnError
					return
				}
				
				// Put back
				PutBuf(buf)
				
				// Verify reset
				if buf.Len() != 0 {
					errors <- assert.AnError
					return
				}
			}(i)
		}
		
		wg.Wait()
		close(errors)
		
		// Check for errors
		errorCount := 0
		for err := range errors {
			if err != nil {
				errorCount++
			}
		}
		assert.Equal(t, 0, errorCount, "Concurrent access produced errors")
	})
}

func BenchmarkPool(b *testing.B) {
	p := NewPool()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := p.Get()
		buf.WriteString("benchmark data")
		p.Put(buf)
	}
}

func BenchmarkGlobalPool(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := GetBuf()
		buf.WriteString("benchmark data")
		PutBuf(buf)
	}
}

func BenchmarkNoPool(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		buf.WriteString("benchmark data")
		// buf goes out of scope and is garbage collected
	}
}

func BenchmarkPoolConcurrent(b *testing.B) {
	p := NewPool()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := p.Get()
			buf.WriteString("concurrent benchmark data")
			p.Put(buf)
		}
	})
}

func TestPoolMemoryReuse(t *testing.T) {
	p := NewPool()
	
	// Get a buffer and note its capacity
	buf1 := p.Get()
	buf1.WriteString("this is a test string that should grow the buffer")
	initialCap := buf1.Cap()
	
	// Put it back
	p.Put(buf1)
	
	// Get another buffer
	buf2 := p.Get()
	
	// The capacity should be preserved (buffer reused, not reallocated)
	assert.GreaterOrEqual(t, buf2.Cap(), initialCap, "Buffer capacity should be preserved when reused")
}

