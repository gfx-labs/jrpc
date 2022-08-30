package jrpc

import (
	"context"
	"encoding/json"
	"io"

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
