package jrpc

import "context"

type Conn interface {
	Do(ctx context.Context, result any, method string, params any) error
	BatchCall(ctx context.Context, b ...BatchElem) error
	SetHeader(key, value string)
	Close() error
}

type StreamingConn interface {
	Conn

	Notify(ctx context.Context, method string, args ...any) error
}

// BatchElem is an element in a batch request.
type BatchElem struct {
	Method string
	Args   any
	// The result is unmarshaled into this field. Result must be set to a
	// non-nil pointer value of the desired type, otherwise the response will be
	// discarded.
	Result any
	// Error is set if the server returns an error for this request, or if
	// unmarshaling into Result fails. It is not set for I/O errors.
	Error error
}
