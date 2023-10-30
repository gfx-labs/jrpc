package middleware

import (
	"gfx.cafe/open/jrpc/pkg/jsonrpc"
)

// New will create a new middleware handler from a jrpc.Handler.
func New(h jsonrpc.Handler) func(next jsonrpc.Handler) jsonrpc.Handler {
	return func(next jsonrpc.Handler) jsonrpc.Handler {
		return jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
			h.ServeRPC(w, r)
		})
	}
}

// contextKey is a value for use with context.WithValue. It's used as
// a pointer so it fits in an interface{} without allocation. This technique
// for defining context keys was copied from Go 1.7's new use of context in net/jrpc.
type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "jrpc/middleware context value " + k.name
}
