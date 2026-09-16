package jrpc

import (
	"context"

	"github.com/gfx-labs/jrpc/contrib/codecs"
	"github.com/gfx-labs/jrpc/pkg/jsonrpc"
	"github.com/gfx-labs/jrpc/pkg/server"
)

// Conn is used to make requests to jsonrpc2 servers
type Conn = jsonrpc.Conn

// Handler is the equivalent of http.Handler, but for jsonrpc.
type Handler = jsonrpc.Handler

// ResponseWriter is used to write responses to the request
type ResponseWriter = jsonrpc.ResponseWriter

type (
	// HandlerFunc is a Handler that exists as a function
	HandlerFunc = jsonrpc.HandlerFunc
	// Request is the request object
	Request = jsonrpc.Request
	// Server is a jrpc server
	// Middleware is a middleware
	Middleware = func(Handler) Handler
)

var (

	// DialContext is to dial a conn with context
	DialContext = codecs.DialContext
	// Dial is to dial a conn with context.Background()
	Dial = codecs.Dial

	// ContextWithConn will attach a conn to the context
	ContextWithConn = jsonrpc.ContextWithConn
	// ContextWithPeerInfo will attach a peerinfo to the context
	ContextWithPeerInfo = server.ContextWithPeerInfo

	// ConnFromContext will retrieve a conn from context
	ConnFromContext = jsonrpc.ConnFromContext
	// PeerInfoFromContext will retrieve a peerinfo from context
	PeerInfoFromContext = server.PeerInfoFromContext
)

// Do will use the conn to perform a jsonrpc2 call.
func Do[T any](ctx context.Context, c Conn, method string, args any) (*T, error) {
	return jsonrpc.Do[T](ctx, c, method, args)
}

// Call will use the conn to perform a jsonrpc2 call, except the args will be encoded as an array of arguments.
func Call[T any](ctx context.Context, c Conn, method string, args ...any) (*T, error) {
	return jsonrpc.Call[T](ctx, c, method, args...)
}

// CallInto is the same as Call, except instead of returning, you provide a pointer to the result
var CallInto = jsonrpc.CallInto
