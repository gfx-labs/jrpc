package jrpc

import (
	"context"
	codec2 "gfx.cafe/open/jrpc/pkg/codec"

	json "github.com/goccy/go-json"
	jsoniter "github.com/json-iterator/go"
)

var jpool = jsoniter.NewIterator(jsoniter.ConfigCompatibleWithStandardLibrary).Pool()

type Request struct {
	RequestMarshaling

	ctx context.Context
}

func (r *Request) UnmarshalJSON(xs []byte) error {
	return json.Unmarshal(xs, &r.RequestMarshaling)
}

func (r *Request) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.RequestMarshaling)
}

type RequestMarshaling struct {
	Version codec2.Version  `json:"jsonrpc"`
	ID      *codec2.ID      `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	Peer    codec2.PeerInfo `json:"-"`
}

func NewRequestInt(ctx context.Context, id int, method string, params any) *Request {
	r := &Request{ctx: ctx}
	pms, _ := json.Marshal(params)
	r.ID = codec2.NewNumberIDPtr(int64(id))
	r.Method = method
	r.Params = pms
	return r
}

func NewRequest(ctx context.Context, id string, method string, params any) *Request {
	r := &Request{ctx: ctx}
	pms, _ := json.Marshal(params)
	r.ID = codec2.NewStringIDPtr(id)
	r.Method = method
	r.Params = pms
	return r
}

func (r *Request) makeError(err error) *codec2.Message {
	m := r.Msg()
	return m.ErrorResponse(err)
}

func (r *Request) isNotification() bool {
	return r.ID == nil && len(r.Method) > 0
}

func (r *Request) isCall() bool {
	return r.hasValidID() && len(r.Method) > 0
}

func (r *Request) hasValidID() bool {
	return r.ID != nil && !r.ID.IsNull()
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

func (r *Request) Msg() codec2.Message {
	return codec2.Message{
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

func (r *Request) Iter(fn func(j *jsoniter.Iterator) error) error {
	it := jpool.BorrowIterator(r.Params)
	defer jpool.ReturnIterator(it)
	return fn(it)
}
