package jrpc

import (
	"context"

	"gfx.cafe/open/jrpc/pkg/codec"
	"gfx.cafe/open/jrpc/pkg/server"
)

// to make the repo cleaner, we export everything here. this way the packages dont need to ever import this

type (
	// Conn
	Conn = codec.Conn
	// Handler
	Handler = codec.Handler
	// HandlerFunc
	HandlerFunc = codec.HandlerFunc
	// ResponseWriter
	ResponseWriter = codec.ResponseWriter
	// StreamingConn
	StreamingConn = codec.Conn
)
type (
	// BatchElem
	BatchElem = codec.BatchElem
)

var (
	ContextWithConn     = codec.ContextWithConn
	ContextWithPeerInfo = server.ContextWithPeerInfo

	ConnFromContext     = codec.ConnFromContext
	PeerInfoFromContext = server.PeerInfoFromContext
)

// Do
func Do[T any](ctx context.Context, c Conn, method string, args any) (*T, error) {
	return codec.Do[T](ctx, c, method, args)
}

// Call
func Call[T any](ctx context.Context, c Conn, method string, args ...any) (*T, error) {
	return codec.Call[T](ctx, c, method, args...)
}

// CallInto
var CallInto = codec.CallInto
