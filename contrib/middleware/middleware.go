package middleware

import (
	"gfx.cafe/open/jrpc/pkg/codec"
)

// New will create a new middleware handler from a jrpc.Handler.
func New(h codec.Handler) func(next codec.Handler) codec.Handler {
	return func(next codec.Handler) codec.Handler {
		return codec.HandlerFunc(func(w codec.ResponseWriter, r *codec.Request) {
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
