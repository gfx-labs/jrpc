package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
)

// http.ResponseWriter interface, but for jrpc
type ResponseWriter interface {
	Send(v any, err error) error
	Notify(method string, v any) error
	ExtraFields() ExtraFields
}

type Request struct {
	ID          *ID                        `json:"id,omitempty"`
	Method      string                     `json:"method,omitempty"`
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
	buf := bytes.NewBuffer(make([]byte, 0, 64+len(r.Method)+len(r.Params)))
	buf.WriteString(`{"jsonrpc":"2.0"`)
	if r.ID != nil {
		buf.WriteString(`,"id":`)
		buf.Write(r.ID.RawMessage())
	}
	if r.Method != "" {
		buf.WriteString(`,"method":`)
		appendQuote(buf, r.Method)
	}
	if r.Params != nil {
		buf.WriteString(`,"params":`)
		buf.Write(r.Params)
	}
	for k, v := range r.ExtraFields {
		buf.WriteByte(',')
		appendQuote(buf, k)
		buf.WriteByte(':')
		buf.Write(v)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}
