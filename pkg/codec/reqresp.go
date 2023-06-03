package codec

import (
	"context"

	json "github.com/goccy/go-json"
)

type Response struct {
	Version Version         `json:"jsonrpc,omitempty"`
	ID      *ID             `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JsonError      `json:"error,omitempty"`
}

func (r *Response) Msg() *Message {
	out := &Message{}
	if r.ID != nil {
		out.ID = r.ID
	}
	if r.Error != nil {
		out.Error = r.Error
	} else {
		out.Result = r.Result
	}
	return out
}

type Request struct {
	RequestMarshaling
	ctx context.Context
}

func NewRequestFromRaw(ctx context.Context, req *RequestMarshaling) *Request {
	return &Request{ctx: ctx, RequestMarshaling: *req}
}

func (r *Request) UnmarshalJSON(xs []byte) error {
	return json.Unmarshal(xs, &r.RequestMarshaling)
}

func (r *Request) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.RequestMarshaling)
}

type RequestMarshaling struct {
	Version Version         `json:"jsonrpc"`
	ID      *ID             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	Peer    PeerInfo        `json:"-"`
}

func NewRequestInt(ctx context.Context, id int, method string, params any) *Request {
	if ctx == nil {
		ctx = context.Background()
	}
	r := &Request{ctx: ctx}
	pms, _ := json.Marshal(params)
	r.ID = NewNumberIDPtr(int64(id))
	r.Method = method
	r.Params = pms
	return r
}

func NewRequest(ctx context.Context, id string, method string, params any) *Request {
	if ctx == nil {
		ctx = context.Background()
	}
	r := &Request{ctx: ctx}
	pms, _ := json.Marshal(params)
	r.ID = NewStringIDPtr(id)
	r.Method = method
	r.Params = pms
	return r
}

func NewNotification(ctx context.Context, method string, params any) *Request {
	if ctx == nil {
		ctx = context.Background()
	}
	r := &Request{ctx: ctx}
	pms, _ := json.Marshal(params)
	r.ID = nil
	r.Method = method
	r.Params = pms
	return r
}

func (r *Request) makeError(err error) *Message {
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

func (r *Request) Context() context.Context {
	return r.ctx
}

func (r *Request) Msg() Message {
	return Message{
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
