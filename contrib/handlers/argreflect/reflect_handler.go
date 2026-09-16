package argreflect

import (
	"context"
	"reflect"
	"unicode"

	"github.com/gfx-labs/jrpc/pkg/jsonrpc"
)

var (
	contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
	errorType   = reflect.TypeOf((*error)(nil)).Elem()
)

func SuitableCallbacks(receiver reflect.Value) map[string]jsonrpc.Handler {
	return suitableCallbacks(receiver)
}

// suitableCallbacks iterates over the methods of the given type. It determines if a method
// satisfies the criteria for a RPC callback or a subscription callback and adds it to the
// collection of callbacks. See server documentation for a summary of these criteria.
func suitableCallbacks(receiver reflect.Value) map[string]jsonrpc.Handler {
	typ := receiver.Type()
	callbacks := make(map[string]jsonrpc.Handler)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		if method.PkgPath != "" {
			continue // method not exported
		}
		cb := newCallback(receiver, method.Func)
		if cb == nil {
			continue // function invalid
		}
		name := formatName(method.Name)
		callbacks[name] = cb
	}
	return callbacks
}

func NewCallback(receiver, fn reflect.Value) jsonrpc.Handler {
	return newCallback(receiver, fn)
}

// newCallback turns fn (a function) into a handler. It returns nil if the function
// is unsuitable as an RPC callback.
func newCallback(receiver, fn reflect.Value) jsonrpc.Handler {
	fntype := fn.Type()
	c := &callback{fn: fn, rcvr: receiver, errPos: -1}
	// Determine parameter types. They must all be exported or builtin types.
	if x := c.makeArgTypes(); x != nil {
		return x
	}

	// Verify return types. The function must return at most one error
	// and/or one other non-error value.
	outs := make([]reflect.Type, fntype.NumOut())
	for i := 0; i < fntype.NumOut(); i++ {
		outs[i] = fntype.Out(i)
	}
	if len(outs) > 2 {
		return nil
	}
	// If an error is returned, it must be the last returned value.
	switch {
	case len(outs) == 1 && isErrorType(outs[0]):
		c.errPos = 0
	case len(outs) == 2:
		if isErrorType(outs[0]) || !isErrorType(outs[1]) {
			return nil
		}
		c.errPos = 1
	}
	return c
}

// callback is a method callback which was registered in the server
type callback struct {
	fn       reflect.Value  // the function
	rcvr     reflect.Value  // receiver object of method, set if fn is method
	argTypes []reflect.Type // input argument types
	hasCtx   bool           // method's first argument is a context (not included in argTypes)
	errPos   int            // err return idx, of -1 when method cannot return error
}

// callback handler implements handler for the original receiver style that geth used
func (e *callback) ServeRPC(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
	argTypes := append([]reflect.Type{}, e.argTypes...)
	args, err := parsePositionalArguments(r.Params, argTypes)
	if err != nil {
		w.Send(nil, jsonrpc.NewInvalidParamsError(err.Error()))
		return
	}
	// Create the argument slice.
	fullargs := make([]reflect.Value, 0, 2+len(args))
	if e.rcvr.IsValid() {
		fullargs = append(fullargs, e.rcvr)
	}
	if e.hasCtx {
		fullargs = append(fullargs, reflect.ValueOf(r.Context()))
	}
	fullargs = append(fullargs, args...)
	// Run the callback.
	results := e.fn.Call(fullargs)
	if e.errPos >= 0 && !results[e.errPos].IsNil() {
		// Method has returned non-nil error value.
		err := results[e.errPos].Interface().(error)
		w.Send(nil, err)
		return
	}
	if len(results) == 0 {
		w.Send(jsonrpc.Null, nil)
		return
	}
	w.Send(results[0].Interface(), nil)
}

var jrpcResponseWriterType = reflect.TypeOf((*jsonrpc.ResponseWriter)(nil)).Elem()
var jrpcRequestType = reflect.TypeOf(&jsonrpc.Request{})

// makeArgTypes composes the argTypes list.
func (c *callback) makeArgTypes() jsonrpc.Handler {
	fntype := c.fn.Type()
	// Skip receiver and context.Context parameter (if present).
	firstArg := 0
	if c.rcvr.IsValid() {
		firstArg++
	}
	if fntype.NumIn() > firstArg && fntype.In(firstArg) == contextType {
		c.hasCtx = true
		firstArg++
	}

	// Add all remaining parameters.
	c.argTypes = make([]reflect.Type, fntype.NumIn()-firstArg)
	for i := firstArg; i < fntype.NumIn(); i++ {
		c.argTypes[i-firstArg] = fntype.In(i)
	}
	if len(c.argTypes) == 2 {
		// special case to see if it is a valid jsonrpc.Handler
		if c.argTypes[1].AssignableTo(jrpcRequestType) && c.argTypes[0].Implements(jrpcResponseWriterType) {
			if !c.rcvr.IsValid() {
				cb, ok := c.fn.Interface().(func(jsonrpc.ResponseWriter, *jsonrpc.Request))
				if !ok {
					panic("invalid callback registered")
				}
				return jsonrpc.HandlerFunc(cb)
			} else {
				return jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
					c.fn.Call([]reflect.Value{c.rcvr, reflect.ValueOf(w), reflect.ValueOf(r)})
				})
			}
		}

	}
	return nil
}

// Does t satisfy the error interface?
func isErrorType(t reflect.Type) bool {
	return t.Implements(errorType)
}

// formatName converts to first character of name to lowercase.
func formatName(name string) string {
	ret := []rune(name)
	if len(ret) > 0 {
		ret[0] = unicode.ToLower(ret[0])
	}
	return string(ret)
}
