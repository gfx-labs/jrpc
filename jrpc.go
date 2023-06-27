package jrpc

import (
	"context"
	"net/http"
)

// ServerCodec implements reading, parsing and writing RPC messages for the server side of
// a RPC session. Implementations must be go-routine safe since the codec can be called in
// multiple go-routines concurrently.
type ServerCodec interface {
	PeerInfo() PeerInfo
	ReadBatch() (msgs []*jsonrpcMessage, isBatch bool, err error)
	Close() error

	JsonWriter
}

// jsonWriter can write JSON messages to its underlying connection.
// Implementations must be safe for concurrent use.
type JsonWriter interface {
	WriteJSON(context.Context, any) error
	// Closed returns a channel which is closed when the connection is closed.
	Closed() <-chan any
	// RemoteAddr returns the peer address of the connection.
	RemoteAddr() string
}

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
