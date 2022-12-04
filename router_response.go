package jrpc

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Response struct {
	ID      *ID             `json:"id,omitempty"`
	Version version         `json:"jsonrpc,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonError      `json:"error,omitempty"`
}

func (r *Response) Msg() *jsonrpcMessage {
	out := &jsonrpcMessage{
		ID: r.ID,
	}
	if r.Error != nil {
		out.Error = r.Error
		return out
	}
	out.Result = r.Result
	return out
}

type ResponseWriterMsg struct {
	r    *Request
	resp *Response
	n    *Notifier
	s    *Subscription

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
		resp: &Response{
			ID: r.ID,
		},
	}
	rw.n, _ = NotifierFromContext(r.ctx)
	return rw
}

func (w *ResponseWriterMsg) Header() http.Header {
	wh := w.r.Peer.HTTP.WriteHeaders
	if wh == nil {
		wh = http.Header{}
	}
	return wh
}

func (w *ResponseWriterMsg) Option(k string, v any) {
}

func (w *ResponseWriterMsg) Send(args any, e error) (err error) {
	if e != nil {
		if c, ok := e.(*jsonError); ok {
			w.resp.Error = c
		} else {
			w.resp.Error = &jsonError{
				Code:    applicationErrorCode,
				Message: e.Error(),
			}
		}
		ec, ok := e.(Error)
		if ok {
			w.resp.Error.Code = ec.ErrorCode()
		}
		de, ok := e.(DataError)
		if ok {
			w.resp.Error.Data = de.ErrorData()
		}
		return nil
	}
	switch c := args.(type) {
	case *Subscription:
		w.s = c
	default:
	}
	w.resp.Result, err = jzon.Marshal(args)
	if err != nil {
		w.resp.Error = &jsonError{
			Code:    -32603,
			Message: err.Error(),
		}
		return nil
	}
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

func (w *ResponseWriterMsg) Response() *Response {
	return w.resp
}

func (w *ResponseWriterMsg) Msg() *jsonrpcMessage {
	return w.resp.Msg()
}
