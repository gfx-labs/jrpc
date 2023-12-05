package jsonrpc

import (
	"context"
	"encoding/json"
)

// http.ResponseWriter interface, but for jrpc
type ResponseWriter interface {
	Send(v any, err error) error
	Notify(method string, v any) error
}

type StreamingResponseWriter interface {
	ResponseWriter
	SendStream(func(MessageStreamer) error) error
	NotifyStream(func(MessageStreamer) error) error
}

type Request struct {
	ID     *ID             `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Peer   PeerInfo        `json:"-"`

	ctx context.Context
}

func NewRawRequest(ctx context.Context, id *ID, method string, params json.RawMessage) (r *Request) {
	if ctx == nil {
		ctx = context.Background()
	}
	r = &Request{ctx: ctx}
	r.ID = id
	r.Method = method
	r.Params = params
	return r
}

// NewRequest makes a new request
func NewRequest(ctx context.Context, id *ID, method string, params any) (r *Request, err error) {
	raw, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	return NewRawRequest(ctx, id, method, raw), nil
}

func (r *Request) Context() context.Context {
	return r.ctx
}

func (r *Request) WithContext(ctx context.Context) *Request {
	if ctx == nil {
		panic("nil context")
	}
	r.ctx = ctx
	r2 := new(Request)
	*r2 = *r
	r2.ctx = ctx
	r2.ID = r.ID
	r2.Method = r.Method
	r2.Params = r.Params
	r2.Peer = r.Peer
	return r2
}
