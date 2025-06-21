package jjson

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/bytebufferpool"
)

func TestGlobalPool(t *testing.T) {
	t.Run("GetBuf and PutBuf", func(t *testing.T) {
		// Get a buffer from global pool
		buf := GetBuf()
		assert.NotNil(t, buf)

		// Use it
		buf.WriteString("global pool test")
		assert.Equal(t, "global pool test", buf.String())

		// Store the data before putting back
		data := buf.String()
		assert.Equal(t, "global pool test", data)

		// Put it back - don't access buf after this!
		PutBuf(buf)
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

				// Verify data before putting back
				if buf.String() != data {
					errors <- assert.AnError
					return
				}

				// Put back - don't access buf after this!
				PutBuf(buf)
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
		buf := &bytebufferpool.ByteBuffer{}
		buf.WriteString("benchmark data")
		// buf goes out of scope and is garbage collected
	}
}
