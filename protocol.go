package jrpc

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	jsoniter "github.com/json-iterator/go"
)

type HandlerFunc func(w ResponseWriter, r *Request)

type Handler interface {
	ServeRPC(w ResponseWriter, r *Request)
}

type ResponseWriter interface {
	Send(v any, err error) error
	Option(k string, v any)
	Header() http.Header

	Notify(v any) error
}

func (fn HandlerFunc) ServeRPC(w ResponseWriter, r *Request) {
	(fn)(w, r)
}

type Request struct {
	ctx context.Context
	msg jsonrpcMessage

	peer PeerInfo
}

func NewRequest(ctx context.Context, id string, method string, params any) *Request {
	r := &Request{ctx: ctx}
	pms, _ := jzon.Marshal(params)
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
	jzon.Unmarshal(r.msg.Params, &params)
	return params
}

var jpool = jsoniter.NewIterator(jsoniter.ConfigCompatibleWithStandardLibrary).Pool()

func (r *Request) Iter(fn func(j *jsoniter.Iterator) error) error {
	it := jpool.BorrowIterator(r.Params())
	defer jpool.ReturnIterator(it)
	return fn(it)
}

func (r *Request) ParamArray(a ...any) error {
	var params []json.RawMessage
	jzon.Unmarshal(r.msg.Params, &params)
	for idx, v := range params {
		if len(v) > idx {
			err := jzon.Unmarshal(v, &a[idx])
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
	return jzon.Unmarshal(r.msg.Params, &v)
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

type ResponseWriterMsg struct {
	r *Request
	n *Notifier
	s *Subscription

	msg *jsonrpcMessage

	options options
}

type options struct {
	sorted bool
}

func UpgradeToSubscription(w ResponseWriter, r *Request) (*Subscription, error) {
	not, ok := NotifierFromContext(r.ctx)
	if !ok || not == nil {
		return nil, errors.New("subscription not supported")
	}
	return not.CreateSubscription(), nil
}

func NewReaderResponseWriterMsg(r *Request) *ResponseWriterMsg {
	rw := &ResponseWriterMsg{
		r: r,
	}
	rw.n, _ = NotifierFromContext(r.ctx)
	return rw
}

func (w *ResponseWriterMsg) Header() http.Header {
	wh := w.r.Peer().HTTP.WriteHeaders
	return wh
}

func (w *ResponseWriterMsg) Option(k string, v any) {
	switch k {
	case "sorted":
		w.options.sorted = true
	case "unsorted":
		w.options.sorted = false
	}
}

func (w *ResponseWriterMsg) Send(args any, e error) (err error) {
	cm := w.r.Msg()
	if e != nil {
		w.msg = cm.errorResponse(e)
		return nil
	}
	switch c := args.(type) {
	case *Subscription:
		w.s = c
	default:
	}
	w.msg = cm.response(args)
	w.msg.sortKeys = w.options.sorted
	return nil
}

func (w *ResponseWriterMsg) Notify(args any) (err error) {
	if w.s == nil || w.n == nil {
		return ErrSubscriptionNotFound
	}
	bts, _ := jzon.Marshal(args)
	err = w.n.send(w.s, bts)
	if err != nil {
		return err
	}
	return nil
}

func (w *ResponseWriterMsg) Result() *jsonrpcMessage {
	return w.msg
}
