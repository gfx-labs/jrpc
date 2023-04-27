package argreflect

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"unicode"

	"gfx.cafe/open/jrpc"
	"tuxpa.in/a/zlog/log"
)

var (
	contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
	errorType   = reflect.TypeOf((*error)(nil)).Elem()
)

func SuitableCallbacks(receiver reflect.Value) map[string]jrpc.Handler {
	return suitableCallbacks(receiver)
}

// suitableCallbacks iterates over the methods of the given type. It determines if a method
// satisfies the criteria for a RPC callback or a subscription callback and adds it to the
// collection of callbacks. See server documentation for a summary of these criteria.
func suitableCallbacks(receiver reflect.Value) map[string]jrpc.Handler {
	typ := receiver.Type()
	callbacks := make(map[string]jrpc.Handler)
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

// callback is a method callback which was registered in the server
type callback struct {
	fn       reflect.Value  // the function
	rcvr     reflect.Value  // receiver object of method, set if fn is method
	argTypes []reflect.Type // input argument types
	hasCtx   bool           // method's first argument is a context (not included in argTypes)
	errPos   int            // err return idx, of -1 when method cannot return error
}

// callback handler implements handler for the original receiver style that geth used
func (e *callback) ServeRPC(w jrpc.ResponseWriter, r *jrpc.Request) {
	argTypes := append([]reflect.Type{}, e.argTypes...)
	args, err := parsePositionalArguments(r.Params, argTypes)
	if err != nil {
		w.Send(nil, jrpc.NewInvalidParamsError(err.Error()))
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
	//Catch panic while running the callback.
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Error().Str("method", r.Method).Interface("err", err).Hex("buf", buf).Msg("crashed")
			//		errRes := errors.New("method handler crashed: " + fmt.Sprint(err))
			w.Send(nil, fmt.Errorf("%s", err))
			return
		}
	}()
	// Run the callback.
	results := e.fn.Call(fullargs)
	if e.errPos >= 0 && !results[e.errPos].IsNil() {
		// Method has returned non-nil error value.
		err := results[e.errPos].Interface().(error)
		w.Send(nil, err)
		return
	}
	w.Send(results[0].Interface(), nil)
}

func NewCallback(receiver, fn reflect.Value) jrpc.Handler {
	return newCallback(receiver, fn)
}

// newCallback turns fn (a function) into a callback object. It returns nil if the function
// is unsuitable as an RPC callback.
func newCallback(receiver, fn reflect.Value) *callback {
	fntype := fn.Type()
	c := &callback{fn: fn, rcvr: receiver, errPos: -1}
	// Determine parameter types. They must all be exported or builtin types.
	c.makeArgTypes()

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

// makeArgTypes composes the argTypes list.
func (c *callback) makeArgTypes() {
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
}

// Does t satisfy the error interface?
func isErrorType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
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
