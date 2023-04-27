package jrpc

import (
	"context"
	"strings"

	json "github.com/goccy/go-json"
	jsoniter "github.com/json-iterator/go"
)

type Request struct {
	Version version         `json:"jsonrpc"`
	ID      *ID             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`

	Peer PeerInfo `json:"-"`

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

func (r *Request) makeError(err error) *jsonrpcMessage {
	m := r.Msg()
	return m.errorResponse(err)
}

// DEPRECATED
// TODO: use our router to do this? jrpc.Namespace(string) (string, string) maybe?
func (r *Request) namespace() string {
	elem := strings.SplitN(r.Method, serviceMethodSeparator, 2)
	return elem[0]
}
func (r *Request) errorResponse(err error) *Response {
	mw := NewReaderResponseWriterMsg(r)
	mw.Send(nil, err)
	return mw.Response()
}

func (r *Request) isSubscribe() bool {
	return strings.HasSuffix(r.Method, subscribeMethodSuffix)
}
func (r *Request) isUnsubscribe() bool {
	return strings.HasSuffix(r.Method, unsubscribeMethodSuffix)
}
func (r *Request) isNotification() bool {
	return r.ID == nil && len(r.Method) > 0
}
func (r *Request) isCall() bool {
	return r.hasValidID() && len(r.Method) > 0
}
func (r *Request) isResponse() bool {
	return false
}
func (r *Request) hasValidID() bool {
	return r.ID != nil && !r.ID.null
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
