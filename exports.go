package jrpc

import (
	"context"

	"gfx.cafe/open/jrpc/contrib/codecs"
	"gfx.cafe/open/jrpc/pkg/codec"
	"gfx.cafe/open/jrpc/pkg/server"
)

// to make the repo cleaner, we export everything here. this way the packages dont need to ever import this

// Conn is used to make requests to jsonrpc2 servers
type Conn interface {
	codec.Conn
}

// Handler is the equivalent of http.Handler, but for jsonrpc.
type Handler interface {
	codec.Handler
}

// StreamingConn is a conn that supports streaming methods
type StreamingConn interface {
	codec.StreamingConn
}

// ResponseWriter is used to write responses to the request
type ResponseWriter = codec.ResponseWriter

type (
	// HandlerFunc is a Handler that exists as a function
	HandlerFunc = codec.HandlerFunc
	// Request is the request object
	Request = codec.Request
	// Server is a jrpc server
	Server = server.Server
	// Middleware is a middleware
	Middleware = func(Handler) Handler
	// BatchElem is an element of a batch request
	BatchElem = codec.BatchElem
)

var (

	// NewServer creates a jrpc server
	NewServer = server.NewServer

	// DialContext is to dial a conn with context
	DialContext = codecs.DialContext
	// Dial is to dial a conn with context.Background()
	Dial = codecs.Dial

	// ContextWithConn will attach a conn to the context
	ContextWithConn = codec.ContextWithConn
	// ContextWithPeerInfo will attach a peerinfo to the context
	ContextWithPeerInfo = server.ContextWithPeerInfo

	// ConnFromContext will retrieve a conn from context
	ConnFromContext = codec.ConnFromContext
	// PeerInfoFromContext will retrieve a peerinfo from context
	PeerInfoFromContext = server.PeerInfoFromContext
)

// Do will use the conn to perform a jsonrpc2 call.
func Do[T any](ctx context.Context, c Conn, method string, args any) (*T, error) {
	return codec.Do[T](ctx, c, method, args)
}

// Call will use the conn to perform a jsonrpc2 call, except the args will be encoded as an array of arguments.
func Call[T any](ctx context.Context, c Conn, method string, args ...any) (*T, error) {
	return codec.Call[T](ctx, c, method, args...)
}

// CallInto is the same as Call, except instead of returning, you provide a pointer to the result
var CallInto = codec.CallInto
