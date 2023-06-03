package jmux

import (
	"gfx.cafe/open/jrpc/pkg/codec"
)

// Chain returns a Middlewares type from a slice of middleware handlers.
func Chain(middlewares ...func(codec.Handler) codec.Handler) Middlewares {
	return Middlewares(middlewares)
}

// Handler builds and returns a Handler from the chain of middlewares,
// with `h Handler` as the final handler.
func (mws Middlewares) Handler(h codec.Handler) codec.Handler {
	return &ChainHandler{h, chain(mws, h), mws}
}

// HandlerFunc builds and returns a Handler from the chain of middlewares,
// with `h Handler` as the final handler.
func (mws Middlewares) HandlerFunc(h codec.HandlerFunc) codec.Handler {
	return &ChainHandler{h, chain(mws, h), mws}
}

// ChainHandler is a Handler with support for handler composition and
// execution.
type ChainHandler struct {
	Endpoint    codec.Handler
	chain       codec.Handler
	Middlewares Middlewares
}

func (c *ChainHandler) ServeRPC(w codec.ResponseWriter, r *codec.Request) {
	c.chain.ServeRPC(w, r)
}

// chain builds a Handler composed of an inline middleware stack and endpoint
// handler in the order they are passed.
func chain(middlewares []func(codec.Handler) codec.Handler, endpoint codec.Handler) codec.Handler {
	// Return ahead of time if there aren't any middlewares for the chain
	if len(middlewares) == 0 {
		return endpoint
	}
	// Wrap the end handler with the middleware chain
	h := middlewares[len(middlewares)-1](endpoint)
	for i := len(middlewares) - 2; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
