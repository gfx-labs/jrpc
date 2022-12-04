package jrpc

import (
	"context"

	json "github.com/goccy/go-json"
	jsoniter "github.com/json-iterator/go"
)

type Request struct {
	ID     *ID             `json:"id,omitempty"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`

	Peer PeerInfo `json:"-"`

	Version version `json:"jsonrpc"`

	ctx context.Context
}

func NewMsgRequest(ctx context.Context, peer PeerInfo, msg jsonrpcMessage) *Request {
	r := &Request{ctx: ctx}
	r.ID = msg.ID
	r.Method = msg.Method
	r.Params = msg.Params
	r.Peer = peer
	r.ctx = ctx
	return r
}

func NewRequest(ctx context.Context, id string, method string, params any) *Request {
	r := &Request{ctx: ctx}
	pms, _ := json.Marshal(params)
	r.ID = NewStringIDPtr(id)
	r.Method = method
	r.Params = pms
	return r
}

func (r *Request) ParamSlice() []any {
	var params []any
	json.Unmarshal(r.Params, &params)
	return params
}

func (r *Request) ParamArray(a ...any) error {
	var params []json.RawMessage
	json.Unmarshal(r.Params, &params)
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
	return json.Unmarshal(r.Params, &v)
}

func (r *Request) Context() context.Context {
	return r.ctx
}

func (r *Request) Msg() jsonrpcMessage {
	return jsonrpcMessage{
		ID:     r.ID,
		Method: r.Method,
		Params: r.Params,
	}
}

func (r *Request) Remote() string {
	return r.Peer.RemoteAddr
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

var jpool = jsoniter.NewIterator(jsoniter.ConfigCompatibleWithStandardLibrary).Pool()

func (r *Request) Iter(fn func(j *jsoniter.Iterator) error) error {
	it := jpool.BorrowIterator(r.Params)
	defer jpool.ReturnIterator(it)
	return fn(it)
}
