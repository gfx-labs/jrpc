package jjson

import (
	"github.com/valyala/bytebufferpool"
)

// GetBuf returns a buffer from the pool or creates a new one
func GetBuf() *bytebufferpool.ByteBuffer {
	return bytebufferpool.Get()
}

// PutBuf puts back a given buffer to the pool
func PutBuf(buf *bytebufferpool.ByteBuffer) {
	bytebufferpool.Put(buf)
}
