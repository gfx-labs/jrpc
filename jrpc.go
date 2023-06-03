package jrpc

import (
	"context"
	"net/http"
)

// http.handler, but for jrpc
type Handler interface {
	ServeRPC(w ResponseWriter, r *Request)
}

// http.HandlerFunc,but for jrpc
type HandlerFunc func(w ResponseWriter, r *Request)

func (fn HandlerFunc) ServeRPC(w ResponseWriter, r *Request) {
	(fn)(w, r)
}

// http.ResponseWriter interface, but for jrpc
type ResponseWriter interface {
	Send(v any, err error) error
	Option(k string, v any)
	Header() http.Header

	Notify(v any) error
}

type Conn interface {
	Do(ctx context.Context, result any, method string, params any) error
	Notify(ctx context.Context, method string, params any) error
	BatchCall(ctx context.Context, b ...*BatchElem) error
	Close() error
}

type StreamingConn interface {
	Conn
	Handler
}

// BatchElem is an element in a batch request.
type BatchElem struct {
	Method string
	Params any

	IsNotification bool

	// The result is unmarshaled into this field. Result must be set to a
	// non-nil pointer value of the desired type, otherwise the response will be
	// discarded.
	Result any
	// Error is set if the server returns an error for this request, or if
	// unmarshaling into Result fails. It is not set for I/O errors.
	Error error
}
