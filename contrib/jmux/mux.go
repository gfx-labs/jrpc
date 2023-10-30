package jmux

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"gfx.cafe/open/jrpc/contrib/handlers/argreflect"
	"gfx.cafe/open/jrpc/pkg/jsonrpc"
)

var _ Router = &Mux{}

const sepRune = '/'
const sepString = string(sepRune)

// Mux is a simple JRPC route multiplexer that parses a request path,
// records any URL params, and executes an end handler. It implements
// the Handler interface and is friendly with the standard library.
//
// Mux is designed to be fast, minimal and offer a powerful API for building
// modular and composable JRPC services with a large set of handlers. It's
// particularly useful for writing large REST API services that break a handler
// into many smaller parts composed of middlewares and end handlers.
type Mux struct {
	// The computed mux handler made of the chained middleware stack and
	// the tree router
	handler jsonrpc.Handler

	// The radix trie router
	tree *node

	// Custom method not allowed handler
	methodNotAllowedHandler jsonrpc.HandlerFunc

	// A reference to the parent mux used by subrouters when mounting
	// to a parent mux
	parent *Mux

	// Routing context pool
	pool *sync.Pool

	// Custom route not found handler
	notFoundHandler jsonrpc.HandlerFunc

	// The middleware stack
	middlewares []func(jsonrpc.Handler) jsonrpc.Handler

	// Controls the behaviour of middleware chain generation when a mux
	// is registered as an inline group inside another mux.
	inline bool
}

// NewMux returns a newly initialized Mux object that implements the Router
// interface.
func NewMux() *Mux {
	mux := &Mux{tree: &node{}, pool: &sync.Pool{}}
	mux.pool.New = func() interface{} {
		return NewRouteContext()
	}
	return mux
}

func (m *Mux) RegisterStruct(name string, rcvr any) error {
	rcvrVal := reflect.ValueOf(rcvr)
	if name == "" {
		return fmt.Errorf("no service name for type %s", rcvrVal.Type().String())
	}
	callbacks := argreflect.SuitableCallbacks(rcvrVal)
	if len(callbacks) == 0 {
		return fmt.Errorf("service %T doesn't have any suitable methods/subscriptions to expose", rcvr)
	}
	m.Route(name, func(r Router) {
		for nm, cb := range callbacks {
			r.Handle(nm, cb)
		}
	})
	return nil
}
func (m *Mux) RegisterFunc(name string, rcvr any) error {
	rcvrVal := reflect.ValueOf(rcvr)
	if name == "" {
		return fmt.Errorf("no service name for type %s", rcvrVal.Type().String())
	}
	cb := argreflect.NewCallback(reflect.ValueOf(nil), rcvrVal)
	m.Mount(name, cb)
	return nil
}

// ServeRPC is the single method of the Handler interface that makes
// Mux interoperable with the standard library. It uses a sync.Pool to get and
// reuse routing contexts for each request.
func (mx *Mux) ServeRPC(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
	// Ensure the mux has some routes defined on the mux
	if mx.handler == nil {
		mx.NotFoundHandler().ServeRPC(w, r)
		return
	}

	// Check if a routing context already exists from a parent router.
	rctx, _ := r.Context().Value(RouteCtxKey).(*Context)
	if rctx != nil {
		mx.handler.ServeRPC(w, r)
		return
	}

	// Fetch a RouteContext object from the sync pool, and call the computed
	// mx.handler that is comprised of mx.middlewares + mx.route
	// Once the request is finished, reset the routing context and put it back
	// into the pool for reuse from another request.
	rctx = mx.pool.Get().(*Context)
	rctx.Reset()
	rctx.Routes = mx
	rctx.parentCtx = r.Context()

	// NOTE: r.WithContext() causes 2 allocations and context.WithValue() causes 1 allocation
	r = r.WithContext(context.WithValue(r.Context(), RouteCtxKey, rctx))

	// Serve the request and once its done, put the request context back in the sync pool
	mx.handler.ServeRPC(w, r)
	mx.pool.Put(rctx)
}

// Use appends a middleware handler to the Mux middleware stack.
//
// The middleware stack for any Mux will execute before searching for a matching
// route to a specific handler, which provides opportunity to respond early,
// change the course of the request execution, or set request-scoped values for
// the next Handler.
func (mx *Mux) Use(middlewares ...func(jsonrpc.Handler) jsonrpc.Handler) {
	if mx.handler != nil {
		panic("chi: all middlewares must be defined before routes on a mux")
	}
	mx.middlewares = append(mx.middlewares, middlewares...)
}

// Handle adds the route `pattern` that matches any jrpc method to
// execute the `handler` Handler.
func (mx *Mux) Handle(pattern string, handler jsonrpc.Handler) {
	mx.handle(pattern, handler)
}

// HandleFunc adds the route `pattern` that matches any jrpc method to
// execute the `handlerFn` HandlerFunc.
func (mx *Mux) HandleFunc(pattern string, handlerFn jsonrpc.HandlerFunc) {
	mx.handle(pattern, handlerFn)
}

// NotFound sets a custom HandlerFunc for routing paths that could
// not be found. The default 404 handler is `NotFound`.
func (mx *Mux) NotFound(handlerFn jsonrpc.HandlerFunc) {
	// Build NotFound handler chain
	m := mx
	hFn := handlerFn
	if mx.inline && mx.parent != nil {
		m = mx.parent
		hFn = Chain(mx.middlewares...).HandlerFunc(hFn).ServeRPC
	}

	// Update the notFoundHandler from this point forward
	m.notFoundHandler = hFn
	m.updateSubRoutes(func(subMux *Mux) {
		if subMux.notFoundHandler == nil {
			subMux.NotFound(hFn)
		}
	})
}

// MethodNotAllowed sets a custom HandlerFunc for routing paths where the
// method is unresolved. The default handler returns a 405 with an empty body.
func (mx *Mux) MethodNotAllowed(handlerFn jsonrpc.HandlerFunc) {
	// Build MethodNotAllowed handler chain
	m := mx
	hFn := handlerFn
	if mx.inline && mx.parent != nil {
		m = mx.parent
		hFn = Chain(mx.middlewares...).HandlerFunc(hFn).ServeRPC
	}

	// Update the methodNotAllowedHandler from this point forward
	m.methodNotAllowedHandler = hFn
	m.updateSubRoutes(func(subMux *Mux) {
		if subMux.methodNotAllowedHandler == nil {
			subMux.MethodNotAllowed(hFn)
		}
	})
}

// With adds inline middlewares for an endpoint handler.
func (mx *Mux) With(middlewares ...func(jsonrpc.Handler) jsonrpc.Handler) Router {
	// Similarly as in handle(), we must build the mux handler once additional
	// middleware registration isn't allowed for this stack, like now.
	if !mx.inline && mx.handler == nil {
		mx.updateRouteHandler()
	}

	// Copy middlewares from parent inline muxs
	var mws Middlewares
	if mx.inline {
		mws = make(Middlewares, len(mx.middlewares))
		copy(mws, mx.middlewares)
	}
	mws = append(mws, middlewares...)

	im := &Mux{
		pool: mx.pool, inline: true, parent: mx, tree: mx.tree, middlewares: mws,
		notFoundHandler: mx.notFoundHandler, methodNotAllowedHandler: mx.methodNotAllowedHandler,
	}
	return im
}

// Group creates a new inline-Mux with a fresh middleware stack. It's useful
// for a group of handlers along the same routing path that use an additional
// set of middlewares. See _examples/.
func (mx *Mux) Group(fn func(r Router)) Router {
	im := mx.With().(*Mux)
	if fn != nil {
		fn(im)
	}
	return im
}

// Route creates a new Mux with a fresh middleware stack and mounts it
// along the `pattern` as a subrouter. Effectively, this is a short-hand
// call to Mount. See _examples/.
func (mx *Mux) Route(pattern string, fn func(r Router)) Router {
	if fn == nil {
		panic(fmt.Sprintf("chi: attempting to Route() a nil subrouter on '%s'", pattern))
	}
	subRouter := NewRouter()
	fn(subRouter)
	mx.Mount(pattern, subRouter)
	return subRouter
}

// Mount attaches another Handler or chi Router as a subrouter along a routing
// path. It's very useful to split up a large API as many independent routers and
// compose them as a single service using Mount. See _examples/.
//
// Note that Mount() simply sets a wildcard along the `pattern` that will continue
// routing at the `handler`, which in most cases is another chi.Router. As a result,
// if you define two Mount() routes on the exact same pattern the mount will panic.
func (mx *Mux) Mount(pattern string, handler jsonrpc.Handler) {
	if handler == nil {
		panic(fmt.Sprintf("chi: attempting to Mount() a nil handler on '%s'", pattern))
	}

	// Provide runtime safety for ensuring a pattern isn't mounted on an existing
	// routing pattern.
	if mx.tree.findPattern(pattern+"*") || mx.tree.findPattern(pattern+sepString+"*") {
		panic(fmt.Sprintf("chi: attempting to Mount() a handler on an existing path, '%s'", pattern))
	}

	// Assign sub-Router's with the parent not found & method not allowed handler if not specified.
	subr, ok := handler.(*Mux)
	if ok && subr.notFoundHandler == nil && mx.notFoundHandler != nil {
		subr.NotFound(mx.notFoundHandler)
	}
	if ok && subr.methodNotAllowedHandler == nil && mx.methodNotAllowedHandler != nil {
		subr.MethodNotAllowed(mx.methodNotAllowedHandler)
	}

	mountHandler := jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		rctx := RouteContext(r.Context())

		// shift the url path past the previous subrouter
		rctx.RoutePath = mx.nextRoutePath(rctx)

		// reset the wildcard URLParam which connects the subrouter
		n := len(rctx.MethodParams.Keys) - 1
		if n >= 0 && rctx.MethodParams.Keys[n] == "*" && len(rctx.MethodParams.Values) > n {
			rctx.MethodParams.Values[n] = ""
		}

		handler.ServeRPC(w, r)
	})

	if pattern == "" || pattern[len(pattern)-1] != sepRune {
		mx.handle(pattern, mountHandler)
		mx.handle(pattern+sepString, mountHandler)
		pattern += sepString
	}

	n := mx.handle(pattern+"*", mountHandler)
	subroutes, _ := handler.(Routes)
	if subroutes != nil {
		n.subroutes = subroutes
	}
}

// Routes returns a slice of routing information from the tree,
// useful for traversing available routes of a router.
func (mx *Mux) Routes() []Route {
	return mx.tree.routes()
}

// Middlewares returns a slice of middleware handler functions.
func (mx *Mux) Middlewares() Middlewares {
	return mx.middlewares
}

// Match searches the routing tree for a handler that matches the method/path.
// It's similar to routing a jrpc request, but without executing the handler
// thereafter.
//
// Note: the *Context state is updated during execution, so manage
// the state carefully or make a NewRouteContext().
func (mx *Mux) Match(rctx *Context, path string) bool {
	node, _, h := mx.tree.FindRoute(rctx, path)
	if node != nil && node.subroutes != nil {
		rctx.RoutePath = mx.nextRoutePath(rctx)
		return node.subroutes.Match(rctx, rctx.RoutePath)
	}
	return h != nil
}

// NotFoundHandler returns the default Mux 404 responder whenever a route
// cannot be found.
func (mx *Mux) NotFoundHandler() jsonrpc.HandlerFunc {
	if mx.notFoundHandler != nil {
		return mx.notFoundHandler
	}
	return NotFound
}

// MethodNotAllowedHandler returns the default Mux 405 responder whenever
// a method cannot be resolved for a route.
func (mx *Mux) MethodNotAllowedHandler() jsonrpc.HandlerFunc {
	if mx.methodNotAllowedHandler != nil {
		return mx.methodNotAllowedHandler
	}
	return methodNotAllowedHandler
}

// handle registers a Handler in the routing tree for a particular jrpc method
// and routing pattern.
func (mx *Mux) handle(pattern string, handler jsonrpc.Handler) *node {
	if len(pattern) == 0 {
		panic(fmt.Sprintf("rpc: routing pattern must not be empty in '%s'", pattern))
	}

	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}

	// Build the computed routing handler for this routing pattern.
	if !mx.inline && mx.handler == nil {
		mx.updateRouteHandler()
	}

	// Build endpoint handler with inline middlewares for the route
	var h jsonrpc.Handler
	if mx.inline {
		mx.handler = jsonrpc.HandlerFunc(mx.routeRPC)
		h = Chain(mx.middlewares...).Handler(handler)
	} else {
		h = handler
	}

	// Add the endpoint to the tree and return the node
	return mx.tree.InsertRoute(pattern, h)
}

// routeRPC routes a Request through the Mux routing tree to serve
// the matching handler for a particular jrpc method.
func (mx *Mux) routeRPC(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
	// Grab the route context object
	rctx := r.Context().Value(RouteCtxKey).(*Context)

	// The request routing path
	routePath := rctx.RoutePath
	if routePath == "" {
		routePath = r.Method
		if routePath == "" {
			routePath = sepString
		}
	}

	// Find the route
	if _, _, h := mx.tree.FindRoute(rctx, routePath); h != nil {
		h.ServeRPC(w, r)
		return
	}
	if rctx.methodNotAllowed {
		mx.MethodNotAllowedHandler().ServeRPC(w, r)
		return
	} else {
		mx.NotFoundHandler().ServeRPC(w, r)
		return
	}
}

func (mx *Mux) nextRoutePath(rctx *Context) string {
	routePath := sepString
	nx := len(rctx.routeParams.Keys) - 1 // index of last param in list
	if nx >= 0 && rctx.routeParams.Keys[nx] == "*" && len(rctx.routeParams.Values) > nx {
		routePath = sepString + rctx.routeParams.Values[nx]
	}
	return routePath
}

// Recursively update data on child routers.
func (mx *Mux) updateSubRoutes(fn func(subMux *Mux)) {
	for _, r := range mx.tree.routes() {
		subMux, ok := r.SubRoutes.(*Mux)
		if !ok {
			continue
		}
		fn(subMux)
	}
}

// updateRouteHandler builds the single mux handler that is a chain of the middleware
// stack, as defined by calls to Use(), and the tree router (Mux) itself. After this
// point, no other middlewares can be registered on this Mux's stack. But you can still
// compose additional middlewares via Group()'s or using a chained middleware handler.
func (mx *Mux) updateRouteHandler() {
	mx.handler = jsonrpc.ChainMiddlewares(mx.middlewares, jsonrpc.HandlerFunc(mx.routeRPC))
}

// methodNotAllowedHandler is a helper function to respond with a 405,
// method not allowed.
func methodNotAllowedHandler(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
	w.Send(nil, errors.New("forbidden"))
}

func NotFound(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
	w.Send(nil, jsonrpc.NewMethodNotFoundError(r.Method))
}
