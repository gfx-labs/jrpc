package jjson

import (
	"bytes"
	"sync"
)

var globalPool *pool

func init() {
	globalPool = NewPool()
}

// New returns a new object of the Buffers Pool
func NewPool() *pool {
	p := pool{
		pool: sync.Pool{
			New: func() any {
				return new(bytes.Buffer)
			},
		},
	}

	return &p
}

type pool struct {
	pool sync.Pool
}

func (p *pool) Get() *bytes.Buffer {
	return p.pool.Get().(*bytes.Buffer)
}
func (p *pool) Put(buf *bytes.Buffer) {
	buf.Reset()
	p.pool.Put(buf)
}

// Get returns a buffer from the pool or creates a new one
func GetBuf() *bytes.Buffer {
	return globalPool.Get()
}

// Put resets and puts back a given buffer to the pool
func PutBuf(buf *bytes.Buffer) {
	globalPool.Put(buf)
}
