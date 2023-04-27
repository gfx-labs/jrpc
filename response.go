package jrpc

import (
	"encoding/json"
	"net/http"

	"gfx.cafe/open/jrpc/codec"
)

type Response struct {
	Version codec.Version    `json:"jsonrpc,omitempty"`
	ID      *codec.ID        `json:"id,omitempty"`
	Result  json.RawMessage  `json:"result,omitempty"`
	Error   *codec.JsonError `json:"error,omitempty"`
}

func (r *Response) Msg() *codec.Message {
	out := &codec.Message{}
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

type ResponseWriterMsg struct {
	r    *Request
	resp *Response
}

type options struct {
}

func NewReaderResponseWriterMsg(r *Request) *ResponseWriterMsg {
	rw := &ResponseWriterMsg{
		r: r,
		resp: &Response{
			ID: r.ID,
		},
	}
	return rw
}

func (w *ResponseWriterMsg) Header() http.Header {
	wh := w.r.Peer.HTTP.Headers
	if wh == nil {
		wh = http.Header{}
	}
	return wh
}

func (w *ResponseWriterMsg) Option(k string, v any) {
}

func (w *ResponseWriterMsg) Send(args any, e error) (err error) {
	if e != nil {
		if c, ok := e.(*codec.JsonError); ok {
			w.resp.Error = c
		} else {
			w.resp.Error = &codec.JsonError{
				Code:    codec.ErrorCodeApplication,
				Message: e.Error(),
			}
		}
		ec, ok := e.(codec.Error)
		if ok {
			w.resp.Error.Code = ec.ErrorCode()
		}
		de, ok := e.(codec.DataError)
		if ok {
			w.resp.Error.Data = de.ErrorData()
		}
		return nil
	}
	w.resp.Result, err = json.Marshal(args)
	if err != nil {
		w.resp.Error = &codec.JsonError{
			Code:    -32603,
			Message: err.Error(),
		}
		return nil
	}
	return nil
}

// TODO: implement
func (w *ResponseWriterMsg) Notify(args any) (err error) {
	return nil
}

func (w *ResponseWriterMsg) Response() *Response {
	return w.resp
}

func (w *ResponseWriterMsg) Msg() *codec.Message {
	return w.resp.Msg()
}
