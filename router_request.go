package jrpc

import (
	"context"

	json "github.com/goccy/go-json"
	jsoniter "github.com/json-iterator/go"
)

type Request struct {
	ctx context.Context
	msg jsonrpcMessage

	peer PeerInfo
}

func NewRequest(ctx context.Context, id string, method string, params any) *Request {
	r := &Request{ctx: ctx}
	pms, _ := json.Marshal(params)
	r.msg = jsonrpcMessage{
		ID:     NewStringIDPtr(id),
		Method: method,
		Params: pms,
	}
	return r
}

func (r *Request) Method() string {
	return r.msg.Method
}

func (r *Request) Params() json.RawMessage {
	return r.msg.Params
}

func (r *Request) ParamSlice() []any {
	var params []any
	json.Unmarshal(r.msg.Params, &params)
	return params
}

func (r *Request) ParamArray(a ...any) error {
	var params []json.RawMessage
	json.Unmarshal(r.msg.Params, &params)
	for idx, v := range params {
		if len(v) > idx {
			err := json.Unmarshal(v, &a[idx])
			if err != nil {
				return err
			}
		} else {
			break
		}
	}
	return nil
}

func (r *Request) ParamInto(v any) error {
	return json.Unmarshal(r.msg.Params, &v)
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

var jpool = jsoniter.NewIterator(jsoniter.ConfigCompatibleWithStandardLibrary).Pool()

func (r *Request) Iter(fn func(j *jsoniter.Iterator) error) error {
	it := jpool.BorrowIterator(r.Params())
	defer jpool.ReturnIterator(it)
	return fn(it)
}
