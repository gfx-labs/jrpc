package jsonrpc

import (
	"context"
	"encoding/json"

	"github.com/go-faster/jx"
)

// http.ResponseWriter interface, but for jrpc
type ResponseWriter interface {
	Send(v any, err error) error
	Notify(method string, v any) error
	ExtraFields() ExtraFields
}

type Request struct {
	ID         *ID                        `json:"id,omitempty"`
	Method     string                     `json:"method,omitempty"`
	Params      json.RawMessage            `json:"params,omitempty"`
	Peer        PeerInfo                   `json:"-"`
	ExtraFields map[string]json.RawMessage `json:"-"`

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
	var raw json.RawMessage
	if params != nil {
		raw, err = json.Marshal(params)
		if err != nil {
			return nil, err
		}
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
	r2 := new(Request)
	*r2 = *r
	r2.ctx = ctx
	r2.ID = r.ID
	r2.Method = r.Method
	r2.Params = r.Params
	r2.Peer = r.Peer
	return r2
}

func (r Request) MarshalJSON() ([]byte, error) {
	enc := jx.GetEncoder()
	enc.Obj(func(e *jx.Encoder) {
		e.FieldStart("jsonrpc")
		e.Str(VersionString)
		if r.ID != nil {
			e.FieldStart("id")
			e.Raw(*r.ID)
		}
		if r.Method != "" {
			e.FieldStart("method")
			e.Str(r.Method)
		}
		if r.Params != nil {
			e.FieldStart("params")
			e.Raw(r.Params)
		}
		if r.ExtraFields != nil {
			for k, v := range r.ExtraFields {
				e.FieldStart(k)
				e.Raw(v)
			}
		}
	})
	return enc.Bytes(), nil
}
