package codec

import (
	"context"
	"net/http"

	json "github.com/goccy/go-json"
)

// http.ResponseWriter interface, but for jrpc
type ResponseWriter interface {
	Send(v any, err error) error
	Header() http.Header

	SetExtraField(k string, v any) error

	Notify(method string, v any) error
}

// BatchElem is an element in a batch request.
type BatchElem struct {
	Method string
	Params any

	IsNotification bool

	// The result is unmarshaled into this field. Result must be set to a
	// non-nil pointer value of the desired type, otherwise the response will be
	// discarded.
	Result any
	// Error is set if the server returns an error for this request, or if
	// unmarshaling into Result fails. It is not set for I/O errors.
	Error error
}

type Request struct {
	ctx  context.Context
	Peer PeerInfo `json:"-"`

	Message
}

func (r *Request) UnmarshalJSON(xs []byte) error {
	return json.Unmarshal(xs, &r.Message)
}

func (r *Request) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Message)
}

func NewRequestFromMessage(ctx context.Context, message *Message) (r *Request) {
	if ctx == nil {
		ctx = context.Background()
	}
	r = &Request{ctx: ctx, Message: *message}
	return r
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
	err := json.Unmarshal(r.Params, &params)
	if err != nil {
		return err
	}
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
	r2.Error = r.Error
	r2.ExtraFields = r.ExtraFields
	r2.Peer = r.Peer
	return r2
}
