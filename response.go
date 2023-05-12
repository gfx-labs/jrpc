package jrpc

import (
	"encoding/json"

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
