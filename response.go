package jrpc

import (
	"encoding/json"
	codec2 "gfx.cafe/open/jrpc/pkg/codec"
)

type Response struct {
	Version codec2.Version    `json:"jsonrpc,omitempty"`
	ID      *codec2.ID        `json:"id,omitempty"`
	Result  json.RawMessage   `json:"result,omitempty"`
	Error   *codec2.JsonError `json:"error,omitempty"`
}

func (r *Response) Msg() *codec2.Message {
	out := &codec2.Message{}
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
