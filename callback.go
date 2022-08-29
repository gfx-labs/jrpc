package jrpc

import (
	"context"
	"encoding/json"
	"io"
	"reflect"
	"runtime"

	"git.tuxpa.in/a/zlog/log"
	jsoniter "github.com/json-iterator/go"
)

type HandlerFunc func(w ResponseWriter, r *Request)

func (fn HandlerFunc) ServeRPC(w ResponseWriter, r *Request) {
	(fn)(w, r)
}

type Handler interface {
	ServeRPC(w ResponseWriter, r *Request)
}

type Request struct {
	ctx context.Context
	msg jsonrpcMessage

	peer PeerInfo
}

func (r *Request) Method() string {
	return r.msg.Method
}

func (r *Request) Params() json.RawMessage {
	return r.msg.Params
}

func (r *Request) ParamSlice() []any {
	var params []any
	jsoniter.Unmarshal(r.msg.Params, &params)
	return params
}

func (r *Request) ParamInto(v any) error {
	return jsoniter.Unmarshal(r.msg.Params, &v)
}

func (r *Request) Context() context.Context {
	return r.ctx
}

func (r *Request) Remote() string {
	return r.peer.RemoteAddr
}

func (r *Request) Peer() PeerInfo {
	return r.peer
}

func (r *Request) WithContext(ctx context.Context) *Request {
	if ctx == nil {
		panic("nil context")
	}
	r.ctx = ctx
	r2 := new(Request)
	*r2 = *r
	r2.ctx = ctx
	r2.msg = r.msg
	return r2
}

func (r *Request) Msg() jsonrpcMessage {
	return r.msg
}

type ResponseWriter interface {
	Send(v any, err error) error
}

type ResponseWriterIo struct {
	r *Request
	w io.Writer
}

func NewReaderResponseWriterIo(r *Request, w io.Writer) ResponseWriter {
	return &ResponseWriterIo{
		w: w,
		r: r,
	}
}

func (w *ResponseWriterIo) Send(args any, e error) (err error) {
	enc := jsoniter.ConfigCompatibleWithStandardLibrary.NewEncoder(w.w)
	if e != nil {
		return enc.Encode(errorMessage(e))
	}
	return enc.Encode(args)
}

type ResponseWriterMsg struct {
	r   *Request
	msg *jsonrpcMessage
}

func NewReaderResponseWriterMsg(r *Request) *ResponseWriterMsg {
	return &ResponseWriterMsg{
		r: r,
	}
}

func (w *ResponseWriterMsg) Send(args any, e error) (err error) {
	cm := w.r.Msg()
	if e != nil {
		w.msg = cm.errorResponse(e)
		return nil
	}
	w.msg = cm.response(args)
	return nil
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
func (e *callback) ServeRPC(w ResponseWriter, r *Request) {
	argTypes := append([]reflect.Type{}, e.argTypes...)
	args, err := parsePositionalArguments(r.msg.Params, argTypes)
	if err != nil {
		w.Send(nil, &invalidParamsError{err.Error()})
		return
	}
	// Create the argument slice.
	fullargs := make([]reflect.Value, 0, 2+len(args))
	if e.rcvr.IsValid() {
		fullargs = append(fullargs, e.rcvr)
	}
	if e.hasCtx {
		fullargs = append(fullargs, reflect.ValueOf(r.ctx))
	}
	fullargs = append(fullargs, args...)
	// Catch panic while running the callback.
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Error().Str("method", r.msg.Method).Interface("err", err).Hex("buf", buf).Msg("crashed")
			//		errRes := errors.New("method handler crashed: " + fmt.Sprint(err))
			w.Send(nil, nil)
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
	return
}
