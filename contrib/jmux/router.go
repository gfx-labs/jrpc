package jmux

import (
	"gfx.cafe/open/jrpc/pkg/codec"
)

// NewRouter returns a new Mux object that implements the Router interface.
func NewRouter() *Mux {
	return NewMux()
}

type StructReflector interface {
	// mimics the behavior of the handlers in the go-ethereum rpc package
	// if you don't know how to use this, just use the chi-like interface instead.
	RegisterStruct(pattern string, rcvr any) error
	// mimics the behavior of the handlers in the go-ethereum rpc package
	// if you don't know how to use this, just use the chi-like interface instead.
	RegisterFunc(pattern string, rcvr any) error
}

// Router consisting of the core routing methods used by chi's Mux,
// adapted to fit json-rpc.
type Router interface {
	codec.Handler
	Routes
	StructReflector

	// Use appends one or more middlewares onto the Router stack.
	Use(middlewares ...func(codec.Handler) codec.Handler)

	// With adds inline middlewares for an endpoint handler.
	With(middlewares ...func(codec.Handler) codec.Handler) Router

	// Group adds a new inline-Router along the current routing
	// path, with a fresh middleware stack for the inline-Router.
	Group(fn func(r Router)) Router

	// Route mounts a sub-Router along a `pattern`` string.
	Route(pattern string, fn func(r Router)) Router

	// Mount attaches another Handler along ./pattern/*
	Mount(pattern string, h codec.Handler)

	// Handle and HandleFunc adds routes for `pattern` that matches
	// all HTTP methods.
	Handle(pattern string, h codec.Handler)
	HandleFunc(pattern string, h codec.HandlerFunc)

	// NotFound defines a handler to respond whenever a route could
	// not be found.
	NotFound(h codec.HandlerFunc)
}

// Routes interface adds two methods for router traversal, which is also
// used by the `docgen` subpackage to generation documentation for Routers.
type Routes interface {
	// Routes returns the routing tree in an easily traversable structure.
	Routes() []Route

	// Middlewares returns the list of middlewares in use by the router.
	Middlewares() Middlewares

	// Match searches the routing tree for a handler that matches
	// the method/path - similar to routing a http request, but without
	// executing the handler thereafter.
	Match(rctx *Context, path string) bool
}

// Middlewares type is a slice of standard middleware handlers with methods
// to compose middleware chains and Handler's.
type Middlewares = codec.Middlewares
