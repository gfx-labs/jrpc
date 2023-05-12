package jmux

import "gfx.cafe/open/jrpc"

// Chain returns a Middlewares type from a slice of middleware handlers.
func Chain(middlewares ...func(jrpc.Handler) jrpc.Handler) Middlewares {
	return Middlewares(middlewares)
}

// Handler builds and returns a Handler from the chain of middlewares,
// with `h Handler` as the final handler.
func (mws Middlewares) Handler(h jrpc.Handler) jrpc.Handler {
	return &ChainHandler{h, chain(mws, h), mws}
}

// HandlerFunc builds and returns a Handler from the chain of middlewares,
// with `h Handler` as the final handler.
func (mws Middlewares) HandlerFunc(h jrpc.HandlerFunc) jrpc.Handler {
	return &ChainHandler{h, chain(mws, h), mws}
}

// ChainHandler is a Handler with support for handler composition and
// execution.
type ChainHandler struct {
	Endpoint    jrpc.Handler
	chain       jrpc.Handler
	Middlewares Middlewares
}

func (c *ChainHandler) ServeRPC(w jrpc.ResponseWriter, r *jrpc.Request) {
	c.chain.ServeRPC(w, r)
}

// chain builds a Handler composed of an inline middleware stack and endpoint
// handler in the order they are passed.
func chain(middlewares []func(jrpc.Handler) jrpc.Handler, endpoint jrpc.Handler) jrpc.Handler {
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
