package jrpc

import (
	"encoding/json"
	"errors"
	"net/http"
)

type ResponseWriterMsg struct {
	r *Request
	n *Notifier
	s *Subscription

	msg *jsonrpcMessage

	//TODO: add options
	// currently there are no useful options so i havent added any
	// find a use case to add
	options options
}

type options struct {
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
	return nil
}

func (w *ResponseWriterMsg) Notify(args any) (err error) {
	if w.s == nil || w.n == nil {
		return ErrSubscriptionNotFound
	}
	bts, _ := json.Marshal(args)
	err = w.n.send(w.s, bts)
	if err != nil {
		return err
	}
	return nil
}

func (w *ResponseWriterMsg) Result() *jsonrpcMessage {
	return w.msg
}
